package helper

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

type logLevel string

const (
	LogLevelPanic logLevel = "panic"
	LogLevelFatal logLevel = "fatal"
	LogLevelError logLevel = "error"
	LogLevelWarn  logLevel = "warn"
	LogLevelDebug logLevel = "debug"
	LogLevelTrace logLevel = "trace"
	LogLevelInfo  logLevel = "info"
)

func logInit() {
	log = logrus.New()

	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
}

func LogPrint(level logLevel, message interface{}) {
	if log != nil {
		switch level {
		case "panic":
			log.Panicf("%v\n", message)
		case "fatal":
			log.Fatalf("%v\n", message)
		case "error":
			log.Errorf("%v\n", message)
		case "warn":
			log.Warnf("%v\n", message)
		case "debug":
			log.Debugf("%v\n", message)
		case "trace":
			log.Tracef("%v\n", message)
		default:
			log.Infof("%v\n", message)
		}
	}
}
