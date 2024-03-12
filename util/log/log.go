package log

import "github.com/sirupsen/logrus"

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Debugf(format string, a ...interface{}) {
	logrus.Debugf(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	logrus.Fatalf(format, a...)
}
