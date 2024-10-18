package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

// 定义颜色代码
const (
	reset       = "\033[0m"
	red         = "\033[31m"
	green       = "\033[32m"
	yellow      = "\033[33m"
	blue        = "\033[34m"
	magenta     = "\033[35m"
	cyan        = "\033[36m"
	white       = "\033[37m"
	boldRed     = "\033[1;31m"
	boldGreen   = "\033[1;32m"
	boldYellow  = "\033[1;33m"
	boldBlue    = "\033[1;34m"
	boldMagenta = "\033[1;35m"
	boldCyan    = "\033[1;36m"
	boldWhite   = "\033[1;37m"
)

// 获取日志级别对应的颜色
func getLevelColor(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return magenta
	case zapcore.InfoLevel:
		return blue
	case zapcore.WarnLevel:
		return yellow
	case zapcore.ErrorLevel:
		return red
	case zapcore.FatalLevel, zapcore.PanicLevel:
		return boldRed
	default:
		return white
	}
}

// 获取状态码对应的颜色
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return green
	case status >= 300 && status < 400:
		return boldBlue
	case status >= 400 && status < 500:
		return yellow
	default:
		return red
	}
}

// 自定义编码器
type customEncoder struct {
	zapcore.Encoder
	zapcore.EncoderConfig
	isConsole bool
}

func (e *customEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	var msg strings.Builder

	if strings.HasPrefix(entry.Message, "[GIN]") {
		// GIN日志格式
		var status, method, path, latency, ip, ua string
		statusCode := 0

		for _, field := range fields {
			switch field.Key {
			case "status":
				statusCode = int(field.Integer)
				status = fmt.Sprintf("%d", statusCode)
			case "method":
				method = field.String
			case "path":
				path = field.String
			case "latency":
				latency = formatLatency(time.Duration(field.Integer))
			case "ip":
				ip = field.String
			case "user-agent":
				ua = field.String
			}
		}

		if e.isConsole {
			msg.WriteString(fmt.Sprintf("%s%s%s %s%s%s %s%s%s %s%s%s %s%s%s %s%s%s %s%s%s %s%s%s %s%s%s",
				cyan, timestamp, reset,
				getLevelColor(entry.Level), strings.ToUpper(entry.Level.String()), reset,
				boldWhite, "GIN", reset,
				getStatusColor(statusCode), status, reset,
				boldGreen, method, reset,
				yellow, latency, reset,
				boldCyan, path, reset,
				magenta, ip, reset,
				white, ua, reset,
			))
		} else {
			msg.WriteString(fmt.Sprintf("%s %s GIN %s %s %s %s %s %s",
				timestamp,
				strings.ToUpper(entry.Level.String()),
				status,
				method,
				latency,
				path,
				ip,
				ua,
			))
		}
	} else {
		// 普通日志格式
		if e.isConsole {
			msg.WriteString(fmt.Sprintf("%s%s%s %s%s%s %s%s%s %s%s%s",
				cyan, timestamp, reset,
				getLevelColor(entry.Level), strings.ToUpper(entry.Level.String()), reset,
				yellow, entry.Caller.TrimmedPath(), reset,
				white, entry.Message, reset,
			))
		} else {
			msg.WriteString(fmt.Sprintf("%s %s %s %s",
				timestamp,
				strings.ToUpper(entry.Level.String()),
				entry.Caller.TrimmedPath(),
				entry.Message,
			))
		}
	}

	buf := buffer.NewPool().Get()
	buf.AppendString(msg.String() + "\n")
	return buf, nil
}

// formatLatency 格式化延迟时间
func formatLatency(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dμs", d.Microseconds())
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// InitLogger 初始化日志系统
func InitLogger() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 文件输出配置
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    10,
		MaxBackups: 30,
		MaxAge:     7,
		Compress:   true,
	})

	// 控制台输出配置
	consoleWriter := zapcore.AddSync(os.Stdout)

	// 创建自定义编码器
	fileEnc := &customEncoder{EncoderConfig: encoderConfig, isConsole: false}
	consoleEnc := &customEncoder{EncoderConfig: encoderConfig, isConsole: true}

	// 设置日志级别
	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	// 创建核心
	core := zapcore.NewTee(
		zapcore.NewCore(fileEnc, fileWriter, level),
		zapcore.NewCore(consoleEnc, consoleWriter, level),
	)

	// 创建logger
	Log = zap.New(core, zap.AddCaller(), zap.Development())
}

// GinLogger Gin中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}

		c.Next()

		Log.Info("[GIN]",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
		)
	}
}
