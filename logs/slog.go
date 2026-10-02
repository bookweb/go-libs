package logs

import (
	"io"
	"log/slog"
	"os"
)

func GetDefaultSlogWriter() io.Writer {
	return os.Stdout
}

type Config struct {
	name string

	level slog.Level

	sensitiveFields map[string]string
}

type Option func(c *Config)

func WithName(name string) Option {
	return func(c *Config) {
		if name != "" {
			c.name = name
		}
	}
}

func WithLevel(level slog.Level) Option {
	return func(c *Config) {
		c.level = level
	}
}

func WithSensitiveFields(sensitiveFields map[string]string) Option {
	return func(c *Config) {
		c.sensitiveFields = sensitiveFields
	}
}
