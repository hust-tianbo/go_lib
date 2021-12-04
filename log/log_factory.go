package log

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	RegisterWriter(OutputConsole, DefaultConsoleWriterFactory)
	RegisterWriter(OutputFile, DefaultFileWriterFactory)
	DefaultLogger = NewZapLog(defaultConfig)
}

var (
	writers = make(map[string]FactoryInterface)
	logs    = make(map[string]Logger)

	DefaultLogFactory           = &Factory{}
	DefaultConsoleWriterFactory = &ConsoleWriterFactory{}
	DefaultFileWriterFactory    = &FileWriterFactory{}
)

type FactoryInterface interface {
	Setup(name string, configDec DecodeInterface) error
}

func Register(name string, logger Logger) {
	logs[name] = logger
}

// 获取句柄
func Get(name string) Logger {
	return logs[name]
}

func RegisterWriter(name string, writer FactoryInterface) {
	writers[name] = writer
}

type Factory struct {
}

func (f *Factory) Setup(name string, configDec DecodeInterface) error {
	if configDec == nil {
		return errors.New("log config decoder empty")
	}

	conf, callerSkip, err := f.setupConfig(configDec)
	if err != nil {
		return err
	}

	logger := NewZapLogWithCallerSkip(conf, callerSkip)
	if logger == nil {
		return errors.New("new zap logger fail")
	}

	Register(name, logger)

	if name == "default" {
		SetLogger(logger)
	}

	return nil
}

func (f *Factory) setupConfig(decoder DecodeInterface) (Config, int, error) {
	conf := Config{}

	err := decoder.Decode(&conf)
	if err != nil {
		return nil, 0, err
	}

	if len(conf) == 0 {
		return nil, 0, errors.New("log config output empty")
	}

	callerSkip := 2
	for i := 0; i < len(conf); i++ {
		if conf[i].CallerSkip != 0 {
			callerSkip = conf[i].CallerSkip
		}
	}
	return conf, callerSkip, nil
}

type DecodeInterface interface {
	Decode(interface{}) error
}

// Decoder log
type Decoder struct {
	OutputConfig *OutputConfig
	Core         zapcore.Core
	ZapLevel     zap.AtomicLevel
}

// Decode 解析writer配置 复制一份
func (d *Decoder) Decode(conf interface{}) error {

	output, ok := conf.(**OutputConfig)
	if !ok {
		return fmt.Errorf("decoder config type:%T invalid, not **OutputConfig", conf)
	}

	*output = d.OutputConfig

	return nil
}

type ConsoleWriterFactory struct {
}

// Setup 启动加载配置 并注册console output writer
func (f *ConsoleWriterFactory) Setup(name string, configDec DecodeInterface) error {

	if configDec == nil {
		return errors.New("console writer decoder empty")
	}
	decoder, ok := configDec.(*Decoder)
	if !ok {
		return errors.New("console writer log decoder type invalid")
	}

	conf := &OutputConfig{}
	err := decoder.Decode(&conf)
	if err != nil {
		return err
	}

	decoder.Core, decoder.ZapLevel = newConsoleCore(conf)
	return nil
}

// FileWriterFactory  new file writer instance
type FileWriterFactory struct {
}

// Setup 启动加载配置 并注册file output writer
func (f *FileWriterFactory) Setup(name string, configDec DecodeInterface) error {

	if configDec == nil {
		return errors.New("file writer decoder empty")
	}

	decoder, ok := configDec.(*Decoder)
	if !ok {
		return errors.New("file writer log decoder type invalid")
	}

	err := f.setupConfig(decoder)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileWriterFactory) setupConfig(decoder *Decoder) error {
	conf := &OutputConfig{}
	err := decoder.Decode(&conf)
	if err != nil {
		return err
	}

	if conf.WriteConfig.LogPath != "" {
		conf.WriteConfig.Filename = filepath.Join(conf.WriteConfig.LogPath, conf.WriteConfig.Filename)
		log.Printf("[setupConfig]logpath:%+v,%+v,%+v",
			conf.WriteConfig.Filename, conf.WriteConfig.LogPath, conf.WriteConfig.Filename)
	}

	if conf.WriteConfig.RollType == "" {
		conf.WriteConfig.RollType = RollBySize
	}

	if conf.WriteConfig.WriteMode == 0 {
		conf.WriteConfig.WriteMode = WriteFast // 默认极速写模式，性能更好，日志满丢弃，防止阻塞服务
	}

	decoder.Core, decoder.ZapLevel = newFileCore(conf)
	return nil
}
