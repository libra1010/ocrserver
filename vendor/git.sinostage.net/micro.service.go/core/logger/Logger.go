package logger

import (
	"fmt"
	"git.sinostage.net/micro.service.go/core/util"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type LogFileConfig struct {
	Enable bool
	//日志文件路径名称
	FileName string
	//日志最大文件，单位M 默认 10
	MaxSize int
	//日志对多允许备份数量 默认 30
	MaxBackups int
	//日志最大保存天数 默认 7
	MaxAge int
	//是否压缩 默认 true
	Compress bool
}

type logKafka struct {
	producer sarama.SyncProducer
	topic    string
}

var logchan = make(chan *sarama.ProducerMessage, 20)
var sendBuffer = make(chan *sarama.ProducerMessage, 1000)

func (lk *logKafka) Write(p []byte) (n int, err error) {
	if p == nil || len(p) <= 0 {
		return 0, nil
	}
	msg := &sarama.ProducerMessage{}
	msg.Topic = lk.topic
	msg.Value = sarama.ByteEncoder(p)
	logchan <- msg

	//go func() {
	//	lk.producer.SendMessage(&sarama.ProducerMessage{Topic: lk.topic, Key: nil, Value: sarama.StringEncoder(p)})
	//}()
	//lk.producer.Input() <- &sarama.ProducerMessage{Topic: lk.topic, Key: nil, Value: sarama.StringEncoder(p)}

	return len(p), nil
}

const (
	LEVEL_DEBUG = "debug"
	LEVEL_INFO  = "info"
	LEVEL_WARN  = "warn"
	LEVEL_ERROR = "error"

	ENCODER_JSON    = "json"
	ENCODER_CONSOLE = "console"
)

type LogConfig struct {
	//日志级别
	Level string
	//日志编码方式
	Encoder string
}

type LogKafkaConfig struct {
	Server   string
	Account  string
	Password string
	Enable   bool
	Topic    string
}

var logStruct = &LogConfig{
	Level:   LEVEL_INFO,
	Encoder: ENCODER_CONSOLE,
}
var logFile = &LogFileConfig{
	Enable:     true,
	FileName:   "log.log",
	MaxSize:    10,
	MaxBackups: 30,
	MaxAge:     7,
	Compress:   true,
}
var logKafkaConfig = &LogKafkaConfig{
	Enable: false,
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

var config map[string]string

func init() {
	config = koloCore.ConfUtil.ReadConfig("log.conf")

	logConfigLoad()
	fileConfigLoad()
	kafkaConfigLoad()
	fmt.Println("文件配置: ", logFile)
	logInit()
}

func logConfigLoad() {
	level := config["log.level"]
	switch level {
	case LEVEL_INFO:
		logStruct.Level = LEVEL_INFO
	case LEVEL_DEBUG:
		logStruct.Level = LEVEL_DEBUG
	case LEVEL_WARN:
		logStruct.Level = LEVEL_WARN
	case LEVEL_ERROR:
		logStruct.Level = LEVEL_ERROR
	default:
		logStruct.Level = LEVEL_INFO
	}

	e := config["log.encoder"]

	switch e {
	case ENCODER_CONSOLE:
		logStruct.Encoder = ENCODER_CONSOLE
	case ENCODER_JSON:
		logStruct.Encoder = ENCODER_JSON
	default:
		logStruct.Encoder = ENCODER_CONSOLE
	}
}

var log *zap.Logger

func logInit() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     timeEncoder,                    // ISO8601 UTC 时间格式 zapcore.ISO8601TimeEncoder
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()

	switch logStruct.Level {
	case LEVEL_INFO:
		atomicLevel.SetLevel(zap.InfoLevel)
	case LEVEL_DEBUG:
		atomicLevel.SetLevel(zap.DebugLevel)
	case LEVEL_WARN:
		atomicLevel.SetLevel(zap.WarnLevel)
	case LEVEL_ERROR:
		atomicLevel.SetLevel(zap.ErrorLevel)
	default:
		atomicLevel.SetLevel(zap.InfoLevel)
	}

	var core zapcore.Core
	var encod zapcore.Encoder
	if logStruct.Encoder == ENCODER_CONSOLE {
		encod = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encod = zapcore.NewJSONEncoder(encoderConfig)
	}

	index := 0
	writes := make([]zapcore.WriteSyncer, 3)
	writes[index] = zapcore.AddSync(os.Stdout)
	index++
	//如果开启了文本日志
	if logFile.Enable {
		writes[index] = zapcore.AddSync(fileHook())
		index++
	}

	if logKafkaConfig.Enable && logKafkaConfig.Topic != "" && logKafkaConfig.Server != "" {
		kafkaWrite := setKafka()
		if kafkaWrite != nil {
			writes[index] = kafkaWrite
			index++
		}
	}

	core = zapcore.NewCore(
		encod, // 编码器配置
		zapcore.NewMultiWriteSyncer(writes[:index]...), // 打印到控制台和文件
		atomicLevel, // 日志级别
	)

	// 开启开发模式，堆栈跟踪
	//caller := zap.AddCaller()
	// 开启文件及行号
	//development := zap.Development()

	// 构造日志
	//Log = zap.New(core, caller, development)
	// 构造日志
	log = zap.New(core)
}

func setKafka() zapcore.WriteSyncer {
	if !logKafkaConfig.Enable || logKafkaConfig.Topic == "" || logKafkaConfig.Server == "" {
		return nil
	}
	var (
		kl  logKafka
		err error
	)
	kl.topic = logKafkaConfig.Topic

	kconfig := sarama.NewConfig()
	kconfig.Producer.Partitioner = sarama.NewRandomPartitioner
	kconfig.Net.SASL.User = logKafkaConfig.Account
	kconfig.Net.SASL.Password = logKafkaConfig.Password
	kconfig.Net.SASL.Enable = true

	kconfig.Producer.Return.Errors = true
	kconfig.Producer.Return.Successes = true
	kconfig.Version = sarama.MaxVersion
	producer, err := sarama.NewSyncProducer(strings.Split(logKafkaConfig.Server, ","), kconfig)
	//producer, err := sarama.NewAsyncProducer(strings.Split(logKafkaConfig.Server, ","), kconfig)

	if err != nil {
		fmt.Println("kafka日志设置异常: ", err)
	}

	kl.producer = producer

	go func() {
		select {
		case msg := <-sendBuffer:
			_, _, err := kl.producer.SendMessage(msg)
			if err != nil {
				fmt.Println("发送至队列发生异常: ", err)
			}
		}
	}()

	go func() {
		select {
		case msg := <-logchan:
			if len(sendBuffer) >= 1000 {
				fmt.Println("消息已满，丢弃消息")
			} else {
				sendBuffer <- msg
			}
		}
	}()

	//go func(waitMsg logKafka) {
	//	// wait response
	//	select {
	//	case msg := <-waitMsg.producer.Successes():
	//		fmt.Println("Produced message successes", msg.Offset)
	//		//log.Printf("Produced message successes: [%s]\n", msg.Value)
	//	case err := <-waitMsg.producer.Errors():
	//		fmt.Println("Produced message failure", err)
	//		//log.Println("Produced message failure: ", err)
	//	default:
	//		fmt.Println("Produced message not result")
	//		//log.Println("Produced message not result")
	//	}
	//}(kl)

	return zapcore.AddSync(&kl)
}

func fileConfigLoad() {

	setStringToBool("log.file.enable", func(val bool) {
		logFile.Enable = val
	}, "获取文件是否启用配置错误,保持默认启用")

	if !logFile.Enable {
		return
	}

	name := config["log.file.name"]

	if strings.TrimSpace(name) == "" {
		fmt.Println("文件名称未配置，保持默认")
	} else {
		logFile.FileName = name
	}

	setStringToInt("log.file.maxSize", func(val int) {
		logFile.MaxSize = val
	}, "文件最大数量配置错误，保持默认")

	setStringToInt("log.file.maxBackups", func(val int) {
		logFile.MaxBackups = val
	}, "文件最大备份数配置错误，保持默认")

	setStringToInt("log.file.maxAge", func(val int) {
		logFile.MaxAge = val
	}, "文件最大保持天数配置错误，保持默认")

	setStringToBool("log.file.compress", func(val bool) {
		logFile.Compress = val
	}, "获取文件是否压缩配置错误,保持默认")
}

func kafkaConfigLoad() {
	setStringToBool("log.kafka.enable", func(val bool) {
		logKafkaConfig.Enable = val
	}, "获取kafka是否启用配置错误,保持默认启用")

	if !logKafkaConfig.Enable {
		return
	}

	logKafkaConfig.Server = config["log.kafka.server"]
	logKafkaConfig.Account = config["log.kafka.account"]
	logKafkaConfig.Password = config["log.kafka.password"]
	logKafkaConfig.Topic = config["log.kafka.topic"]
}

func setStringToBool(key string, setFun func(val bool), msg string) {
	if val, ok := config[key]; ok {
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			fmt.Println(msg, err, val)
		} else {
			setFun(boolVal)
		}
	}
}

