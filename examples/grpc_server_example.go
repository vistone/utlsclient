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
	"time"

	"google.golang.org/grpc"

	"utls_client/ippool"
	pb "utls_client/proto/ippool"
	"utls_client/server"
)

func main() {
	// 命令行参数
	var (
		port      = flag.String("port", "50051", "gRPC 服务监听端口")
		baseURL   = flag.String("base-url", "http://tile0.zeromaps.cn:9005", "IP 池 API 基础地址")
		dataDir   = flag.String("data-dir", "./ippool_data", "本地数据存储目录")
		autoSync  = flag.Bool("auto-sync", true, "是否启用自动同步")
		syncInt   = flag.Duration("sync-interval", 5*time.Minute, "自动同步间隔")
	)
	flag.Parse()

	fmt.Println("=== IP 池 gRPC 服务器 ===")
	fmt.Printf("监听端口: %s\n", *port)
	fmt.Printf("数据目录: %s\n", *dataDir)
	fmt.Printf("自动同步: %v\n", *autoSync)
	if *autoSync {
		fmt.Printf("同步间隔: %v\n", *syncInt)
	}
	fmt.Println()

	// 创建 IP 池库
	library := ippool.NewIPPoolLibrary(*baseURL, *dataDir)
	defer library.Close()

	// 启动自动同步
	if *autoSync {
		if err := library.StartAutoSync(*syncInt); err != nil {
			log.Printf("启动自动同步失败: %v", err)
		} else {
			log.Println("✅ 自动同步已启动")
		}
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()
	ippoolServer := server.NewIPPoolServer(library)
	pb.RegisterIPPoolServiceServer(grpcServer, ippoolServer)

	// 启动服务器
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	log.Printf("🚀 gRPC 服务器启动在端口 %s", *port)
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

