//go:build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "utls_client/proto/rocktreeTasks"
)

func main() {
	// 解析命令行参数
	serverAddr := flag.String("server", "localhost:50053", "gRPC 服务器地址")
	clientID := flag.String("client-id", "test-client-001", "客户端 ID")
	taskType := flag.String("type", "BULK_METADATA", "任务类型 (BULK_METADATA 或 NODE_DATA)")
	tilekey := flag.String("tilekey", "t:0:0:0", "瓦片键")
	epoch := flag.Int("epoch", 1, "纪元")
	imageryEpoch := flag.Int("imagery-epoch", 0, "图像纪元（可选，设为0则不使用）")
	flag.Parse()

	fmt.Println("=== RockTree 任务客户端示例 ===")
	fmt.Printf("服务器: %s\n", *serverAddr)
	fmt.Println()

	// 连接 gRPC 服务器
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接服务器失败: %v", err)
	}
	defer conn.Close()

	client := pb.NewRockTreeTaskServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 解析任务类型
	var taskTypeEnum pb.Type
	switch *taskType {
	case "BULK_METADATA", "bulk_metadata", "0":
		taskTypeEnum = pb.Type_BULK_METADATA
	case "NODE_DATA", "node_data", "1":
		taskTypeEnum = pb.Type_NODE_DATA
	default:
		log.Fatalf("无效的任务类型: %s (支持: BULK_METADATA, NODE_DATA)", *taskType)
	}

	// 创建任务请求
	req := &pb.TaskRequest{
		ClientId: *clientID,
		Type:     taskTypeEnum,
		Tilekey:  *tilekey,
		Epoch:    int32(*epoch),
	}

	// 如果指定了图像纪元，则添加到请求中
	if *imageryEpoch > 0 {
		req.ImageryEpoch = int32(*imageryEpoch)
	}

	fmt.Printf("发送任务请求:\n")
	fmt.Printf("  客户端 ID: %s\n", req.ClientId)
	fmt.Printf("  任务类型: %v\n", req.Type)
	fmt.Printf("  瓦片键: %s\n", req.Tilekey)
	fmt.Printf("  纪元: %d\n", req.Epoch)
	if req.ImageryEpoch > 0 {
		fmt.Printf("  图像纪元: %d\n", req.ImageryEpoch)
	}
	fmt.Println()

	// 发送请求
	resp, err := client.ProcessTask(ctx, req)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	// 显示结果
	fmt.Println("=== 响应结果 ===")
	fmt.Printf("客户端 ID: %s\n", resp.ClientId)
	fmt.Printf("任务类型: %v\n", resp.Type)
	fmt.Printf("瓦片键: %s\n", resp.Tilekey)
	fmt.Printf("纪元: %d\n", resp.Epoch)
	if resp.ImageryEpoch > 0 {
		fmt.Printf("图像纪元: %d\n", resp.ImageryEpoch)
	}
	fmt.Printf("状态码: %d\n", resp.StatusCode)

	// 显示响应体（限制长度）
	bodyLen := len(resp.Body)
	fmt.Printf("\n响应体长度: %d 字节\n", bodyLen)
	if bodyLen > 0 {
		previewLen := 500
		if bodyLen < previewLen {
			previewLen = bodyLen
		}
		fmt.Printf("响应体预览（前 %d 字节）:\n", previewLen)
		fmt.Println(string(resp.Body[:previewLen]))
		if bodyLen > previewLen {
			fmt.Printf("... (还有 %d 字节)\n", bodyLen-previewLen)
		}
	} else {
		fmt.Println("响应体为空（状态码非 200）")
	}

	fmt.Println("\n✅ 请求完成")
}


