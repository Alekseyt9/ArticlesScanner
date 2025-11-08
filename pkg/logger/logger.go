package logger

import (
	"fmt"
	"log"
	"os"
)

// New returns a stdlib-backed logger with component prefix.
func New(component string) *log.Logger {
	prefix := fmt.Sprintf("[%s] ", component)
	return log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile)
}
