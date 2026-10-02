package httpservers

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"gitlab.com/pmtrade/pm-go-libs/logs"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Config struct {
	HttpRouter *http.ServeMux

	// tls
	TlsConfig *tls.Config

	// cors
	CorsEnabled        bool
	CorsAllowedOrigins []string

	// prometheus
	PromRegistry *prometheus.Registry

	Logger      *slog.Logger
	PanicLogger *slog.Logger
}

type Option func(cfg *Config)

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

func WithTlsConfig(tlsConfig *tls.Config) Option {
	return func(cfg *Config) {
		if tlsConfig != nil {
			cfg.TlsConfig = tlsConfig
		}
	}
}

func WithCORsEnabled(corsEnabled bool) Option {
	return func(cfg *Config) {
		cfg.CorsEnabled = corsEnabled
	}
}

func WithCORsAllowedOrigins(allowedOrigins []string) Option {
	return func(cfg *Config) {
		if allowedOrigins != nil {
			cfg.CorsAllowedOrigins = allowedOrigins
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

// type (
// 	CreateGRPCServeMuxOptions struct {
// 		UseTls    bool
// 		TlsConfig *tls.Config
// 	}

// 	CreateGRPCDialOptions struct {
// 		UseTls    bool
// 		TlsConfig *tls.Config
// 	}
// )

type CreateHTTPServerFunc func()

func CreateHTTPServer(httpRouter *http.ServeMux, opts ...Option) (*http.Server, error) {
	return CreateHTTPServerFunc(func() {}).Create(httpRouter, opts...)
}

func (f CreateHTTPServerFunc) Create(httpRouter *http.ServeMux, opts ...Option) (*http.Server, error) {
	if httpRouter == nil {
		return nil, errors.New("")
	}

	cfg := &Config{
		Logger:      logs.NewSimpleJSONLogger(),
		PanicLogger: logs.NewSimplePanicLogger(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.PromRegistry != nil {
		httpRouter.Handle("/metrics", promhttp.HandlerFor(
			cfg.PromRegistry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
	}

	panicRecoveryMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					ctx := r.Context()
					msg := "recovered from panic: " + fmt.Sprint(err) + ", stack trace: " + string(debug.Stack())
					cfg.PanicLogger.ErrorContext(ctx, msg, logs.TraceIdFromContext(ctx))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}

	middlewares := []Middleware{
		panicRecoveryMiddleware,
	}

	if cfg.CorsEnabled {
		cfg.Logger.Info(fmt.Sprintf("CORS enabled: %v", cfg.CorsEnabled))
		corsHandler := cors.New(cors.Options{
			// AllowedOrigins: []string{
			// 	"http://localhost:4201",
			// 	"http://localhost:4202",
			// 	"http://localhost:4204",
			// },
			AllowedOrigins:   cfg.CorsAllowedOrigins,
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			ExposedHeaders: []string{
				"Grpc-Metadata-Authorization",
				"Content-Type",
				"Content-Disposition",
				"File-Name",
				"Content-Transfer-Encoding",
				"Status-Code",
				"Trace-Id",
			},
			Debug: false,
		})
		middlewares = append(middlewares, corsHandler.Handler)
	}

	httpMuxWithMiddlewares := MiddlewaresChain(httpRouter, middlewares...)

	var httpServer *http.Server
	if cfg.TlsConfig != nil {
		httpServer = &http.Server{
			Handler:   httpMuxWithMiddlewares,
			TLSConfig: cfg.TlsConfig,
		}
	} else {
		h2s := &http2.Server{}
		httpServer = &http.Server{
			Handler: h2c.NewHandler(httpMuxWithMiddlewares, h2s),
		}
	}

	return httpServer, nil
}

// func CreateGRPCServeMux(opts *CreateGRPCServeMuxOptions) (*runtime.ServeMux, []grpc.DialOption, error) {
// 	grpcMux := runtime.NewServeMux()
// 	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
// 	if opts.UseTls && opts.TlsConfig != nil {
// 		tlsCredentials := credentials.NewTLS(opts.TlsConfig)
// 		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(tlsCredentials)}
// 	}

// 	return grpcMux, dialOpts, nil
// }

// func BuildGRPCDialOptions(opts *CreateGRPCDialOptions) ([]grpc.DialOption, error) {
// 	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
// 	if opts.UseTls && opts.TlsConfig != nil {
// 		tlsCredentials := credentials.NewTLS(opts.TlsConfig)
// 		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(tlsCredentials)}
// 	}

// 	return dialOpts, nil
// }

// func CreateHTTPServerFromServeMux(httpMux *http.ServeMux, opts *Config) (*http.Server, error) {
// 	var logger *slog.Logger = opts.Logger
// 	if logger == nil {
// 		logger = logs.NewSimpleJSONLogger()
// 	}

// 	if httpMux == nil {
// 		logger.Error("Http Server: httpMux cannot null")
// 		return nil, errors.New("")
// 	}

// 	httpMux.Handle("/metrics", promhttp.HandlerFor(
// 		opts.PromRegistry,
// 		promhttp.HandlerOpts{
// 			EnableOpenMetrics: true,
// 		},
// 	))

// 	panicRecoveryMiddleware := func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			defer func() {
// 				if err := recover(); err != nil {
// 					logger.PanicContext(r.Context(), "recovered from panic: "+fmt.Sprint(err)+", stack trace: "+string(debug.Stack()))
// 					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 				}
// 			}()
// 			next.ServeHTTP(w, r)
// 		})
// 	}

// 	middlewares := []Middleware{
// 		panicRecoveryMiddleware,
// 	}

// 	if opts.CorsEnabled {
// 		logger.Info(fmt.Sprintf("CORS enabled: %v", opts.CorsEnabled))
// 		corsHandler := cors.New(cors.Options{
// 			// AllowedOrigins: []string{
// 			// 	"http://localhost:4201",
// 			// 	"http://localhost:4202",
// 			// 	"http://localhost:4204",
// 			// },
// 			AllowedOrigins:   opts.CorsAllowedOrigins,
// 			AllowCredentials: true,
// 			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
// 			AllowedHeaders:   []string{"*"},
// 			ExposedHeaders: []string{
// 				"Grpc-Metadata-Authorization",
// 				"Content-Type",
// 				"Content-Disposition",
// 				"File-Name",
// 				"Content-Transfer-Encoding",
// 				"Status-Code",
// 				"Trace-Id",
// 			},
// 			Debug: false,
// 		})
// 		middlewares = append(middlewares, corsHandler.Handler)
// 	}

// 	httpMuxWithMiddlewares := MiddlewaresChain(httpMux, middlewares...)

// 	var httpServer *http.Server
// 	if opts.UseTls && opts.TlsConfig != nil {
// 		httpServer = &http.Server{
// 			Handler:   httpMuxWithMiddlewares,
// 			TLSConfig: opts.TlsConfig,
// 		}
// 	} else {
// 		h2s := &http2.Server{}
// 		httpServer = &http.Server{
// 			Handler: h2c.NewHandler(httpMuxWithMiddlewares, h2s),
// 		}
// 	}

// 	return httpServer, nil
// }
