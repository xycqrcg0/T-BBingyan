package log

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Logger *logrus.Logger

func init() {
	w1 := os.Stdout
	filePath, ok := os.LookupEnv("LOG_PATH")
	if !ok {
		filePath = "./logs/log.log"
	}
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		err := os.Mkdir("./logs", 0755)
		if err != nil {
			panic(err)
		}
	}
	w2, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	Logger = logrus.New()
	Logger.SetOutput(io.MultiWriter(w1, w2))
	Logger.SetFormatter(&logrus.JSONFormatter{})
	//Logger.SetReportCaller(true)

	Logger.Infof("Finish Initializing log")
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}
