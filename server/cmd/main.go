package main

import (
	"flag"
	"fmt"
	"log"

	"hunjie/internal/config"
	"hunjie/internal/db"
	"hunjie/internal/pkg/logger"
	"hunjie/internal/pkg/sensitive"
	"hunjie/internal/pkg/sms"
	"hunjie/internal/pkg/wechat"
	"hunjie/internal/router"
	"hunjie/internal/service"
)

func main() {
	configPath := flag.String("c", "config.dev.yaml", "配置文件路径")
	flag.Parse()

	if err := config.Load(*configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	logger.Init(config.C.Server.Mode)
	defer logger.Sync()

	if err := db.Init(); err != nil {
		logger.Fatalf("初始化数据库失败: %v", err)
	}
	logger.Infof("数据库已连接: %s", config.C.MySQL.Database)

	sensitive.Init("configs/sensitive_words.txt")
	wechat.Init()
	sms.Init()
	service.InitWS()
	service.StartMaintenance()

	r := router.Setup()
	addr := fmt.Sprintf(":%d", config.C.Server.Port)
	logger.Infof("服务启动于 http://localhost%s", addr)

	if err := r.Run(addr); err != nil {
		logger.Fatalf("服务启动失败: %v", err)
	}
}
