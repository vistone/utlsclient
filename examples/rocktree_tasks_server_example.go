//go:build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	rocktreeServer "utls_client/server/rocktreeTasks"
	pb "utls_client/proto/rocktreeTasks"
)

func main() {
	// 命令行参数
	var (
		port = flag.String("port", "50053", "gRPC 服务监听端口")
	)
	flag.Parse()

	fmt.Println("=== RockTree 任务 gRPC 服务器 ===")
	fmt.Printf("监听端口: %s\n", *port)
	fmt.Println()

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()
	taskServer := rocktreeServer.NewRockTreeTaskServer()
	defer taskServer.Close()

	pb.RegisterRockTreeTaskServiceServer(grpcServer, taskServer)

	// 启动服务器
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	log.Printf("🚀 RockTree 任务 gRPC 服务器启动在端口 %s", *port)
	log.Println("按 Ctrl+C 停止服务器")

	// 在 goroutine 中启动服务器
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n正在关闭服务器...")
	grpcServer.GracefulStop()
	fmt.Println("✅ 服务器已关闭")
}


