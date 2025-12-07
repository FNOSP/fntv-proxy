package store

import (
	"sync"

	"fntv-proxy/model"
)

var (
	proxyInfo     model.ProxyInfo
	proxyInfoLock sync.RWMutex
)

// SetProxyInfo updates proxy info
func SetProxyInfo(url, cookie string) {
	proxyInfoLock.Lock()
	defer proxyInfoLock.Unlock()
	proxyInfo.URL = url
	proxyInfo.Cookie = cookie
}

// GetProxyInfo gets proxy info
func GetProxyInfo() model.ProxyInfo {
	proxyInfoLock.RLock()
	defer proxyInfoLock.RUnlock()
	return proxyInfo
}
