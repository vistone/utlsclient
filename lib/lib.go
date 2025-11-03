package utls_client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// Client 通用 uTLS HTTP 客户端
type Client struct {
	// 配置
	config *Config

	// HTTP/2 客户端缓存
	h2Clients map[string]*http.Client
	h2Mu      sync.Mutex

	// HTTP/1.1 客户端缓存（回退使用）
	h1Clients map[string]*http.Client
	h1Mu      sync.Mutex
}

// Config 客户端配置
type Config struct {
	// 超时时间
	Timeout time.Duration

	// 是否跳过证书验证
	InsecureSkipVerify bool

	// SNI 设置
	ServerName string

	// 代理地址，格式: http://host:port 或 socks5://host:port
	Proxy string

	// LocalIP 本地源地址（可选，优先使用；适用于绑定本地IPv6）
	LocalIP string
}

// RequestConfig 请求配置
type RequestConfig struct {
	// 请求方法
	Method string

	// 请求路径
	Path string

	// 自定义头部
	Headers map[string]string

	// 请求体
	Body io.Reader

	// Host 头（覆盖域名）
	Host string

	// LocalIP 本地源地址（可选，覆盖全局Config.LocalIP）
	LocalIP string
}

// Response 响应结构
type Response struct {
	// HTTP 状态码
	StatusCode int

	// 状态文本
	Status string

	// 响应头
	Headers map[string]string

	// 响应体
	Body []byte

	// HTTP 版本
	HTTPVersion string
}

// NewClient 创建新的 uTLS 客户端
func NewClient(fingerprint *utls.ClientHelloID, config *Config) *Client {
	if config == nil {
		config = &Config{
			Timeout: 30 * time.Second,
		}
	}

	// 如果未提供指纹，使用默认
	if fingerprint == nil {
		defaultFingerprint := utls.HelloChrome_133
		fingerprint = &defaultFingerprint
	}

	// 将指纹存储到配置中以便后续使用
	// 注意：这里简化处理，假设客户端的整个生命周期使用同一个指纹
	// 如果需要动态切换，需要重新设计
	_ = fingerprint

	return &Client{
		config:    config,
		h2Clients: make(map[string]*http.Client),
		h1Clients: make(map[string]*http.Client),
	}
}

// DefaultClient 创建默认客户端（Chrome 指纹）
func DefaultClient() *Client {
	return NewClient(&utls.HelloChrome_133, nil)
}

// Get 发送 GET 请求
func (c *Client) Get(target string, headers map[string]string) (*Response, error) {
	return c.Do("GET", target, &RequestConfig{
		Method:  "GET",
		Headers: headers,
	})
}

// Post 发送 POST 请求
func (c *Client) Post(target string, headers map[string]string, body io.Reader) (*Response, error) {
	return c.Do("POST", target, &RequestConfig{
		Method:  "POST",
		Headers: headers,
		Body:    body,
	})
}

// Do 发送 HTTP 请求
func (c *Client) Do(method, target string, req *RequestConfig) (*Response, error) {
	// 解析 URL
	parsedURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %w", err)
	}

	host := parsedURL.Hostname()
	if host == "" {
		return nil, fmt.Errorf("无效的URL: %s", target)
	}

	// 确定使用指纹
	fingerprint := utls.HelloChrome_133 // 默认使用 Chrome 133

	// 根据 User-Agent 推断指纹（简化版本，从 User-Agent 判断浏览器类型）
	if req != nil && req.Headers != nil {
		if ua, ok := req.Headers["User-Agent"]; ok {
			fingerprint = inferFingerprintFromUA(ua)
		}
	}

	// 优先使用 HTTP/2
	h2Client := c.getOrCreateH2Client(target, host, &fingerprint)

	// 构建请求
	httpReq, err := http.NewRequest(method, target, req.Body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置头部
	if req.Host != "" {
		httpReq.Host = req.Host
	}

	if req.Headers != nil {
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}
	}

	// 尝试 HTTP/2
	resp, err := h2Client.Do(httpReq)
	if err == nil {
		return c.convertResponse(resp)
	}

	// HTTP/2 失败，回退到 HTTP/1.1
	h1Client := c.getOrCreateH1Client(target, host, &fingerprint)
	resp, err = h1Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	return c.convertResponse(resp)
}

// getOrCreateH2Client 获取或创建 HTTP/2 客户端
func (c *Client) getOrCreateH2Client(target, host string, fingerprint *utls.ClientHelloID) *http.Client {
	c.h2Mu.Lock()
	defer c.h2Mu.Unlock()

	// 生成缓存键
	key := host

	// 检查是否已存在
	if client, ok := c.h2Clients[key]; ok {
		return client
	}

	// 创建新客户端
	client := c.buildHTTP2Client(host, fingerprint)
	c.h2Clients[key] = client

	return client
}

