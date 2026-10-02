package logs

import "log/slog"

type SimplePanicLogger func()

func NewSimplePanicLogger(opts ...Option) *slog.Logger {
	return SimplePanicLogger(func() {}).Create(opts...)
}

func (f SimplePanicLogger) Create(opts ...Option) *slog.Logger {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	handlerOptions := slog.HandlerOptions{
		Level: slog.LevelError,
	}

	if cfg.level.String() != "" {
		handlerOptions.Level = cfg.level
	}

	f()
	return slog.Default()
}
