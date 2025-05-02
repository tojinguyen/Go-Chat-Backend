package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	errorConstants "gochat-backend/error"

	"github.com/sirupsen/logrus"
)

type logger struct {
	Log *logrus.Entry
}

func InitLogger(logType LogType, processId string) *logger {
	log := logrus.WithFields(logrus.Fields{
		"logType":   logType,
		"processId": processId,
	})
	return &logger{
		Log: log,
	}
}

func Log(ctx context.Context,
	logLevel LogLevel,
	errorMessage string,
	errorCodeSrc *errorConstants.ErrorCode,
) *errorConstants.ErrorCode {
	counter, filename, _, _ := runtime.Caller(1)
	functionNameRaw := strings.Split(runtime.FuncForPC(counter).Name(), ".")
	functionName := functionNameRaw[len(functionNameRaw)-1]
	screenId := ctx.Value("screenId")
	apiOrder := ctx.Value("apiOrder")
	var errorCode *errorConstants.ErrorCode
	if errorCodeSrc != nil {
		if screenId == nil && apiOrder == nil && errorCodeSrc != nil {
			errorCode = &errorConstants.ErrorCode{
				HTTPCode:    errorCodeSrc.HTTPCode,
				Type:        errorCodeSrc.Type,
				Code:        errorConstants.Code(errorCodeSrc.Code),
				FieldErrors: errorCodeSrc.FieldErrors,
			}
		} else {
			errorCode = &errorConstants.ErrorCode{
				HTTPCode:    errorCodeSrc.HTTPCode,
				Type:        errorCodeSrc.Type,
				Code:        errorConstants.Code(fmt.Sprintf("%s-%s-%s", errorCodeSrc.Code, screenId, apiOrder)),
				FieldErrors: errorCodeSrc.FieldErrors,
			}
		}
	}

	loggerRaw := ctx.Value("logger")
	logger := loggerRaw.(*logrus.Entry)
	logger = logger.WithFields(logrus.Fields{
		"logType":      LogTypeHandler,
		"filename":     filename,
		"functionName": functionName,
	})
	if errorCode != nil {
		logger = logger.WithField("errorCode", errorCode.Code)
	}
	switch logLevel {
	case LogLevelDebug:
		logger.Debug(errorMessage)
	case LogLevelInfo:
		logger.Info(errorMessage)
	case LogLevelFatal:
		logger.Fatal(errorMessage)
	case LogLevelWarn:
		logger.Warn(errorMessage)
	case LogLevelError:
		logger.Error(errorMessage)
	default:
	}
	return errorCode
}

func (l *logger) Info(message logrus.Fields, msg string) {
	if message != nil {
		l.Log = l.Log.WithFields(message)
	}
	l.Log.Info(msg)
}
func (l *logger) Error(message logrus.Fields, msg string) {
	if message != nil {
		l.Log = l.Log.WithFields(message)
	}
	l.Log.Error(msg)
}
