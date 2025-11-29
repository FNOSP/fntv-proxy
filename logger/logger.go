package logger

import (
	"log"
	"os"
)

var StdoutLogger *log.Logger

func Init() {
	StdoutLogger = log.New(os.Stdout, "", log.LstdFlags)
}
