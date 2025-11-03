// +build integration

package server

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"utls_client/ippool"
	pb "utls_client/proto/ippool"
)

// TestIntegrationServer 集成测试：启动真实的 gRPC 服务器并测试
func TestIntegrationServer(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试（使用 -short 标志）")
	}

	// 检查环境变量，只有设置了 INTEGRATION_TEST=true 才运行
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("跳过集成测试（设置 INTEGRATION_TEST=true 来运行）")
	}

	dataDir := "./testdata/ippool_integration"
	os.MkdirAll(dataDir, 0755)

	// 创建 IP 池库
	library := ippool.NewIPPoolLibrary("", dataDir)
	defer library.Close()

	// 创建服务器
	server := NewIPPoolServer(library)

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":0") // 使用随机端口
	if err != nil {
		t.Fatalf("监听失败: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterIPPoolServiceServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("服务器启动失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端连接
	addr := lis.Addr().String()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	client := pb.NewIPPoolServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试 GetAllHosts
	t.Run("GetAllHosts", func(t *testing.T) {
		resp, err := client.GetAllHosts(ctx, &pb.GetAllHostsRequest{})
		if err != nil {
			t.Errorf("GetAllHosts 失败: %v", err)
			return
		}
		t.Logf("找到 %d 个主机", len(resp.Hosts))
	})

	// 测试 GetServiceStatus
	t.Run("GetServiceStatus", func(t *testing.T) {
		resp, err := client.GetServiceStatus(ctx, &pb.GetServiceStatusRequest{})
		if err != nil {
			t.Errorf("GetServiceStatus 失败: %v", err)
			return
		}
		t.Logf("服务状态: 离线模式=%v, 主机数=%d", resp.OfflineMode, resp.TotalHosts)
	})

	// 清理
	grpcServer.GracefulStop()
}

// TestConcurrentRequests 测试并发请求
func TestConcurrentRequests(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	// 并发请求数量
	concurrency := 10
	requests := 100

	// 使用 channel 来协调 goroutine
	errors := make(chan error, requests)
	
	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < requests/concurrency; j++ {
				req := &pb.GetAllHostsRequest{}
				_, err := server.GetAllHosts(ctx, req)
				errors <- err
			}
		}()
	}

	// 收集错误
	var errorCount int
	for i := 0; i < requests; i++ {
		if err := <-errors; err != nil {
			errorCount++
		}
	}

	t.Logf("并发测试完成: 总请求数=%d, 错误数=%d", requests, errorCount)
}

