package log

import (
	. "github.com/chefsgo/base"
)

//定义列表
func Levels() map[Level]string {
	return levels
}

// //语法糖
func Debug(args ...Any) {
	module.Debug(args...)
}
func Trace(args ...Any) {
	module.Trace(args...)
}
func Info(args ...Any) {
	module.Info(args...)
}
func Notice(args ...Any) {
	module.Notice(args...)
}
func Warning(args ...Any) {
	module.Warning(args...)
}
func Panic(args ...Any) {
	module.Panic(args...)
}
func Fatal(args ...Any) {
	module.Fatal(args...)
}
