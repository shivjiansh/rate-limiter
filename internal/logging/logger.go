package logging

import "log"

type Logger struct{}

func New() *Logger { return &Logger{} }

func (l *Logger) Info(msg string) {
	log.Println(msg)
}
