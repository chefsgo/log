package log

import (
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/chefsgo/base"
	"github.com/chefsgo/chef"
)

func init() {
	chef.Register(NAME, module)
}

var (
	module = &Module{
		config: Config{
			Driver: chef.DEFAULT, Level: LevelDebug,
			Json: false, Sync: false, Pool: 1000,
			Format: "%time% [%level%] %body%",
		},
		drivers: map[string]Driver{},
	}
)

type (
	// Level 日志级别，从小到大，数字越小越严重
	Level = int

	// 日志模块定义
	Module struct {
		//mutex 锁
		mutex sync.Mutex

		// 几项运行状态
		connected, initialized, launched bool

		//config 日志配置
		config Config

		//drivers 驱动注册表
		drivers map[string]Driver

		// connect 日志连接
		connect Connect

		waiter sync.WaitGroup

		// logger 日志发送管道
		logger chan *Log

		// signal 信号管道，用于flush缓存区，或是结束循环
		// false 表示flush缓存区
		// true 表示结束关闭循环
		signal chan bool
	}

	// LogConfig 日志模块配置
	Config struct {
		// Driver 日志驱动，默认为 default
		Driver string

		// Level 输出的日志级别
		// fatal, panic, warning, notice, info, trace, debug
		Level Level

		// Json 是否开启json输出模式
		// 开启后，所有日志 body 都会被包装成json格式输出
		Json bool

		// Sync 是否开启同步输出，默认为false，表示异步输出
		// 注意：如果开启同步输出，有可能影响程序性能
		Sync bool

		// Pool 异步缓冲池大小
		Pool int

		// Format 日志输出格式，默认格式为 %time% [%level%] %body%
		// 可选参数，参数使用 %% 包裹，如 %time%
		// time		格式化后的时间，如：2006-01-02 15:03:04.000
		// unix		unix时间戳，如：1650271473
		// level	日志级别，如：TRACE
		// body		日志内容
		Format string `toml:"format"`

		// Setting 是为不同驱动准备的自定义参数
		// 具体参数表，请参考各不同的驱动
		Setting Map `toml:"setting"`
	}

	Log struct {
		format string `json:"-"`
		Time   int64  `json:"time"`
		Level  Level  `json:"level"`
		Body   string `json:"body"`
	}
)

//Driver 为log模块注册驱动
func (this *Module) Driver(name string, driver Driver, override bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if driver == nil {
		panic("Invalid log driver: " + name)
	}

	if override {
		this.drivers[name] = driver
	} else {
		if this.drivers[name] == nil {
			this.drivers[name] = driver
		}
	}
}

func (this *Module) Config(name string, config Config, override bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.config = config
}

//Log format
func (log *Log) Format() string {
	message := log.format

	// message := strings.Replace(format, "%flag%", log.Flag, -1)
	message = strings.Replace(message, "%nano%", strconv.FormatInt(log.Time, 10), -1)
	message = strings.Replace(message, "%time%", time.Unix(0, log.Time).Format("2006-01-02 15:04:05.000"), -1)
	message = strings.Replace(message, "%level%", levels[log.Level], -1)
	// message = strings.Replace(message, "%file%", log.File, -1)
	// message = strings.Replace(message, "%line%", strconv.Itoa(log.Line), -1)
	// message = strings.Replace(message, "%func%", log.Func, -1)
	message = strings.Replace(message, "%body%", log.Body, -1)

	return message
}

// flush 调用连接在write
func (this *Module) write(msg *Log) error {
	if this.connect == nil {
		return errInvalidLogConnection
	}

	//格式传过去
	msg.format = this.config.Format

	return this.connect.Write(msg)
}

//flush 真flush
func (this *Module) flush() {
	if false == this.config.Sync {
		for {
			if len(this.logger) > 0 {
				log := <-this.logger
				this.write(log)
				this.waiter.Done()
			} else {
				break
			}
		}
	}
	this.connect.Flush()
}

// Write 写入日志，对外的，需要处理逻辑
func (this *Module) Write(msg *Log) error {
	if this.config.Level < msg.Level {
		return nil
	}
	if msg.Time <= 0 {
		msg.Time = time.Now().UnixNano()
	}

	if this.config.Sync {
		// 同步模式下 直接写消息
		return this.write(msg)
	} else {
		//异步模式写入管道
		this.waiter.Add(1)
		this.logger <- msg
		return nil
	}
}

func (this *Module) Flush() {
	if false == this.config.Sync {
		this.signal <- false
		this.waiter.Wait()
	} else {
		this.flush()
	}
}

// Logging 对外按日志级写日志的方法
func (this *Module) Logging(level Level, body string) error {
	msg := &Log{Time: time.Now().UnixNano(), Level: level, Body: body}
	return this.Write(msg)
}

// asyncLoop 异步循环
func (this *Module) eventLoop() {
	for {
		select {
		case log := <-this.logger:
			this.write(log)
			this.waiter.Done()
		case signal := <-this.signal:
			if signal {
				this.flush()
			} else {
				this.flush()
			}
		}
	}
}
