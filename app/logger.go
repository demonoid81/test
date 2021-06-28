package app

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"
	"github.com/sphera-erp/sphera/pkg/tarantool"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	_, file, line, _ := runtime.Caller(3)
	e.Str("runtime", fmt.Sprintf("%s:%d", filepath.Base(file), line))
}

// New is
func (app *App) NewLogger() *zerolog.Logger {
	logLevel := zerolog.InfoLevel
	if app.Cfg.Logs.Level != "" {
		logLevel = getLevel(app.Cfg.Logs.Level)
	}
	zerolog.SetGlobalLevel(logLevel)

	logger1 := zerolog.New(os.Stderr).With().Timestamp().Logger()

	logger := logger1.Hook(SeverityHook{})

	return &logger
}

// NewConsole is
func (app *App) NewConsoleLogger() *zerolog.Logger {
	// по умолчанию только информирование
	logLevel := zerolog.InfoLevel
	if app.Cfg.Logs.Level != "" {
		logLevel = getLevel(app.Cfg.Logs.Level)
	}
	zerolog.SetGlobalLevel(logLevel)

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return &logger
}

func getLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "undef":
		return zerolog.NoLevel
	}
	return zerolog.Disabled
}

// Output duplicates the global logger and sets w as its output.
func (app *App) Output(w io.Writer) zerolog.Logger {
	return app.Logger.Output(w)
}

func (app *App) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	var zlevel zerolog.Level
	switch level {
	case pgx.LogLevelNone:
		zlevel = zerolog.NoLevel
	case pgx.LogLevelError:
		zlevel = zerolog.ErrorLevel
	case pgx.LogLevelWarn:
		zlevel = zerolog.WarnLevel
	case pgx.LogLevelInfo:
		zlevel = zerolog.InfoLevel
	case pgx.LogLevelDebug:
		zlevel = zerolog.DebugLevel
	default:
		zlevel = zerolog.DebugLevel
	}

	pgxLog := app.Logger.With().Fields(data).Logger()
	pgxLog.WithLevel(zlevel).Msg(msg)
}

func (app *App) Report(event tarantool.ConnLogKind, conn *tarantool.Connection, v ...interface{}) {
	tarantoolLog := app.Logger.With().Logger()
	switch event {
	case tarantool.LogReconnectFailed:
		reconnects := v[0].(uint)
		err := v[1].(error)
		// MaxReconnects
		tarantoolLog.Info().Msgf("tarantool: reconnect (%d/%d) to %s failed: %s\n", reconnects, conn.MaxReconnects(), conn.Addr(), err.Error())
	case tarantool.LogLastReconnectFailed:
		err := v[0].(error)
		tarantoolLog.Info().Msgf("tarantool: last reconnect to %s failed: %s, giving it up.\n", conn.Addr(), err.Error())
	case tarantool.LogUnexpectedResultId:
		resp := v[0].(*tarantool.Response)
		tarantoolLog.Info().Msgf("tarantool: connection %s got unexpected resultId (%d) in response", conn.Addr(), resp.RequestId)
	default:
		args := ""
		for i, val := range v {
			if i > 0 {
				args += ", "
			}
			args += val.(string)
		}
		tarantoolLog.Info().Msgf("tarantool: connection %s got unexpecting event %s: args: %s ", conn.Addr(), event, args)
	}
}
