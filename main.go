package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"gopkg.in/ini.v1"
)

// 全局变量存储代理信息
type ProxyInfo struct {
	URL    string
	Cookie string
}

var (
	proxyInfo     ProxyInfo
	proxyInfoLock sync.RWMutex
)

type Config struct {
	Port int `ini:"port"`
}

var stdoutLogger *log.Logger

func main() {
	// 读取配置文件
	config, err := loadConfig("config.ini")
	if err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}

	// 启动HTTP服务器
	port := config.Port
	if port == 0 {
		port = 1999 // 默认端口
	}

	mux := http.NewServeMux()

	// 设置路由
	mux.HandleFunc("/proxy/info", handleProxyInfo)
	mux.HandleFunc("/proxyGet", handleProxyGet)
	mux.HandleFunc("/", handleVLCRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	stdoutLogger = log.New(os.Stdout, "", log.LstdFlags)
	stdoutLogger.Printf("VLC代理服务启动，监听端口: %d", port)
	log.Fatal(server.ListenAndServe())
}

// 加载配置文件
func loadConfig(filename string) (*Config, error) {
	config := &Config{}

	// 检查配置文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 创建默认配置文件
		err = createDefaultConfig(filename)
		if err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}
		stdoutLogger.Printf("已创建默认配置文件: %s", filename)
	}

	// 加载配置文件
	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}

	port, err := cfg.Section("server").Key("port").Int()
	if err != nil {
		log.Printf("配置文件中端口格式不正确，使用默认端口1999")
		port = 1999
	}

	config.Port = port
	return config, nil
}

// 创建默认配置文件
func createDefaultConfig(filename string) error {
	cfg := ini.Empty()
	cfg.Section("server").Key("port").SetValue("1999")
	return cfg.SaveTo(filename)
}

// 处理代理信息设置请求
func handleProxyInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体中的JSON数据
	var requestData struct {
		URL    string `json:"url"`
		Cookie string `json:"cookie"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "解析请求体失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.URL == "" {
		http.Error(w, "url参数不能为空", http.StatusBadRequest)
		return
	}

	// 存储代理信息到全局变量
	proxyInfoLock.Lock()
	proxyInfo.URL = requestData.URL
	proxyInfo.Cookie = requestData.Cookie
	proxyInfoLock.Unlock()

	stdoutLogger.Printf("更新代理信息: URL=%s, Cookie=%s", requestData.URL, requestData.Cookie)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"code": 0, "status": "success", "message": "代理信息已更新", "data": true}`)
}

// 获取代理信息请求
func handleProxyGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持GET请求", http.StatusMethodNotAllowed)
		return
	}

	proxyInfoLock.RLock()
	url := proxyInfo.URL
	cookie := proxyInfo.Cookie
	proxyInfoLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"url": "%s", "cookie": "%s"}`, url, cookie)
}

// 处理VLC请求
func handleVLCRequest(w http.ResponseWriter, r *http.Request) {
	// 读取代理信息
	proxyInfoLock.RLock()
	targetURL := proxyInfo.URL
	cookie := proxyInfo.Cookie
	proxyInfoLock.RUnlock()

	if targetURL == "" {
		http.Error(w, "未设置代理信息，请先调用/proxyInfo接口", http.StatusBadRequest)
		return
	}

	requestPath := r.URL.Path
	targetFullURL, err := buildTargetURL(targetURL, requestPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("构建目标URL失败: %v", err), http.StatusBadRequest)
		return
	}

	proxyReq, err := http.NewRequest(r.Method, targetFullURL.String(), r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("创建代理请求失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 复制原始请求头
	for name, values := range r.Header {
		if strings.EqualFold(name, "Host") {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 添加Cookie
	if cookie != "" {
		proxyReq.Header.Set("Cookie", cookie)
	}

	if strings.Contains(requestPath, "/v/api/v1/media/range/") {
		proxyReq.Header.Set("Range", "bytes=0-")
	}

	// 打印请求信息
	stdoutLogger.Printf("代理请求: %s %s", r.Method, targetFullURL.String())

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("发送请求失败: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("复制响应体失败: %v", err)
	}
}

// 构建目标URL
func buildTargetURL(baseURL, path string) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	reference, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	targetURL := u.ResolveReference(reference)
	return targetURL, nil
}
