package httpforward

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	clientLib "utls_client/lib"
	pb "utls_client/proto/httpforward"
)

// HTTPForwardServer gRPC 服务器实现
type HTTPForwardServer struct {
	pb.UnimplementedHTTPForwardServiceServer
	client *clientLib.Client

	// 客户端 IP -> 编码映射
	clientIPToCode map[string]int32
	clientCodeToIP map[int32]string
	nextClientCode int32
	clientMapMu    sync.RWMutex

	// 主机名 -> 编码映射（全局共享）
	hostnameToCode     map[string]int32
	hostnameCodeToName map[int32]string
	nextHostnameCode   int32
	hostnameMapMu      sync.RWMutex
}

// NewHTTPForwardServer 创建新的 HTTP 转发服务器
func NewHTTPForwardServer() *HTTPForwardServer {
	// 创建 uTLS 客户端（使用默认 Chrome 指纹）
	config := &clientLib.Config{
		Timeout: 30 * time.Second, // 30秒超时
	}
	client := clientLib.NewClient(nil, config)

	return &HTTPForwardServer{
		client:             client,
		clientIPToCode:     make(map[string]int32),
		clientCodeToIP:     make(map[int32]string),
		nextClientCode:     1,
		hostnameToCode:     make(map[string]int32),
		hostnameCodeToName: make(map[int32]string),
		nextHostnameCode:   1,
	}
}

// Handshake 客户端握手（首次连接，分配编码）
func (s *HTTPForwardServer) Handshake(ctx context.Context, req *pb.HandshakeRequest) (*pb.HandshakeResponse, error) {
	if req.ClientIp == "" {
		return nil, status.Errorf(codes.InvalidArgument, "client_ip 不能为空")
	}

	s.clientMapMu.Lock()
	defer s.clientMapMu.Unlock()

	// 检查是否已存在
	if code, exists := s.clientIPToCode[req.ClientIp]; exists {
		return &pb.HandshakeResponse{ClientCode: code}, nil
	}

	// 分配新编码
	clientCode := s.nextClientCode
	s.nextClientCode++

	s.clientIPToCode[req.ClientIp] = clientCode
	s.clientCodeToIP[clientCode] = req.ClientIp

	return &pb.HandshakeResponse{ClientCode: clientCode}, nil
}

// ForwardRequest 转发 HTTP 请求（使用 uTLS 客户端）
func (s *HTTPForwardServer) ForwardRequest(ctx context.Context, req *pb.ForwardRequestRequest) (*pb.ForwardRequestResponse, error) {
	// 解析客户端标识（IP 或编码）
	var clientCode int32
	if clientIP := req.GetClientIp(); clientIP != "" {
		// 首次使用 IP，需要先分配编码
		handshakeResp, err := s.Handshake(ctx, &pb.HandshakeRequest{ClientIp: clientIP})
		if err != nil {
			return nil, err
		}
		clientCode = handshakeResp.ClientCode
	} else if clientCodeVal := req.GetClientCode(); clientCodeVal != 0 {
		// 使用编码
		s.clientMapMu.RLock()
		_, exists := s.clientCodeToIP[clientCodeVal]
		s.clientMapMu.RUnlock()
		if !exists {
			return nil, status.Errorf(codes.InvalidArgument, "无效的客户端编码: %d", clientCodeVal)
		}
		clientCode = clientCodeVal
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "必须提供 client_ip 或 client_code")
	}

	// 解析主机名（原始值或编码）
	var hostname string
	var hostnameCode int32
	if hostnameRaw := req.GetHostnameRaw(); hostnameRaw != "" {
		// 首次使用原始主机名
		s.hostnameMapMu.Lock()
		if code, exists := s.hostnameToCode[hostnameRaw]; exists {
			hostnameCode = code
			hostname = hostnameRaw
		} else {
			// 分配新编码
			hostnameCode = s.nextHostnameCode
			s.nextHostnameCode++
			s.hostnameToCode[hostnameRaw] = hostnameCode
			s.hostnameCodeToName[hostnameCode] = hostnameRaw
			hostname = hostnameRaw
		}
		s.hostnameMapMu.Unlock()
	} else if hostnameCodeVal := req.GetHostnameCode(); hostnameCodeVal != 0 {
		// 使用编码
		s.hostnameMapMu.RLock()
		var exists bool
		hostname, exists = s.hostnameCodeToName[hostnameCodeVal]
		s.hostnameMapMu.RUnlock()
		if !exists {
			return nil, status.Errorf(codes.InvalidArgument, "无效的主机名编码: %d", hostnameCodeVal)
		}
		hostnameCode = hostnameCodeVal
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "必须提供 hostname_raw 或 hostname_code")
	}

	path := req.GetPath()
	if path == "" {
		path = "/"
	}

	// 构建完整的 URL
	url := fmt.Sprintf("https://%s%s", hostname, path)

	// 设置默认请求头
	headers := map[string]string{
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.9",
		"Accept-Encoding":           "gzip, deflate, br",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	}

	// 使用 uTLS 客户端发送 GET 请求
	resp, err := s.client.Get(url, headers)
	if err != nil {
		return &pb.ForwardRequestResponse{
			ClientCode:   clientCode,
			HostnameCode: hostnameCode,
			Path:         path,
			Body:         []byte{}, // 非 200 状态码返回空 body
			StatusCode:   500,
		}, nil // 返回错误但不返回 gRPC 错误，让客户端处理
	}

	// 只有状态码 200 时才返回 body，其他状态码返回空 body 以节省流量
	var responseBody []byte
	if resp.StatusCode == 200 {
		responseBody = resp.Body
	} else {
		responseBody = []byte{} // 非 200 状态码，返回空 body
	}

	return &pb.ForwardRequestResponse{
		ClientCode:   clientCode,
		HostnameCode: hostnameCode,
		Path:         path,
		Body:         responseBody,
		StatusCode:   int32(resp.StatusCode),
	}, nil
}

// Close 关闭服务器
func (s *HTTPForwardServer) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
