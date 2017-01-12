package misc

import (
	"errors"
	"fmt"
	"github.com/uber-go/zap"
	"runtime"
)

var (
	NotImplementedErr = errors.New("not implemented")
)

type ErrorRef struct {
	Id Id `json:"id"`
}

func (e *ErrorRef) Error() string {
	return fmt.Sprintf("errorRef: %s", e.Id.String())
}

type Log interface {
	Location()
	Debug(...zap.Field)
	DebugErr(error) error
	DebugUserErr(Id, error) error
	Info(...zap.Field)
	InfoErr(error) error
	InfoUserErr(Id, error) error
	Warn(...zap.Field)
	WarnErr(error) error
	WarnUserErr(Id, error) error
	Error(...zap.Field)
	ErrorErr(error) error
	ErrorUserErr(Id, error) error
	Panic(...zap.Field)
	PanicErr(error) error
	PanicUserErr(Id, error) error
	Fatal(...zap.Field)
	FatalErr(error) error
	FatalUserErr(Id, error) error
}

type zapLogger interface {
	Log(zap.Level, string, ...zap.Field)
}

func NewLog(logger zapLogger) Log {
	return newLog(logger, NewId)
}

func newLog(logger zapLogger, genNewId GenNewId) Log {
	return &log{
		logger:   logger,
		genNewId: genNewId,
	}
}

type log struct {
	logger   zapLogger
	genNewId GenNewId
}

func (l *log) log(callerDepth int, level zap.Level, fields ...zap.Field) {
	if l.logger != nil {
		pc, file, line, _ := runtime.Caller(callerDepth)
		f := runtime.FuncForPC(pc)
		extendedFields := make([]zap.Field, 0, len(fields)+3)
		extendedFields = append(extendedFields, zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line))
		extendedFields = append(extendedFields, fields...)
		l.logger.Log(level, "", extendedFields...)
	}
}

func (l *log) logErr(level zap.Level, err error) error {
	if level == zap.DebugLevel || level == zap.InfoLevel {
		l.log(3, level, zap.Error(err))
		return err
	} else {
		id, _ := l.genNewId()
		l.log(3, level, zap.String("errorRef", id.String()), zap.Error(err))
		return &ErrorRef{Id: id}
	}
}

func (l *log) logUserErr(level zap.Level, userId Id, err error) error {
	if level == zap.DebugLevel || level == zap.InfoLevel {
		l.log(3, level, zap.String("user", userId.String()), zap.Error(err))
		return err
	} else {
		id, _ := l.genNewId()
		l.log(3, level, zap.String("user", userId.String()), zap.String("errorRef", id.String()), zap.Error(err))
		return &ErrorRef{Id: id}
	}
}

func (l *log) Location() {
	l.log(2, zap.DebugLevel)
}

func (l *log) Debug(fields ...zap.Field) {
	l.log(2, zap.DebugLevel, fields...)
}

func (l *log) DebugErr(err error) error {
	return l.logErr(zap.DebugLevel, err)
}

func (l *log) DebugUserErr(userId Id, err error) error {
	return l.logUserErr(zap.DebugLevel, userId, err)
}

func (l *log) Info(fields ...zap.Field) {
	l.log(2, zap.InfoLevel, fields...)
}

func (l *log) InfoErr(err error) error {
	return l.logErr(zap.InfoLevel, err)
}

func (l *log) InfoUserErr(userId Id, err error) error {
	return l.logUserErr(zap.InfoLevel, userId, err)
}

func (l *log) Warn(fields ...zap.Field) {
	l.log(2, zap.WarnLevel, fields...)
}

func (l *log) WarnErr(err error) error {
	return l.logErr(zap.WarnLevel, err)
}

func (l *log) WarnUserErr(userId Id, err error) error {
	return l.logUserErr(zap.WarnLevel, userId, err)
}

func (l *log) Error(fields ...zap.Field) {
	l.log(2, zap.ErrorLevel, fields...)
}

func (l *log) ErrorErr(err error) error {
	return l.logErr(zap.ErrorLevel, err)
}

func (l *log) ErrorUserErr(userId Id, err error) error {
	return l.logUserErr(zap.ErrorLevel, userId, err)
}

func (l *log) Panic(fields ...zap.Field) {
	l.log(2, zap.PanicLevel, fields...)
}

func (l *log) PanicErr(err error) error {
	return l.logErr(zap.PanicLevel, err)
}

func (l *log) PanicUserErr(userId Id, err error) error {
	return l.logUserErr(zap.PanicLevel, userId, err)
}

func (l *log) Fatal(fields ...zap.Field) {
	l.log(2, zap.FatalLevel, fields...)
}

func (l *log) FatalErr(err error) error {
	return l.logErr(zap.FatalLevel, err)
}

func (l *log) FatalUserErr(userId Id, err error) error {
	return l.logUserErr(zap.FatalLevel, userId, err)
}
