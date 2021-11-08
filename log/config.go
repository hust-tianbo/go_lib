package log

// Config log config每个log可以支持多个output
type Config []OutputConfig

type OutputConfig struct {
	Writer      string
	WriteConfig WriteConfig `yaml:"writer_config"`

	Formatter    string
	FormatConfig FormatConfig `yaml:"formatter_config"`

	// RemoteConfig 远程日志格式 配置格式业务随便定 由第三方组件自己注册
	// RemoteConfig yaml.Node `yaml:"remote_config"`

	// Level 控制日志级别 debug info error
	Level string

	// CallerSkip 控制log函数嵌套深度
	CallerSkip int `yaml:"caller_skip"`
}

type TimeSplit string

const (
	// Hour 按小时分割
	Hour = "hour"

	// Day 按天分割
	Day = "day"

	// Month 按月分割
	Month = "month"

	// Year 按年分割
	Year = "year"
)

type WriteConfig struct {
	// LogPath 日志路径名
	LogPath string `yaml:"log_path"`
	// Filename 日志路径文件名
	FileName string `yaml:"filename"`
	// WriteMode 日志写入模式 1.同步，2.异步
	WriteMode int `yaml:"write_mode"`
	// RollType 文件滚动类型，按大小分割文件，按时间分割文件
	RollType string `yaml:"roll_type"`
	// MaxDay 日志最大保留天数
	MaxDay int `yaml:"max_day"`
	// MaxHistory 日志最大历史文件数
	MaxHistory int `yaml:"max_history"`

	// 是否压缩
	Compress bool `yaml:"compress"`

	// MaxSize  日志最大大小
	MaxSize int `yam:"max_size"`

	// 按时间分割时，作为时间分割文件的时间单位
	TimeSplit TimeSplit `yaml:"time_split"`
}

type FormatConfig struct {
	// TimeFmt 日志输出时间格式
	TimeFmt string `yaml:"time_fmt"`

	// TimeKey 日志输出时间Key
	TimeKey string `yaml:"time_key"`

	// LevelKey 日志级别输出Key
	LevelKey string `yaml:"level_key"`

	// NameKey 日志名称Key
	NameKey string `yaml:"name_Key"`

	// CallerKey 日志输出调用者Key
	CallerKey string `yaml:"caller_key"`

	// MessageKey 日志输出消息体Key
	MessageKey string `yaml:"message_key"`

	// StackTraceKey 日志输出堆栈trace key
	StacktraceKey string `yaml:"stacktrace_Key"`
}
