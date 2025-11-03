package main

import (
	"fmt"
	"time"
	"utls_client/fingerprint"
	clientLib "utls_client/lib"
)

func main() {
	fmt.Println("测试 uTLS 客户端 + 指纹库...")
	fmt.Println("使用指纹库自动匹配指纹和 User-Agent")
	fmt.Println()

	// 从指纹库随机获取指纹
	lib := fingerprint.NewFingerprintLibrary()
	profile := lib.GetRandomProfile()

	fmt.Printf("✅ 使用随机指纹: %s\n", profile.Name)
	fmt.Printf("   User-Agent: %s\n\n", profile.UserAgent)

	// 创建客户端，使用指纹库的指纹配置，并设置代理
	config := &clientLib.Config{
		Timeout: 60 * time.Second,
		Proxy:   "http://127.0.0.1:20172", // HTTP 代理
	}
	client := clientLib.NewClient(&profile.HelloID, config)
	defer client.Close()

	// 使用指纹库匹配的 User-Agent 和浏览器头部
	headers := map[string]string{
		"Accept":     "application/json, text/javascript, */*; q=0.01",
		"User-Agent": profile.UserAgent, // 使用匹配的 User-Agent
		"Origin":     "https://earth.google.com",
		"Referer":    "https://earth.google.com/",
	}

	// 测试地址
	target := "https://kh.google.com/rt/earth/PlanetoidMetadata"

	fmt.Printf("请求: %s\n", target)
	resp, err := client.Get(target, headers)

	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	fmt.Printf("✅ 状态码: %d\n", resp.StatusCode)
	fmt.Printf("HTTP 版本: %s\n", resp.HTTPVersion)

	if len(resp.Body) > 0 {
		fmt.Printf("响应长度: %d 字节\n", len(resp.Body))
		if len(resp.Body) < 500 {
			fmt.Printf("响应内容: %s\n", string(resp.Body))
		} else {
			fmt.Printf("响应内容(前500字节): %s\n", string(resp.Body[:500]))
		}
	}

	// 打印重要响应头
	fmt.Println("\n响应头:")
	for key, value := range resp.Headers {
		if len(value) > 100 {
			value = value[:100] + "..."
		}
		fmt.Printf("  %s: %s\n", key, value)
	}
}
