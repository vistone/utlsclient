package utls_client

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	utls "github.com/refraction-networking/utls"
)

// ConnPoolManager 针对远端IPv6地址的长连接预热与复用
type ConnPoolManager struct {
	mu       sync.RWMutex
	clients  map[string]*Client   // remoteIP -> client
	lastOK   map[string]time.Time // 最近成功时间
	helloID  utls.ClientHelloID
	baseConf *Config
}

func NewConnPoolManager(hello utls.ClientHelloID, base *Config) *ConnPoolManager {
	if base == nil {
		base = &Config{Timeout: 30 * time.Second}
	}
	return &ConnPoolManager{
		clients:  make(map[string]*Client),
		lastOK:   make(map[string]time.Time),
		helloID:  hello,
		baseConf: base,
	}
}

// WarmUp 针对一组远端IPv6预热HTTP/2连接（SNI/Host 请由请求时设置）
func (m *ConnPoolManager) WarmUp(remoteIPs []string) {
	var wg sync.WaitGroup
	for _, ip := range remoteIPs {
		ip := ip
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 构建客户端（可选绑定本地源IP：通过 m.baseConf.LocalIP）
			cli := NewClient(&m.helloID, m.baseConf)
			// 进行一次轻量请求，促进建立HTTP/2会话（由上层传入时再真正请求）
			// 这里不立即发网络请求，交由上层第一次 Do 时建立；仅缓存 client
			m.mu.Lock()
			m.clients[ip] = cli
			m.mu.Unlock()
		}()
	}
	wg.Wait()
}

func (m *ConnPoolManager) Get(remoteIP string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[remoteIP]
	return c, ok
}

// MarkResult 记录结果，便于上层实现黑/白名单等策略
func (m *ConnPoolManager) MarkResult(remoteIP string, status int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err == nil && status == http.StatusOK {
		m.lastOK[remoteIP] = time.Now()
	}
}

// BindLocalIPv6 为连接池设置统一的本地源IPv6（可选）
func (m *ConnPoolManager) BindLocalIPv6(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("无效的IPv6: %s", ip)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.baseConf == nil {
		m.baseConf = &Config{}
	}
	m.baseConf.LocalIP = ip
	return nil
}
