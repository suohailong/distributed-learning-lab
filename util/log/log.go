package log

import (
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Debugf(format string, a ...interface{}) {
	logrus.Debugf(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	logrus.Fatalf(format, a...)
}
