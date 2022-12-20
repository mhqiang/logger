package logger

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/mcuadros/go-defaults"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Option func(c *Config)

// Config .
type Config struct {
	Level          string `default:"info" toml:"level"`
	LogPath        string `default:"logs" toml:"log_path"`
	MaxLogSize     int    `default:"100" toml:"max_log_size"`
	ServiceName    string `default:"test" toml:"service_name,omitempty"`
	InfoOutput     string `toml:"info_log_file"`
	ErrorOutput    string `toml:"error_log_file"`
	DebugOutput    string `toml:"debug_log_file"`
	NotDisplayLine bool   `default:"false" toml:"not_display_file_linenum"`
	Stdout         bool   `default:"true" toml:"not_stdout"`

	MaxBackup int  `default:"100" toml:"max_backup"`
	MaxAge    int  `default:"7" toml:"max_age"`
	Compress  bool `default:"true" toml:"compress"`
}

func (config *Config) NewLogger() error {

	level := new(zapcore.Level)
	err := level.UnmarshalText([]byte(config.Level))
	if err != nil {
		return err
	}

	maxSize := config.MaxLogSize
	maxBackups := config.MaxLogSize
	maxAge := config.MaxAge
	compress := config.Compress

	var infoPath, debugPath, errPath string

	if _, err := os.Stat(config.LogPath); os.IsNotExist(err) {
		os.Mkdir(config.LogPath, 0755)
	}

	if len(config.LogPath) == 0 {
		config.LogPath = "logs"
	}
	infoPath = fmt.Sprintf("%s/%v.log", config.LogPath, config.ServiceName)
	if config.InfoOutput != "" {
		infoPath = config.InfoOutput
	}
	sugarInfoLogger = createLogger(infoPath, *level,
		maxSize, maxBackups, maxAge, compress, config.Stdout, config.NotDisplayLine)
	sugarInfoPath = infoPath

	sugarDebugLogger = sugarInfoLogger
	sugarDebugPath = infoPath

	sugarErrorLogger = sugarInfoLogger
	sugarErrPath = infoPath

	sugarWarnLogger = sugarInfoLogger
	sugarWarnPath = infoPath

	if config.DebugOutput != "" {
		debugPath = config.DebugOutput
		sugarDebugLogger = createLogger(debugPath, *level,
			maxSize, maxBackups, maxAge, compress, config.Stdout, config.NotDisplayLine)
		sugarDebugPath = debugPath
	}

	if config.ErrorOutput != "" {
		errPath = config.ErrorOutput
		sugarErrorLogger = createLogger(errPath, *level,
			maxSize, maxBackups, maxAge, compress, config.Stdout, config.NotDisplayLine)
		sugarErrPath = errPath
	}

	// logger = zap.New(core, zap.AddCaller(), zap.Development(), zap.Fields(zap.String("serviceName", serviceName)))
	return nil
}

func WithNotDisplayLineNum(flag bool) Option {
	return func(c *Config) {
		c.NotDisplayLine = flag
	}
}

func WithStdout(flag bool) Option {
	return func(c *Config) {
		c.Stdout = flag
	}
}

func WithLevel(level string) Option {
	return func(c *Config) {
		c.Level = level
	}
}

func WithInfoPath(path string) Option {
	return func(c *Config) {
		c.LogPath = path
	}
}

func WithDebugPath(path string) Option {
	return func(c *Config) {
		c.DebugOutput = path
	}
}

func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

func WithErrPath(path string) Option {
	return func(c *Config) {
		c.ErrorOutput = path
	}
}

func InitDefaultLogger(ops ...Option) {
	DefaultConfig = new(Config)
	defaults.SetDefaults(DefaultConfig)
	for _, op := range ops {
		op(DefaultConfig)
	}

	DefaultConfig.NewLogger()
}

// var logger *zap.Logger
var (
	DefaultConfig *Config

	sugarInfoLogger  *zap.SugaredLogger
	sugarInfoPath    string
	sugarDebugLogger *zap.SugaredLogger
	sugarDebugPath   string
	sugarErrorLogger *zap.SugaredLogger
	sugarErrPath     string
	sugarWarnLogger  *zap.SugaredLogger
	sugarWarnPath    string
)

func GetInfoLogPath() string {
	return sugarInfoPath
}

func GetDebugLogPath() string {
	return sugarDebugPath
}

func GetErrLogPath() string {
	return sugarErrPath
}

func GetWarnLogPath() string {
	return sugarWarnPath
}

func formatArgs(v ...interface{}) string {
	var formatStrings []string
	for i := 0; i < len(v); i++ {
		t := v[i]
		switch reflect.TypeOf(t).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				formatStrings = append(formatStrings, `%v`)
			}
		}

	}
	// fmt.Println(v, len(v), formatStrings)
	return strings.Join(formatStrings, " ")
}

func Info(v ...interface{}) {
	format := formatArgs(v)
	sugarInfoLogger.Info("", fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	format := formatArgs(v)
	sugarErrorLogger.Error("", fmt.Sprintf(format, v...))
}

func Errorln(v ...interface{}) {
	sugarErrorLogger.Error("", fmt.Sprintln(v...))
}

func Warn(v ...interface{}) {
	format := formatArgs(v)
	sugarWarnLogger.Warn("", fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	format := formatArgs(v)
	sugarDebugLogger.Debug("", fmt.Sprintf(format, v...))
}

func Panic(v ...interface{}) {
	format := formatArgs(v)
	sugarErrorLogger.Panic("", fmt.Sprintf(format, v...))
}

func createLogger(path string, level zapcore.Level, maxSize int, maxBackups int,
	maxAge int, compress, stdout, notDisplayLine bool) *zap.SugaredLogger {
	core := newCore(path, level, maxSize, maxBackups, maxAge, compress, stdout)

	var logger *zap.Logger
	if !notDisplayLine {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		logger = zap.New(core, zap.AddCaller(), zap.WithCaller(false))
	}

	// logger := zap.New(core, zap.AddCaller(), zap.WithCaller(false))
	return logger.Sugar()
}

/**
 * zapcore构造
 */
func newCore(filePath string, level zapcore.Level, maxSize int, maxBackups int,
	maxAge int, compress, stdout bool) zapcore.Core {
	//日志文件路径配置2
	hook := lumberjack.Logger{
		Filename:   filePath,   // 日志文件路径
		MaxSize:    maxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: maxBackups, // 日志文件最多保存多少个备份
		MaxAge:     maxAge,     // 文件最多保存多少天
		Compress:   compress,   // 是否压缩
	}
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)
	//公用编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	var syncer zapcore.WriteSyncer
	if stdout {
		syncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(&hook), zapcore.AddSync(os.Stdout))
	} else {
		syncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(&hook))
	}
	return zapcore.NewCore(
		// zapcore.NewJSONEncoder(encoderConfig),               // 编码器json配置

		zapcore.NewConsoleEncoder(encoderConfig), // 编码器设置成date  level linenum msg 不需要key
		// zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), // 打印控制台
		syncer,      // 打印文件
		atomicLevel, // 日志级别
	)
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000"))
}
