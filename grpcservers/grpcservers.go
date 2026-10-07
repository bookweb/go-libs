package grpcservers

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"runtime/debug"
	"slices"

	" buf.build/go/protovalidate"
	"github.com/bookweb/go-libs/contexts"
	"github.com/bookweb/go-libs/logs"
	"github.com/bookweb/go-libs/telemetries"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_protovalidate "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	grpc_ratelimit "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_selector "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Config struct {
	AppName string
	Version string

	RateLimit int64

	UseTls    bool
	TlsConfig *tls.Config

	MsAuthAdapter      grpc_middlewares_auth.MsAuthAdapter
	OfficerAuthAdapter grpc_middlewares_auth.OfficerAuthAdapter
	UserAuthAdapter    grpc_middlewares_auth.UserAuthAdapter
	AuthExcludePaths   []string

	PromRegistry *prometheus.Registry

	Logger                 *slog.Logger
	PanicLogger            *slog.Logger
	LoggingSensitiveFields map[string]string
}

type Option func(cfg *Config)

func WithRateLimit(rateLimit int64) Option {
	return func(cfg *Config) {
		if rateLimit > 0 {
			cfg.RateLimit = rateLimit
		}
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(cfg *Config) {
		if logger != nil {
			cfg.Logger = logger
		}
	}
}

func WithPanicLogger(panicLogger *slog.Logger) Option {
	return func(cfg *Config) {
		if panicLogger != nil {
			cfg.PanicLogger = panicLogger
		}
	}
}

func WithLoggingSensitiveFields(loggingSensitiveFields map[string]string) Option {
	return func(cfg *Config) {
		if loggingSensitiveFields != nil {
			cfg.LoggingSensitiveFields = loggingSensitiveFields
		}
	}
}

func WithTlsConfig(tlsConfig *tls.Config) Option {
	return func(cfg *Config) {
		if tlsConfig != nil {
			cfg.TlsConfig = tlsConfig
		}
	}
}

func WithMsAuthAdapter(authAdapter grpc_middlewares_auth.MsAuthAdapter) Option {
	return func(cfg *Config) {
		if authAdapter != nil {
			cfg.MsAuthAdapter = authAdapter
		}
	}
}

func WithUserAuthAdapter(authAdapter grpc_middlewares_auth.UserAuthAdapter) Option {
	return func(cfg *Config) {
		if authAdapter != nil {
			cfg.UserAuthAdapter = authAdapter
		}
	}
}

func WithOfficerAuthAdapter(authAdapter grpc_middlewares_auth.OfficerAuthAdapter) Option {
	return func(cfg *Config) {
		if authAdapter != nil {
			cfg.OfficerAuthAdapter = authAdapter
		}
	}
}

func WithAuthExcludePaths(authExcludePaths []string) Option {
	return func(cfg *Config) {
		if authExcludePaths != nil {
			cfg.AuthExcludePaths = authExcludePaths
		}
	}
}

func WithPrometheusRegistry(promRegistry *prometheus.Registry) Option {
	return func(cfg *Config) {
		if promRegistry != nil {
			cfg.PromRegistry = promRegistry
		}
	}
}

type CreateGRPCServerFunc func()

func CreateGRPCServer(opts ...Option) (*grpc.Server, error) {
	return CreateGRPCServerFunc(func() {}).Create(opts...)
}

func (f CreateGRPCServerFunc) Create(opts ...Option) (*grpc.Server, error) {
	cfg := &Config{
		RateLimit:   0,
		Logger:      logs.NewSimpleJSONLogger(),
		PanicLogger: logs.NewSimplePanicLogger(),
	}

	for _, opt := range opts {
		opt(cfg)
	}
	// listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opts.GrpcAddress))
	// if err != nil {
	// 	loghelper.Logger.Panic("Server: Can not create listner - err: ", err)
	// }

	// auth
	authSelectorMatch := func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		if callMeta.Method == "Liveness" || callMeta.Method == "Readiness" {
			return false
		}

		if slices.Contains(cfg.AuthExcludePaths, callMeta.Method) {
			return false
		}

		return true
	}

	// proto protoValidator
	protoValidator, err := protovalidate.New()
	if err != nil {
		cfg.Logger.Error("Failed to auto migration", logs.Error(err))
		return nil, err
	}

	if cfg.PromRegistry == nil {
		cfg.PromRegistry = prometheus.NewRegistry()
	}
	// prometheus metrics
	prometheusServerMetrics := grpc_prometheus.NewServerMetrics(
		grpc_prometheus.WithServerHandlingTimeHistogram(
			grpc_prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	cfg.PromRegistry.MustRegister(prometheusServerMetrics)

	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		prometheusLabels := prometheus.Labels{
			// "app": cfg.App,
		}
		// if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
		// 	return prometheus.Labels{"traceID": span.TraceID().String()}
		// }
		return prometheusLabels
	}

	// zap logging
	loggingSelectorMatch := func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		return callMeta.Method != "Liveness" && callMeta.Method != "Readiness"
	}
	fieldsFromContextFunc := func(ctx context.Context) grpc_logging.Fields {
		traceId := ctx.Value(contexts.ContextKey__TraceId)
		if traceId != nil {
			return grpc_logging.Fields{"traceId", traceId}
		}
		return nil
	}
	customCodeToLevel := func(code codes.Code) grpc_logging.Level {
		switch code {
		case codes.OK:
			// return grpc_logging.LevelDebug
			return grpc_logging.LevelInfo
		case codes.NotFound, codes.InvalidArgument:
			return grpc_logging.LevelWarn
		case codes.Internal, codes.Unavailable:
			return grpc_logging.LevelError
		default:
			return grpc_logging.LevelInfo
		}
	}
	logEvents := []grpc_logging.LoggableEvent{
		// grpc_logging.FinishCall,
	}
	if cfg.Logger.Enabled(context.Background(), slog.LevelInfo) {
		logEvents = append(logEvents, grpc_logging.PayloadReceived, grpc_logging.PayloadSent)
	}
	loggingOpts := []grpc_logging.Option{
		grpc_logging.WithFieldsFromContext(fieldsFromContextFunc),
		grpc_logging.WithLevels(customCodeToLevel),
		grpc_logging.WithLogOnEvents(logEvents...),
	}
	zapLoggerInterceptor := grpc_middlewares_logging.InterceptorSlogLogger(
		cfg.Logger,
		cfg.LoggingSensitiveFields,
	)

	// traces
	tp, err := telemetries.CreateOtelTracerProvider(cfg.AppName, cfg.Version, nil)
	if err != nil {
		cfg.Logger.Error("Failed to auto migration", logs.Error(err))
		return nil, err
	}

	// recovery
	panicsTotal := promauto.With(cfg.PromRegistry).NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})
	grpcPanicRecoveryHandler := func(ctx context.Context, p any) (err error) {
		panicsTotal.Inc()
		// slog.Error("msg", "recovered from panic", "panic", p, "stack", debug.Stack())
		cfg.PanicLogger.ErrorContext(ctx, "recovered from panic: "+fmt.Sprint(p)+", stack trace: "+string(debug.Stack()), logs.TraceIdFromContext(ctx))
		return status.Errorf(codes.Internal, "%s", p)
	}

	chainUnaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_middlewares.SetIDUnaryServerInterceptor(),
	}
	if cfg.RateLimit != 0 {
		rateLimiter := grpc_middlewares.NewJujuLimiter(cfg.RateLimit)
		chainUnaryInterceptors = append(chainUnaryInterceptors, grpc_ratelimit.UnaryServerInterceptor(rateLimiter))
	}
	if cfg.UserAuthAdapter != nil {
		chainUnaryInterceptors = append(chainUnaryInterceptors,
			grpc_selector.UnaryServerInterceptor(
				grpc_auth.UnaryServerInterceptor(verify_firebase_jwt.AuthFunc(cfg.UserAuthAdapter)),
				grpc_selector.MatchFunc(authSelectorMatch),
			),
		)
	} else if cfg.OfficerAuthAdapter != nil {
		chainUnaryInterceptors = append(chainUnaryInterceptors,
			grpc_selector.UnaryServerInterceptor(
				grpc_auth.UnaryServerInterceptor(verify_admin_jwt.AuthFunc(cfg.OfficerAuthAdapter)),
				grpc_selector.MatchFunc(authSelectorMatch),
			),
		)
	} else if cfg.MsAuthAdapter != nil {
		chainUnaryInterceptors = append(chainUnaryInterceptors,
			grpc_selector.UnaryServerInterceptor(
				grpc_auth.UnaryServerInterceptor(verify_ms_jwt.AuthFunc(cfg.MsAuthAdapter)),
				grpc_selector.MatchFunc(authSelectorMatch),
			),
		)
	}
	chainUnaryInterceptors = append(chainUnaryInterceptors,
		grpc_protovalidate.UnaryServerInterceptor(protoValidator),
		// otelgrpc.UnaryServerInterceptor(),
		prometheusServerMetrics.UnaryServerInterceptor(grpc_prometheus.WithExemplarFromContext(exemplarFromContext)),
		grpc_selector.UnaryServerInterceptor(
			grpc_logging.UnaryServerInterceptor(zapLoggerInterceptor, loggingOpts...),
			grpc_selector.MatchFunc(loggingSelectorMatch),
		),
		grpc_middlewares.SetTimeMsUnaryServerInterceptor(),
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(grpcPanicRecoveryHandler)),
	)

	// grpcmiddlewarehelper.VerifyFirebaseJwtUnaryServerInterceptor(opts.FirebaseAuthClientAdapter, opts.AuthExcludePaths)
	// middlewares chain
	grpcMiddlewareServerOption := grpc.ChainUnaryInterceptor(
		chainUnaryInterceptors...,
	)

	grpcOpts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(tp),
		)),
		grpcMiddlewareServerOption,
	}

	// tls config
	if cfg.UseTls && cfg.TlsConfig != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(
			credentials.NewTLS(cfg.TlsConfig),
		))
	}

	grpcServer := grpc.NewServer(grpcOpts...)

	reflection.Register(grpcServer)
	// pbpcbss.RegisterPCBSSAPIServer(grpcServer, api)
	prometheusServerMetrics.InitializeMetrics(grpcServer)

	// err = grpcServer.Serve(listener)
	// if err != nil {
	// 	loghelper.Logger.Panic("Server: Failed to serve - err: ", err)
	// 	return nil, err
	// }

	// grpcServer for listen
	// promRegistry for serve /metrics endpoint
	return grpcServer, nil
}
