package logger

import (
	"go.uber.org/zap"
)

// L 全局日志。小项目直接用全局变量，不做依赖注入
var L *zap.SugaredLogger

func Init(mode string) {
	var l *zap.Logger
	var err error
	if mode == "release" {
		l, err = zap.NewProduction()
	} else {
		l, err = zap.NewDevelopment()
	}
	if err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	L = l.Sugar()
}

func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

func Info(args ...any)                    { L.Info(args...) }
func Infof(tpl string, args ...any)       { L.Infof(tpl, args...) }
func Warnf(tpl string, args ...any)       { L.Warnf(tpl, args...) }
func Errorf(tpl string, args ...any)      { L.Errorf(tpl, args...) }
func Debugf(tpl string, args ...any)      { L.Debugf(tpl, args...) }
func Fatalf(tpl string, args ...any)      { L.Fatalf(tpl, args...) }
