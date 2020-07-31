package log

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func init() {
	var err error
	Logger, err = DefaultSugarLogger()
	if err != nil {
		panic(fmt.Sprintf("Log 初始化失败: %v", err))

	}

}

func DefaultSugarLogger() (*zap.SugaredLogger, error) {
	log, err := BuildSugarLogger("console", "info", "", false, map[string]interface{}{})

	return log, err

}
func SetUp(logFormat, level, logPath string, development bool, defaultFiled map[string]interface{}) error {

	log, err := BuildSugarLogger(logFormat, level, logPath, development, defaultFiled)
	Logger = log
	return err
}

func BuildSugarLogger(logFormat, level, logPath string, development bool, defaultFiled map[string]interface{}) (*zap.SugaredLogger, error) {

	var l zapcore.Level
	err := l.Set(level)
	if err != nil {
		return nil, err
	}
	output := []string{"stdout"}
	if logPath != "" {
		if path.IsAbs(logPath) {
			output = append(output, logPath)
		} else {
			dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				return nil, err
			}
			output = append(output, filepath.Join(dir, logPath))

		}
	}

	logLevel := zap.NewAtomicLevelAt(l)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	zapConfig := zap.Config{
		Level:            logLevel, // 日志级别
		Development:      development,
		Encoding:         logFormat,     // 输出格式 console 或 json
		EncoderConfig:    encoderConfig, // 编码器配置
		InitialFields:    defaultFiled,  // 初始化字段
		OutputPaths:      output,
		ErrorOutputPaths: output,
	}
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil

}
