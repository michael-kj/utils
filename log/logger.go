package log

import "C"
import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.SugaredLogger

func init() {
	var err error
	Logger, err = DefaultSugarLogger()
	if err != nil {
		panic(fmt.Sprintf("Log 初始化失败: %v", err))
	}

}

type Config struct {
	Format       string                 `json:"format"` //console 或 json
	Level        string                 `json:"level"`
	Path         string                 `json:"path,omitempty"`
	Development  bool                   `json:"development,omitempty"`
	DefaultFiled map[string]interface{} `json:"defaultFiled,omitempty"`
}

func DefaultSugarLogger() (*zap.SugaredLogger, error) {
	// 默认只输出info级别到标准输出
	log, err := BuildSugarLogger(Config{"console", "info", "", false, map[string]interface{}{}})

	return log, err

}
func SetUpLog(c Config) error {

	log, err := BuildSugarLogger(c)
	Logger = log
	return err
}

func BuildSugarLogger(c Config) (*zap.SugaredLogger, error) {

	var l zapcore.Level
	err := l.Set(c.Level)
	if err != nil {
		return nil, err
	}
	output := []string{"stdout"}
	if c.Path != "" {
		if path.IsAbs(c.Path) {
			output = append(output, c.Path)
		} else {
			dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				return nil, err
			}
			output = append(output, filepath.Join(dir, c.Path))

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
		Development:      c.Development,
		Encoding:         c.Format,       // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,  // 编码器配置
		InitialFields:    c.DefaultFiled, // 初始化字段
		OutputPaths:      output,
		ErrorOutputPaths: output,
	}
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil

}

func getWriter(output string) io.Writer {
	logFile := ""
	if output == "" {
		panic(errors.New("path for log file is empty "))
	}

	if path.IsAbs(output) {
		logFile = output
	} else {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			panic(err)
		}
		logFile = filepath.Join(dir, output)

	}

	hook, err := rotatelogs.New(
		logFile+".%Y-%m-%d_%H:%M:%S",
		rotatelogs.WithLinkName(logFile),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithLocation(time.Local),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func enableLevel(settingLevel zapcore.Level) zap.LevelEnablerFunc {
	return zap.LevelEnablerFunc(func(logLevel zapcore.Level) bool {
		return logLevel >= settingLevel
	})
}

func SetRotateLog(c Config, logProvider string) error {
	var l zapcore.Level
	err := l.Set(c.Level)
	if err != nil {
		return err
	}

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
	var encoder zapcore.Encoder

	switch c.Format {

	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "json":

		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		return errors.New("wrong log format set ,only accept value: console or json ")

	}

	logFile := ""
	if c.Path == "" {
		panic(errors.New("path for log file is empty "))
	}

	if path.IsAbs(c.Path) {
		logFile = c.Path
	} else {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			panic(err)
		}
		logFile = filepath.Join(dir, c.Path)

	}

	cores := getZapCores(encoder, logFile, logProvider, l)

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	Logger = logger.Sugar()
	return nil

}

func getZapCores(encoder zapcore.Encoder, logFileName string, logProvider string, settingLogLevel zapcore.Level) []zapcore.Core {

	stdout := os.Stdout
	var cores []zapcore.Core
	if logProvider == "v1" {
		//file-rotatelogs的文件名有bug，时分秒不消失，但是如果按照天来切分的话   问题不大
		logfile := getWriter(logFileName)

		cores = append(cores,
			zapcore.NewCore(encoder, zapcore.AddSync(stdout), enableLevel(settingLogLevel)),
			zapcore.NewCore(encoder, zapcore.AddSync(logfile), enableLevel(settingLogLevel)),
		)

	} else {
		hook := &lumberjack.Logger{
			Filename:   logFileName,
			MaxSize:    1024, // MB
			MaxBackups: 10000,
			MaxAge:     365,   //days
			Compress:   false, // disabled by default
		}

		cores = append(cores,
			zapcore.NewCore(encoder, zapcore.AddSync(stdout), enableLevel(settingLogLevel)),
			zapcore.NewCore(encoder, zapcore.AddSync(hook), enableLevel(settingLogLevel)),
		)
	}
	return cores

}
