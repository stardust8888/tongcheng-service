package logs

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"tongcheng/conf"
	"tongcheng/utils/x/timex"
)

var (
	logger *zap.Logger
)

const (
	RequestIDKey = "RequestID"
	CallerKey    = "Caller"
	PCallerKey   = "Caller1"
	GCallerKey   = "Caller2"
)

func InitLogger(path string) {
	config := conf.GetConfiguration().Log

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    customLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	level := zapcore.Level(config.Level)
	if level < zap.DebugLevel || level > zap.FatalLevel {
		level = zap.DebugLevel // 默认为Debug级别
	}
	atomicLevel.SetLevel(level)

	// 根据output配置决定输出方式
	var writers []zapcore.WriteSyncer

	switch config.Output {
	case "console":
		// 只输出到控制台
		writers = append(writers, zapcore.AddSync(os.Stdout))
	case "file":
		// 只输出到文件
		hook := lumberjack.Logger{
			Filename:   config.Path + path,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
		writers = append(writers, zapcore.AddSync(&hook))
	case "both":
		// 同时输出到控制台和文件
		writers = append(writers, zapcore.AddSync(os.Stdout))
		hook := lumberjack.Logger{
			Filename:   config.Path + path,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
		writers = append(writers, zapcore.AddSync(&hook))
	default:
		// 默认只输出到控制台
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 创建核心
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writers...),
		atomicLevel,
	)

	caller := zap.AddCaller()
	development := zap.Development()
	field := zap.Fields(zap.String("service", conf.GetConfiguration().App.Name))

	logger = zap.New(core, caller, development, field)
	zap.ReplaceGlobals(logger)
}

// 自定义时间编码器
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(timex.TimeMilliFormat))
}

// 自定义级别编码器
func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

// Sync 同步日志缓冲区
func Sync() {
	if logger != nil {
		logger.Sync()
	}
}

func Logger() *zap.Logger {
	return logger
}

func Trace(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	traceID, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return ""
	}
	return traceID
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := buildContextFields(ctx)
	fields = append(contextFields, fields...)
	logger.Debug(msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := buildContextFields(ctx)
	fields = append(contextFields, fields...)
	logger.Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := buildContextFields(ctx)
	fields = append(contextFields, fields...)
	logger.Warn(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := buildContextFields(ctx)
	fields = append(contextFields, fields...)
	logger.Error(msg, fields...)
}

func DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := buildContextFields(ctx)
	fields = append(contextFields, fields...)
	logger.DPanic(msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := buildContextFields(ctx)
	fields = append(contextFields, fields...)
	logger.Panic(msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := buildContextFields(ctx)
	fields = append(contextFields, fields...)
	logger.Fatal(msg, fields...)
}

// 构造包含调用栈信息的字段
func buildContextFields(ctx context.Context) []zap.Field {
	caller, pCaller, gCaller := getCallerInfo(3)
	fields := []zap.Field{
		zap.String(RequestIDKey, Trace(ctx)),
		zap.String(CallerKey, caller),
		zap.String(PCallerKey, pCaller),
		zap.String(GCallerKey, gCaller),
	}
	return fields
}

// 获取调用栈信息
func getCallerInfo(skip int) (string, string, string) {
	return buildCaller(skip), buildCaller(skip + 1), buildCaller(skip + 2)
}

func buildCaller(skip int) string {
	var function string

	pc, file, line, ok := runtime.Caller(skip + 1)
	if ok {
		f := runtime.FuncForPC(pc)
		if f != nil {
			function = f.Name()
		}
	}

	position := extractCallerName(file, 1) + ":" + strconv.Itoa(line) + "(" + extractCallerName(function, 0) + ")"

	return position
}

func extractCallerName(fullName string, ty int) string {
	switch ty {
	case 0:
		for i := len(fullName) - 1; i >= 0; i-- {
			if fullName[i] == '/' {
				fullName = fullName[i+1:]
				break
			}
		}
	case 1:
		if strings.Contains(fullName, conf.GetConfiguration().App.Name) {
			index := strings.Index(fullName, conf.GetConfiguration().App.Name) + len(conf.GetConfiguration().App.Name)
			fullName = fullName[index:]
		} else {
			for i := len(fullName) - 1; i >= 0; i-- {
				if fullName[i] == '/' {
					fullName = fullName[i+1:]
					break
				}
			}
		}
	}

	return fullName
}
