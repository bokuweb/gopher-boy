package logger

import (
	"log"
)

// LogLevel is kind of Loglavel
type LogLevel string

const (
	// LogDebug is
	LogDebug LogLevel = "Debug"
	// LogInfo is
	LogInfo LogLevel = "Info"
	// LogSilent is
	LogSilent LogLevel = "Silent"
)

// Log is
type Log struct {
	Level LogLevel
}

// NewLogger is logger constructor
func NewLogger(level LogLevel) *Log {
	// now := time.Now().Unix()
	// n := strconv.FormatInt(now, 10)
	// logfile, err := os.OpenFile("./logs/test-"+n+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	panic("cannnot open test.log:" + err.Error())
	// }
	// log.SetOutput(io.MultiWriter(logfile))
	// log.SetFlags(log.Ldate | log.Ltime)

	return &Log{
		Level: level,
	}
}

// Debug is
func (l *Log) Debug(args ...interface{}) {
	if l.Level != "Debug" {
		return
	}
	log.Println("[DEBUG] ", args)
}

// Info is
func (l *Log) Info(args ...interface{}) {
	log.Println("[Info] ", args)
}

// Error is
func (l *Log) Error(args ...interface{}) {
	log.Println("[ERROR] ", args)
}

// Warn is
func (l *Log) Warn(args ...interface{}) {
	log.Println("[WARNING] ", args)
}