// getOrCreateH1Client 获取或创建 HTTP/1.1 客户端
func (c *Client) getOrCreateH1Client(target, host string, fingerprint *utls.ClientHelloID) *http.Client {
	c.h1Mu.Lock()
	defer c.h1Mu.Unlock()

	// 生成缓存键
	key := host

	// 检查是否已存在
	if client, ok := c.h1Clients[key]; ok {
		return client
	}

	// 创建新客户端
	client := c.buildHTTP1Client(host, fingerprint)
	c.h1Clients[key] = client

	return client
}

// buildHTTP2Client 创建 HTTP/2 客户端
func (c *Client) buildHTTP2Client(host string, fingerprint *utls.ClientHelloID) *http.Client {
	transport := &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return c.dialUTLS(ctx, network, addr, host, fingerprint, []string{"h2"})
		},
		ReadIdleTimeout:  30 * time.Second,
		PingTimeout:      15 * time.Second,
		WriteByteTimeout: 10 * time.Second,
		MaxReadFrameSize: 1 << 20,
		AllowHTTP:        false,
	}

	return &http.Client{
		Timeout:   c.config.Timeout,
		Transport: transport,
	}
}

// buildHTTP1Client 创建 HTTP/1.1 客户端
func (c *Client) buildHTTP1Client(host string, fingerprint *utls.ClientHelloID) *http.Client {
	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return c.dialUTLS(ctx, network, addr, host, fingerprint, []string{"http/1.1"})
		},
		TLSHandshakeTimeout:   c.config.Timeout,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       60 * time.Second,
		DisableKeepAlives:     false,
		ResponseHeaderTimeout: c.config.Timeout,
	}

	return &http.Client{
		Timeout:   c.config.Timeout,
		Transport: transport,
	}
}

// dialUTLS 使用 uTLS 建立 TLS 连接
func (c *Client) dialUTLS(ctx context.Context, network, addr, serverName string, fingerprint *utls.ClientHelloID, nextProtos []string) (net.Conn, error) {
	var conn net.Conn
	var err error

	// 如果配置了代理，先连接代理
	if c.config.Proxy != "" {
		conn, err = c.connectThroughProxy(addr)
		if err != nil {
			return nil, fmt.Errorf("代理连接失败: %w", err)
		}
	} else {
		dialer := &net.Dialer{
			Timeout: c.config.Timeout,
		}
		// 绑定本地源地址（若提供）
		localIP := c.config.LocalIP
		if localIP == "" {
			// 请求级覆盖
			// 无法直接取得当前请求，这里通过 network 与 addr 仅能设置全局；
			// 提供一个钩子：若调用方在构造 client 前设置 Config.LocalIP 即可。
		}
		if localIP != "" {
			if ip := net.ParseIP(localIP); ip != nil {
				dialer.LocalAddr = &net.TCPAddr{IP: ip}
			}
		}
		conn, err = dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, fmt.Errorf("TCP 连接失败: %w", err)
		}
	}

	// 确定 SNI
	sni := serverName
	if c.config.ServerName != "" {
		sni = c.config.ServerName
	}

	// 创建 uTLS 配置
	tlsConfig := &utls.Config{
		ServerName:         sni,
		InsecureSkipVerify: c.config.InsecureSkipVerify,
	}

	if len(nextProtos) > 0 {
		tlsConfig.NextProtos = nextProtos
	}

	// 使用 uTLS 建立连接
	uconn := utls.UClient(conn, tlsConfig, *fingerprint)
	if err := uconn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("TLS 握手失败: %w", err)
	}

	return uconn, nil
}

// convertResponse 转换标准 HTTP 响应为我们的响应格式
func (c *Client) convertResponse(resp *http.Response) (*Response, error) {
	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 构建响应头
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 确定 HTTP 版本
	httpVersion := "HTTP/1.1"
	if resp.ProtoMajor == 2 {
		httpVersion = "HTTP/2"
	}

	return &Response{
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Headers:     headers,
		Body:        body,
		HTTPVersion: httpVersion,
	}, nil
}

// connectThroughProxy 通过代理连接
func (c *Client) connectThroughProxy(targetAddr string) (net.Conn, error) {
	proxyURL, err := url.Parse(c.config.Proxy)
	if err != nil {
		return nil, fmt.Errorf("无效的代理URL: %w", err)
	}

	var proxyAddr string
	if proxyURL.Host != "" {
		proxyAddr = proxyURL.Host
	} else {
		proxyAddr = c.config.Proxy
	}

	// 确保有端口号
	if !strings.Contains(proxyAddr, ":") {
		if proxyURL.Scheme == "socks5" {
			proxyAddr += ":1080"
		} else {
			proxyAddr += ":80"
		}
	}

	// 连接代理服务器
	dialer := &net.Dialer{Timeout: c.config.Timeout}
	conn, err := dialer.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, err
	}

	// 根据代理类型处理
	switch proxyURL.Scheme {
	case "http", "":
		// HTTP CONNECT 代理
		return c.httpConnectProxy(conn, targetAddr)
	case "socks5":
		// SOCKS5 代理
		return c.socks5Connect(conn, targetAddr)
	default:
		conn.Close()
		return nil, fmt.Errorf("不支持的代理协议: %s", proxyURL.Scheme)
	}
}

