package log

import (
	"strings"

	. "github.com/chefsgo/base"
)

func (this *Module) Register(name string, value Any, override bool) {
	switch obj := value.(type) {
	case Driver:
		this.Driver(name, obj, override)
	}
}
func (this *Module) Configure(value Any) {
	if cfg, ok := value.(Config); ok {
		this.config = cfg
		return
	}

	var global Map
	if cfg, ok := value.(Map); ok {
		global = cfg
	} else {
		return
	}

	var config Map
	if vvv, ok := global["log"].(Map); ok {
		config = vvv
	}

	//设置驱动
	if driver, ok := config["driver"].(string); ok {
		this.config.Driver = driver
	}
	//设置级别
	if level, ok := config["level"].(string); ok {
		for l, s := range levels {
			if strings.ToUpper(level) == s {
				this.config.Level = l
			}
		}
	}
	//是否json
	if json, ok := config["json"].(bool); ok {
		this.config.Json = json
	}
	//设置是否同步
	if sync, ok := config["sync"].(bool); ok {
		this.config.Sync = sync
	}
	// 设置输出格式
	if format, ok := config["format"].(string); ok {
		this.config.Format = format
	}

	// 设置缓存池大小
	if pool, ok := config["pool"].(int64); ok && pool > 0 {
		this.config.Pool = int(pool)
	}
	if pool, ok := config["pool"].(int); ok && pool > 0 {
		this.config.Pool = pool
	}

	if setting, ok := config["setting"].(Map); ok {
		this.config.Setting = setting
	}
}
func (this *Module) Initialize() {
	if this.initialized {
		return
	}

	this.initialized = true
}
func (this *Module) Connect() {
	if this.connected {
		return
	}

	driver, ok := this.drivers[this.config.Driver]
	if ok == false {
		panic("Invalid log driver: " + this.config.Driver)
	}

	// 建立连接
	connect, err := driver.Connect(this.config)
	if err != nil {
		panic("Failed to connect to log: " + err.Error())
	}

	// 打开连接
	err = connect.Open()
	if err != nil {
		panic("Failed to open log connect: " + err.Error())
	}

	// 保存连接，设置管道大小
	this.connect = connect
	this.logger = make(chan *Log, 100)
	this.signal = make(chan bool, 1)

	// 如果非同步模式，就开始异步循环
	if false == this.config.Sync {
		go this.eventLoop()
	}

	this.connected = true
}
func (this *Module) Launch() {
	if this.launched {
		return
	}

	this.launched = true
}
func (this *Module) Terminate() {
	if this.connect != nil {
		this.Flush()
		this.connect.Close()
	}

	this.launched = false
	this.connected = false
	this.initialized = false
}
