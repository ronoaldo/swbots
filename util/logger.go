package util

import (
	"fmt"
	"log"
)

// Simple named logger
type Logger struct {
	Name string
}

func (l *Logger) Infof(msgfmt string, arguments ...interface{}) {

}

func (l *Logger) logf(msgfmt string, level string, arguments ...interface{}) {
	log.Printf(fmt.Sprintf("%s %s", level, msgfmt), arguments...)
}
