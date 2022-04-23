package log

type (

	// LogDriver log驱动
	Driver interface {
		// 连接到驱动
		Connect(config Config) (Connect, error)
	}
	// LogConnect 日志连接
	Connect interface {
		// Open 打开连接
		Open() error

		// Close 关闭结束
		Close() error

		// Write 写入日志
		Write(*Log) error

		// Flush 冲马桶
		Flush()
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
