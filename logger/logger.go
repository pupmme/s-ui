package logger

import (
	"log"
	"os"
)

func InitLogger() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Debug(args ...interface{}) { log.Println(args...) }
func Info(args ...interface{})   { log.Println(args...) }
func Warning(args ...interface{}) { log.Println("[WARN]", args) }
func Error(args ...interface{})  { log.Println("[ERROR]", args) }
func Fatal(args ...interface{})  { log.Println("[FATAL]", args); os.Exit(1) }
func Infof(format string, args ...interface{}) { log.Printf(format, args...) }
func Errorf(format string, args ...interface{}) { log.Printf("[ERROR] "+format, args...) }
func Debugf(format string, args ...interface{}) { log.Printf("[DEBUG] "+format, args...) }

var logBuffer []string

func GetLogs(c int, level string) []string {
	n := c
	if n > len(logBuffer) {
		n = len(logBuffer)
	}
	return logBuffer[len(logBuffer)-n:]
}

func addToBuffer(msg string) {
	logBuffer = append(logBuffer, msg)
	if len(logBuffer) > 1000 {
		logBuffer = logBuffer[1:]
	}
}
