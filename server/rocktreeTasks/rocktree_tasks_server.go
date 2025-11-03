package rocktreeTasks

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	clientLib "utls_client/lib"
	pb "utls_client/proto/rocktreeTasks"
)

// RockTreeTaskServer gRPC 服务器实现
type RockTreeTaskServer struct {
	pb.UnimplementedRockTreeTaskServiceServer
	client *clientLib.Client
}

// NewRockTreeTaskServer 创建新的 RockTree 任务服务器
func NewRockTreeTaskServer() *RockTreeTaskServer {
	// 创建 uTLS 客户端（使用默认 Chrome 指纹）
	config := &clientLib.Config{
		Timeout: 30 * time.Second, // 30秒超时
	}
	client := clientLib.NewClient(nil, config)

	return &RockTreeTaskServer{
		client: client,
	}
}

// ProcessTask 处理任务请求
func (s *RockTreeTaskServer) ProcessTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	// 参数验证
	if req.ClientId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "client_id 不能为空")
	}
	if req.Tilekey == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tilekey 不能为空")
	}

	// 根据任务类型构建不同的 URL
	var url string
	switch req.Type {
	case pb.Type_BULK_METADATA:
		// 批量元数据请求 URL
		if req.ImageryEpoch > 0 {
			url = fmt.Sprintf("https://tile.googleapis.com/tile/v1/bulkmetadata?tilekey=%s&epoch=%d&imagery_epoch=%d",
				req.Tilekey, req.Epoch, req.ImageryEpoch)
		} else {
			url = fmt.Sprintf("https://tile.googleapis.com/tile/v1/bulkmetadata?tilekey=%s&epoch=%d",
				req.Tilekey, req.Epoch)
		}
	case pb.Type_NODE_DATA:
		// 节点数据请求 URL
		if req.ImageryEpoch > 0 {
			url = fmt.Sprintf("https://tile.googleapis.com/tile/v1/nodedata?tilekey=%s&epoch=%d&imagery_epoch=%d",
				req.Tilekey, req.Epoch, req.ImageryEpoch)
		} else {
			url = fmt.Sprintf("https://tile.googleapis.com/tile/v1/nodedata?tilekey=%s&epoch=%d",
				req.Tilekey, req.Epoch)
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "无效的任务类型: %v", req.Type)
	}

	// 设置默认请求头
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "*/*",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
		"Connection":      "keep-alive",
		"Origin":          "https://www.google.com",
		"Referer":         "https://www.google.com/",
	}

	// 使用 uTLS 客户端发送 GET 请求
	resp, err := s.client.Get(url, headers)
	if err != nil {
		return &pb.TaskResponse{
			ClientId:     req.ClientId,
			Type:         req.Type,
			Tilekey:      req.Tilekey,
			Epoch:        req.Epoch,
			ImageryEpoch: req.ImageryEpoch,
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

	return &pb.TaskResponse{
		ClientId:     req.ClientId,
		Type:         req.Type,
		Tilekey:      req.Tilekey,
		Epoch:        req.Epoch,
		ImageryEpoch: req.ImageryEpoch,
		Body:         responseBody,
		StatusCode:   int32(resp.StatusCode),
	}, nil
}

// Close 关闭服务器
func (s *RockTreeTaskServer) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}


