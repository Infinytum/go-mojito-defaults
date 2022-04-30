package defaults

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/infinytum/go-mojito"
	"github.com/infinytum/go-mojito/util"
	"github.com/rs/zerolog"
)

type zerologLogger struct {
	fields util.LogFields
	logger zerolog.Logger
}

// Debug will write a debug log
func (z *zerologLogger) Debug(msg interface{}) {
	z.log(msg, z.logger.Debug())
}

// Debugf will write a debug log sprintf-style
func (z *zerologLogger) Debugf(msg string, values ...interface{}) {
	msg = fmt.Sprintf(msg, values...)
	z.Debug(msg)
}

// Error will write a error log
func (z *zerologLogger) Error(msg interface{}) {
	z.log(msg, z.logger.Error())
}

// Errorf will write a error log sprintf-style
func (z *zerologLogger) Errorf(msg string, values ...interface{}) {
	msg = fmt.Sprintf(msg, values...)
	z.Error(msg)
}

// Fatal will write a fatal log
func (z *zerologLogger) Fatal(msg interface{}) {
	z.log(msg, z.logger.Fatal())
}

// Fatalf will write a fatal log sprintf-style
func (z *zerologLogger) Fatalf(msg string, values ...interface{}) {
	msg = fmt.Sprintf(msg, values...)
	z.Fatal(msg)
}

// Field will add a field to a new logger and return it
func (z *zerologLogger) Field(name string, val interface{}) mojito.Logger {
	return z.Fields(util.LogFields{name: val})
}

// Fields will add multiple fields to a new logger and return it
func (z *zerologLogger) Fields(fields util.LogFields) mojito.Logger {
	newLog := &zerologLogger{
		fields: z.fields.Clone(),
		logger: z.logger.With().Logger(),
	}
	for name, val := range fields {
		newLog.fields[name] = val
	}
	return newLog
}

// Info will write a info log
func (z *zerologLogger) Info(msg interface{}) {
	z.log(msg, z.logger.Info())
}

// Infof will write a info log sprintf-style
func (z *zerologLogger) Infof(msg string, values ...interface{}) {
	msg = fmt.Sprintf(msg, values...)
	z.Info(msg)
}

// Trace will write a trace log
func (z *zerologLogger) Trace(msg interface{}) {
	z.log(msg, z.logger.Trace())
}

// Tracef will write a trace log sprintf-style
func (z *zerologLogger) Tracef(msg string, values ...interface{}) {
	msg = fmt.Sprintf(msg, values...)
	z.Trace(msg)
}

// Warn will write a warn log
func (z *zerologLogger) Warn(msg interface{}) {
	z.log(msg, z.logger.Warn())
}

// Warnf will write a warn log sprintf-style
func (z *zerologLogger) Warnf(msg string, values ...interface{}) {
	msg = fmt.Sprintf(msg, values...)
	z.Warn(msg)
}

func (z *zerologLogger) log(msg interface{}, event *zerolog.Event) {
	for name, val := range z.fields {
		event = event.Interface(name, val)
	}
	event.Msg(fmt.Sprint(msg))
}

// newZerologLogger will create a new instance of the mojito zerolog implementation
func newZerologLogger() mojito.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("%s", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	log := zerolog.New(output).With().Timestamp().Logger()
	return &zerologLogger{
		fields: make(util.LogFields),
		logger: log,
	}
}
