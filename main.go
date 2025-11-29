package main

import (
	"fmt"
	"log"
	"net/http"

	"fn-tv-vlc-convert-url/config"
	"fn-tv-vlc-convert-url/handler"
	"fn-tv-vlc-convert-url/logger"
)

func main() {
	// 初始化日志
	logger.Init()

	// 读取配置文件
	conf, err := config.LoadConfig("config.ini")
	if err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}

	// 启动HTTP服务器
	port := conf.Port
	if port == 0 {
		port = 1999 // 默认端口
	}

	mux := http.NewServeMux()

	// 设置路由
	mux.HandleFunc("/proxy/info", handler.HandleProxyInfo)
	mux.HandleFunc("/proxyGet", handler.HandleProxyGet)
	mux.HandleFunc("/", handler.HandleVLCRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	logger.StdoutLogger.Printf("VLC代理服务启动，监听端口: %d", port)
	log.Fatal(server.ListenAndServe())
}
