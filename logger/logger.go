package logger

import "go.uber.org/zap"

var (
	logger *zap.SugaredLogger
)

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = l.Sugar()
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}
