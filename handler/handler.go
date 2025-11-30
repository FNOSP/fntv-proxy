package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"fn-tv-vlc-convert-url/logger"
	"fn-tv-vlc-convert-url/model"
	"fn-tv-vlc-convert-url/store"
)

// HandleProxyInfo 处理代理信息设置请求
func HandleProxyInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体中的JSON数据
	var requestData model.ProxyInfoRequest

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "解析请求体失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.URL == "" {
		http.Error(w, "url参数不能为空", http.StatusBadRequest)
		return
	}

	// 存储代理信息到全局变量
	store.SetProxyInfo(requestData.URL, requestData.Cookie)

	logger.StdoutLogger.Printf("更新代理信息: URL=%s, Cookie=%s", requestData.URL, requestData.Cookie)

	w.Header().Set("Content-Type", "application/json")
	response := model.CommonResponse{
		Code:    0,
		Status:  "success",
		Message: "代理信息已更新",
		Data:    true,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.StdoutLogger.Printf("序列化响应内容失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// HandleProxyGet 获取代理信息请求
func HandleProxyGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持GET请求", http.StatusMethodNotAllowed)
		return
	}

	info := store.GetProxyInfo()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		logger.StdoutLogger.Printf("序列化响应内容失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// HandleVLCRequest 处理VLC请求
func HandleVLCRequest(w http.ResponseWriter, r *http.Request) {
	// 读取代理信息
	info := store.GetProxyInfo()
	targetURL := info.URL
	cookie := info.Cookie

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
	logger.StdoutLogger.Printf("代理请求: %s %s", r.Method, targetFullURL.String())

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("发送请求失败: %v", err), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.StdoutLogger.Printf("关闭响应体失败: %v", err)
		}
	}()

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

// buildTargetURL 构建目标URL
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
