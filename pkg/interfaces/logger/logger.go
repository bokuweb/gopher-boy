package logger

// Logger is
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
}
