package utils

import (
	"io"
	"log"

	"github.com/badico-cloud-hub/pubsub/interfaces"
)

//Logger is struct of logger default
type Logger struct {
	infoLog    *log.Logger
	errorLog   *log.Logger
	warningLog *log.Logger
}

//NewLogger return new logger
func NewLogger(output io.Writer) interfaces.ServiceLogger {
	return &Logger{
		infoLog:    log.New(output, "INFO: ", log.Ldate|log.Ltime),
		errorLog:   log.New(output, "ERROR: ", log.Ldate|log.Ltime),
		warningLog: log.New(output, "WARNING: ", log.Ldate|log.Ltime),
	}
}

//Info print info logs
func (l *Logger) Info(message string) {
	l.infoLog.Println(message)
}

//Warning print warning logs
func (l *Logger) Warning(message string) {
	l.warningLog.Println(message)
}

//Error print error logs
func (l *Logger) Error(message string) {
	l.errorLog.Println(message)
}