func setStringToInt(key string, setFun func(val int), msg string) {
	if val, ok := config[key]; ok {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			fmt.Println(msg, err, val)
		} else {
			setFun(intVal)
		}
	}
}

func fileHook() *lumberjack.Logger {
	hook := lumberjack.Logger{
		Filename:   logFile.FileName,   // 日志文件路径
		MaxSize:    logFile.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: logFile.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     logFile.MaxAge,     // 文件最多保存多少天
		Compress:   logFile.Compress,   // 是否压缩
	}

	return &hook
}

func Debug(msg string, args ...interface{}) {
	val := msg
	for _, a := range args {
		val += " " + fmt.Sprint(a) + " "
	}

	log.Debug(val, getCaller())
}

func Info(msg string, args ...interface{}) {
	val := msg
	for _, a := range args {
		val += " " + fmt.Sprint(a) + " "
	}

	log.Info(val, getCaller())
}

func Warn(msg string, args ...interface{}) {
	val := msg
	for _, a := range args {
		val += " " + fmt.Sprint(a) + " "
	}

	log.Warn(val, getCaller())
}

func Error(msg string, err error, args ...interface{}) {
	val := msg

	if err != nil {
		val += "\r\n" + err.Error() + "\r\n"
	}

	for _, a := range args {
		val += "  " + fmt.Sprint(a)
	}

	log.Error(val, getCaller())
}

func getCaller() zap.Field {
	_, file, line, _ := runtime.Caller(2)
	// 设置初始化字段
	filed := zap.String("file", file+":"+strconv.Itoa(line))
	return filed
}
