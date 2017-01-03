package misc

import (
	"github.com/uber-go/zap"
	"runtime"
)

type Log interface {
	Location()
	Debug(...zap.Field)
	DebugErr(error) error
	Info(...zap.Field)
	InfoErr(error) error
	Warn(...zap.Field)
	WarnErr(error) error
	Error(...zap.Field)
	ErrorErr(error) error
	Panic(...zap.Field)
	PanicErr(error) error
	Fatal(...zap.Field)
	FatalErr(error) error
}

func NewLog(logger zap.Logger) Log {
	return &log{
		logger: logger,
	}
}

type log struct {
	logger zap.Logger
}

func (l *log) log(level zap.Level, fields ...zap.Field) {
	if l.logger != nil {
		pc, file, line, _ := runtime.Caller(2)
		f := runtime.FuncForPC(pc)
		extendedFields := make([]zap.Field, 0, len(fields)+3)
		extendedFields = append(extendedFields, zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line))
		extendedFields = append(extendedFields, fields...)
		l.logger.Log(level, "", extendedFields...)
	}
}

func (l *log) Location() {
	l.log(zap.DebugLevel)
}

func (l *log) Debug(fields ...zap.Field) {
	l.log(zap.DebugLevel, fields...)
}

func (l *log) DebugErr(err error) error {
	l.log(zap.DebugLevel, zap.Error(err))
	return err
}

func (l *log) Info(fields ...zap.Field) {
	l.log(zap.InfoLevel, fields...)
}

func (l *log) InfoErr(err error) error {
	l.log(zap.InfoLevel, zap.Error(err))
	return err
}

func (l *log) Warn(fields ...zap.Field) {
	l.log(zap.WarnLevel, fields...)
}

func (l *log) WarnErr(err error) error {
	l.log(zap.WarnLevel, zap.Error(err))
	return err
}

func (l *log) Error(fields ...zap.Field) {
	l.log(zap.ErrorLevel, fields...)
}

func (l *log) ErrorErr(err error) error {
	l.log(zap.ErrorLevel, zap.Error(err))
	return err
}

func (l *log) Panic(fields ...zap.Field) {
	l.log(zap.PanicLevel, fields...)
}

func (l *log) PanicErr(err error) error {
	l.log(zap.PanicLevel, zap.Error(err))
	return err
}

func (l *log) Fatal(fields ...zap.Field) {
	l.log(zap.FatalLevel, fields...)
}

func (l *log) FatalErr(err error) error {
	l.log(zap.FatalLevel, zap.Error(err))
	return err
}