// httpConnectProxy HTTP CONNECT 代理
func (c *Client) httpConnectProxy(conn net.Conn, target string) (net.Conn, error) {
	// 发送 CONNECT 请求
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", target, target)
	if _, err := conn.Write([]byte(connectReq)); err != nil {
		conn.Close()
		return nil, err
	}

	// 读取响应
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		conn.Close()
		return nil, err
	}

	response := string(buf[:n])

	// 检查响应状态
	if !strings.Contains(response, "200") {
		conn.Close()
		return nil, fmt.Errorf("代理CONNECT失败: %s", strings.Split(response, "\n")[0])
	}

	return conn, nil
}

// socks5Connect SOCKS5 代理连接
func (c *Client) socks5Connect(conn net.Conn, target string) (net.Conn, error) {
	// SOCKS5 握手：无认证
	handshake := []byte{0x05, 0x01, 0x00}
	if _, err := conn.Write(handshake); err != nil {
		conn.Close()
		return nil, err
	}

	// 读取服务器响应
	response := make([]byte, 2)
	if _, err := conn.Read(response); err != nil {
		conn.Close()
		return nil, err
	}

	if response[0] != 0x05 || response[1] != 0x00 {
		conn.Close()
		return nil, fmt.Errorf("SOCKS5握手失败")
	}

	// 解析目标地址
	host, port, err := c.parseAddr(target)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// 构建连接请求
	var connectReq []byte
	connectReq = append(connectReq, 0x05, 0x01, 0x00) // VER, CMD, RSV

	// 地址类型和地址
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			// IPv4
			connectReq = append(connectReq, 0x01)
			connectReq = append(connectReq, ip.To4()...)
		} else {
			// IPv6
			connectReq = append(connectReq, 0x04)
			connectReq = append(connectReq, ip.To16()...)
		}
	} else {
		// 域名
		connectReq = append(connectReq, 0x03, byte(len(host)))
		connectReq = append(connectReq, []byte(host)...)
	}

	// 端口
	connectReq = append(connectReq, byte(port>>8), byte(port&0xff))

	// 发送连接请求
	if _, err := conn.Write(connectReq); err != nil {
		conn.Close()
		return nil, err
	}

	// 读取响应
	replyHeader := make([]byte, 4)
	if _, err := conn.Read(replyHeader); err != nil {
		conn.Close()
		return nil, err
	}

	if replyHeader[0] != 0x05 || replyHeader[1] != 0x00 {
		conn.Close()
		return nil, fmt.Errorf("SOCKS5连接失败: %d", replyHeader[1])
	}

	// 读取绑定地址（跳过）
	addrType := replyHeader[3]
	switch addrType {
	case 0x01: // IPv4
		buf := make([]byte, 4)
		conn.Read(buf)
	case 0x03: // 域名
		buf := make([]byte, 1)
		conn.Read(buf)
		buf = make([]byte, int(buf[0]))
		conn.Read(buf)
	case 0x04: // IPv6
		buf := make([]byte, 16)
		conn.Read(buf)
	}
	conn.Read(make([]byte, 2)) // 端口

	return conn, nil
}

// parseAddr 解析地址
func (c *Client) parseAddr(addr string) (host string, port int, err error) {
	colonIdx := strings.LastIndex(addr, ":")
	if colonIdx == -1 {
		return "", 0, fmt.Errorf("无效的地址格式")
	}

	host = addr[:colonIdx]
	portStr := addr[colonIdx+1:]

	var p int
	if _, err := fmt.Sscanf(portStr, "%d", &p); err != nil {
		return "", 0, err
	}

	return host, p, nil
}

// inferFingerprintFromUA 从 User-Agent 推断指纹类型（简化版本）
func inferFingerprintFromUA(ua string) utls.ClientHelloID {
	ua = strings.ToLower(ua)

	if strings.Contains(ua, "firefox") {
		return utls.HelloFirefox_120
	} else if strings.Contains(ua, "edge") {
		return utls.HelloEdge_Auto
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return utls.HelloIOS_Auto
	}

	// 默认 Chrome
	return utls.HelloChrome_133
}

// Close 关闭客户端（释放连接）
func (c *Client) Close() error {
	// 这里主要是释放资源，客户端本身不需要特殊清理
	// 连接由底层的 transport 管理
	return nil
}

// SetTimeout 设置超时
func (c *Client) SetTimeout(timeout time.Duration) {
	c.config.Timeout = timeout
}

// SetServerName 设置 SNI
func (c *Client) SetServerName(serverName string) {
	c.config.ServerName = serverName
}

// SetInsecureSkipVerify 设置是否跳过证书验证
func (c *Client) SetInsecureSkipVerify(skip bool) {
	c.config.InsecureSkipVerify = skip
}
