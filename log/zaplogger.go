package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Levels zapcore level
var Levels = map[string]zapcore.Level{
	"":      zapcore.DebugLevel,
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

var levelToZapLevel = map[Level]zapcore.Level{
	LevelTrace: zapcore.DebugLevel,
	LevelDebug: zapcore.DebugLevel,
	LevelInfo:  zapcore.InfoLevel,
	LevelWarn:  zapcore.WarnLevel,
	LevelError: zapcore.ErrorLevel,
	LevelFatal: zapcore.FatalLevel,
}

var zapLevelToLevel = map[zapcore.Level]Level{
	zapcore.DebugLevel: LevelDebug,
	zapcore.InfoLevel:  LevelInfo,
	zapcore.WarnLevel:  LevelWarn,
	zapcore.ErrorLevel: LevelError,
	zapcore.FatalLevel: LevelFatal,
}

// NewZapLog 创建一个trpc框架zap默认实现的logger, callerskip为2
func NewZapLog(c Config) Logger {
	return NewZapLogWithCallerSkip(c, 2)
}

// NewZapLogWithCallerSkip 创建一个trpc框架zap默认实现的logger
func NewZapLogWithCallerSkip(c Config, callerSkip int) Logger {

	cores := make([]zapcore.Core, 0, len(c))
	levels := make([]zap.AtomicLevel, 0, len(c))
	for _, o := range c {
		writer, ok := writers[o.Writer]
		if !ok {
			fmt.Printf("log writer core:%s no registered!\n", o.Writer)
			return nil
		}

		decoder := &Decoder{OutputConfig: &o}
		err := writer.Setup(o.Writer, decoder)
		if err != nil {
			fmt.Printf("log writer setup core:%s fail:%v!\n", o.Writer, err)
			return nil
		}

		cores = append(cores, decoder.Core)
		levels = append(levels, decoder.ZapLevel)
	}

	logger := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCallerSkip(callerSkip),
		zap.AddCaller(),
	)

	// 收集标准库log的标准输出
	zap.RedirectStdLog(logger)

	return &zapLog{
		levels: levels,
		logger: logger,
	}
}
