package log

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
var Level *zap.AtomicLevel

func init() {
	var err error
	Logger, Level, err = DefaultSugarLogger()
	if err != nil {
		panic(fmt.Sprintf("Log 初始化失败: %v", err))
	}

}

// Color represents a text color.
type Color uint8

// Add adds the coloring to the given string.
func (c Color) Add(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
}

const (
	Black Color = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

type Config struct {
	Format       string                 `json:"format"` //console 或 json
	Level        string                 `json:"level"`
	Path         string                 `json:"path,omitempty"`
	Development  bool                   `json:"development,omitempty"`
	DefaultFiled map[string]interface{} `json:"defaultFiled,omitempty"`
}

func DefaultSugarLogger() (*zap.SugaredLogger, *zap.AtomicLevel, error) {
	// 默认只输出info级别到标准输出
	log, level, err := BuildSugarLogger(Config{"console", "info", "", false, map[string]interface{}{}})

	return log, level, err

}
func SetUpLog(c Config) error {

	log, level, err := BuildSugarLogger(c)
	Logger = log
	Level = level
	return err
}

func ShortColorCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(Green.Add(caller.TrimmedPath()))
}

func TimeLayoutEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}
	layout := "2006-01-02T15:04:05.000Z0700"
	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}
	enc.AppendString(Cyan.Add(t.Format(layout)))
}
func buildBaseEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func buildConsoleEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := buildBaseEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeCaller = ShortColorCallerEncoder
	//encoderConfig.EncodeTime = TimeLayoutEncoder
	return encoderConfig
}

func buildAtomicLevel(level string) (*zap.AtomicLevel, error) {
	var l zapcore.Level
	err := l.Set(level)
	if err != nil {
		return nil, err
	}
	logLevel := zap.NewAtomicLevelAt(l)
	return &logLevel, nil
}
func BuildSugarLogger(c Config) (*zap.SugaredLogger, *zap.AtomicLevel, error) {
	logLevel, err := buildAtomicLevel(c.Level)
	if err != nil {
		return nil, nil, err
	}
	output := []string{"stdout"}
	if c.Path != "" {
		if path.IsAbs(c.Path) {
			output = append(output, c.Path)
		} else {
			dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				return nil, nil, err
			}
			output = append(output, filepath.Join(dir, c.Path))

		}
	}

	var encoderConfig zapcore.EncoderConfig

	switch c.Format {
	case "json":
		encoderConfig = buildBaseEncoderConfig()

	case "console":
		encoderConfig = buildConsoleEncoderConfig()

	default:
		encoderConfig = buildBaseEncoderConfig()

	}

	zapConfig := zap.Config{
		Level:            *logLevel, // 日志级别
		Development:      c.Development,
		Encoding:         c.Format,       // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,  // 编码器配置
		InitialFields:    c.DefaultFiled, // 初始化字段
		OutputPaths:      output,
		ErrorOutputPaths: output,
	}
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, nil, err
	}

	return logger.Sugar(), logLevel, nil

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

func SetRotateLog(c Config, rotateType string) error {

	logLevel, err := buildAtomicLevel(c.Level)
	if err != nil {
		return err
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

	cores := getZapCores(c.Format, logFile, rotateType, logLevel)

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	Logger = logger.Sugar()
	Level = logLevel

	return nil

}

func getZapCores(format string, logFileName string, rotateType string, settingLogLevel *zap.AtomicLevel) []zapcore.Core {

	stdout := os.Stdout
	var cores []zapcore.Core
	consoleEncoder := zapcore.NewConsoleEncoder(buildConsoleEncoderConfig())
	var baseEncoder zapcore.Encoder
	switch format {

	case "console":
		baseEncoder = consoleEncoder
	case "json":
		baseEncoder = zapcore.NewJSONEncoder(buildBaseEncoderConfig())

	default:
		baseEncoder = consoleEncoder

	}
	switch rotateType {
	case "time":
		logfile := getWriter(logFileName)

		cores = append(cores,
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(stdout), settingLogLevel),
			zapcore.NewCore(baseEncoder, zapcore.AddSync(logfile), settingLogLevel),
		)
	case "chunk":
		hook := &lumberjack.Logger{
			Filename:   logFileName,
			MaxSize:    1024, // MB
			MaxBackups: 10000,
			MaxAge:     365,   //days
			Compress:   false, // disabled by default
		}

		cores = append(cores,
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(stdout), settingLogLevel),
			zapcore.NewCore(baseEncoder, zapcore.AddSync(hook), settingLogLevel),
		)

	}

	return cores

}
