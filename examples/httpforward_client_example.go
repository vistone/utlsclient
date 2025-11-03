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

	pb "utls_client/proto/httpforward"
)

func main() {
	// 解析命令行参数
	serverAddr := flag.String("server", "localhost:50051", "gRPC 服务器地址")
	clientIP := flag.String("client-ip", "192.168.1.100", "客户端 IP 地址（首次握手需要）")
	hostname := flag.String("hostname", "www.example.com", "目标主机名")
	path := flag.String("path", "/", "请求路径")
	flag.Parse()

	// 连接 gRPC 服务器
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接服务器失败: %v", err)
	}
	defer conn.Close()

	client := pb.NewHTTPForwardServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. 首次握手，获取客户端编码
	fmt.Printf("1. 握手 - 发送客户端 IP: %s\n", *clientIP)
	handshakeReq := &pb.HandshakeRequest{
		ClientIp: *clientIP,
	}
	handshakeResp, err := client.Handshake(ctx, handshakeReq)
	if err != nil {
		log.Fatalf("握手失败: %v", err)
	}
	clientCode := handshakeResp.ClientCode
	fmt.Printf("   ✅ 获得客户端编码: %d\n\n", clientCode)

	// 2. 首次请求 - 使用原始主机名
	fmt.Printf("2. 首次请求 - 使用原始主机名: %s\n", *hostname)
	req1 := &pb.ForwardRequestRequest{
		ClientId: &pb.ForwardRequestRequest_ClientCode{ClientCode: clientCode},
		Hostname: &pb.ForwardRequestRequest_HostnameRaw{HostnameRaw: *hostname},
		Path:     *path,
	}

	resp1, err := client.ForwardRequest(ctx, req1)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	fmt.Printf("   Status Code: %d\n", resp1.StatusCode)
	fmt.Printf("   客户端编码: %d\n", resp1.ClientCode)
	fmt.Printf("   主机名编码: %d\n", resp1.HostnameCode)
	fmt.Printf("   响应体长度: %d 字节\n\n", len(resp1.Body))

	hostnameCode := resp1.HostnameCode

	// 3. 后续请求 - 使用编码（节省流量）
	fmt.Printf("3. 后续请求 - 使用编码（客户端编码=%d, 主机名编码=%d）\n", clientCode, hostnameCode)
	req2 := &pb.ForwardRequestRequest{
		ClientId: &pb.ForwardRequestRequest_ClientCode{ClientCode: clientCode},
		Hostname: &pb.ForwardRequestRequest_HostnameCode{HostnameCode: hostnameCode},
		Path:     *path,
	}

	resp2, err := client.ForwardRequest(ctx, req2)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	fmt.Printf("   Status Code: %d\n", resp2.StatusCode)
	fmt.Printf("   响应体长度: %d 字节\n\n", len(resp2.Body))

	// 4. 测试不同主机名
	testHostname := "www.google.com"
	fmt.Printf("4. 测试新主机名: %s\n", testHostname)
	req3 := &pb.ForwardRequestRequest{
		ClientId: &pb.ForwardRequestRequest_ClientCode{ClientCode: clientCode},
		Hostname: &pb.ForwardRequestRequest_HostnameRaw{HostnameRaw: testHostname},
		Path:     "/",
	}

	resp3, err := client.ForwardRequest(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	fmt.Printf("   Status Code: %d\n", resp3.StatusCode)
	fmt.Printf("   新主机名编码: %d\n", resp3.HostnameCode)
	fmt.Printf("   响应体长度: %d 字节\n\n", len(resp3.Body))

	fmt.Println("✅ 测试完成！")
	fmt.Println("\n💡 流量节省说明：")
	fmt.Println("   - 首次握手: 传输客户端 IP")
	fmt.Println("   - 后续请求: 只需传输编码（1,2,3,4...）")
	fmt.Println("   - 相同主机名: 使用编码，无需重复传输主机名字符串")
	fmt.Println("   - 非 200 状态码: body 为空，节省流量")
}
