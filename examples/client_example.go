package main

import (
	"fmt"
	"strings"
	clientLib "utls_client/lib"

	utls "github.com/refraction-networking/utls"
)

func main() {
	// 示例 1: 使用默认客户端
	fmt.Println("=== 示例 1: 默认客户端 ===")
	demo1()

	// 示例 2: 自定义指纹
	fmt.Println("\n=== 示例 2: Chrome 133 指纹 ===")
	demo2()

	// 示例 3: Firefox 指纹
	fmt.Println("\n=== 示例 3: Firefox 120 指纹 ===")
	demo3()

	// 示例 4: 自定义头部
	fmt.Println("\n=== 示例 4: 自定义头部 ===")
	demo4()

	// 示例 5: POST 请求
	fmt.Println("\n=== 示例 5: POST 请求 ===")
	demo5()

	// 示例 6: IP 地址请求
	fmt.Println("\n=== 示例 6: IP 地址请求 ===")
	demo6()
}

// 示例 1: 使用默认客户端
func demo1() {
	// 创建默认客户端（Chrome 120 指纹）
	client := clientLib.DefaultClient()
	defer client.Close()

	// 发送 GET 请求
	resp, err := client.Get("https://httpbin.org/get", nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应长度: %d 字节\n", len(resp.Body))
	if len(resp.Body) < 200 {
		fmt.Printf("响应内容: %s\n", resp.Body)
	}
}

// 示例 2: Chrome 133 指纹
func demo2() {
	// 创建 Chrome 133 客户端
	client := clientLib.NewClient(&utls.HelloChrome_133, nil)
	defer client.Close()

	// 发送请求
	resp, err := client.Get("https://httpbin.org/get", map[string]string{
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	})

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("HTTP 版本: %s\n", resp.HTTPVersion)
}

// 示例 3: Firefox 指纹
func demo3() {
	// 创建 Firefox 120 客户端
	client := clientLib.NewClient(&utls.HelloFirefox_120, nil)
	defer client.Close()

	// 发送请求
	resp, err := client.Get("https://httpbin.org/get", map[string]string{
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	})

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应头数量: %d\n", len(resp.Headers))

	// 打印部分响应头
	count := 0
	for k, v := range resp.Headers {
		if count >= 3 {
			break
		}
		fmt.Printf("%s: %s\n", k, v)
		count++
	}
}

// 示例 4: 自定义头部
func demo4() {
	client := clientLib.DefaultClient()
	defer client.Close()

	// 自定义所有头部
	headers := map[string]string{
		"User-Agent":      "MyCustomClient/1.0",
		"Accept":          "application/json",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
		"Connection":      "keep-alive",
		"DNT":             "1",
		"Sec-Fetch-Dest":  "document",
		"Sec-Fetch-Mode":  "navigate",
		"Sec-Fetch-Site":  "none",
		"Cache-Control":   "max-age=0",
	}

	resp, err := client.Get("https://httpbin.org/headers", headers)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	if len(resp.Body) < 500 {
		fmt.Printf("响应内容: %s\n", resp.Body)
	} else {
		fmt.Printf("响应长度: %d 字节\n", len(resp.Body))
	}
}

// 示例 5: POST 请求
func demo5() {
	client := clientLib.DefaultClient()
	defer client.Close()

	// POST 数据
	body := strings.NewReader(`{"name": "test", "value": 123}`)

	resp, err := client.Post("https://httpbin.org/post", map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "uTLS-Client/1.0",
	}, body)

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	if len(resp.Body) < 500 {
		fmt.Printf("响应内容: %s\n", resp.Body)
	} else {
		fmt.Printf("响应长度: %d 字节\n", len(resp.Body))
	}
}

// 示例 6: IP 地址请求
func demo6() {
	client := clientLib.NewClient(&utls.HelloChrome_120, nil)
	defer client.Close()

	// 直接使用 IP 地址
	// 需要设置 Host 头来访问特定域名
	resp, err := client.Do("GET", "https://1.1.1.1/", &clientLib.RequestConfig{
		Method: "GET",
		Host:   "cloudflare-dns.com",
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
	})

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
}

// 高级示例：使用指纹库
func advancedExample() {
	// 导入 fingerprint 包
	// import "utls_client/fingerprint"

	// 从指纹库获取随机指纹
	// lib := fingerprint.NewFingerprintLibrary()
	// profile := lib.GetRandomProfile()
	//
	// 创建客户端
	// client := clientLib.NewClient(&profile.HelloID, nil)
	// defer client.Close()
	//
	// 发送请求
	// resp, err := client.Get("https://example.com", map[string]string{
	// 	"User-Agent": profile.UserAgent,
	// })
}

// 流式读取示例
func streamingExample() {
	client := clientLib.DefaultClient()
	defer client.Close()

	// 创建一个读取器
	reader := strings.NewReader("streaming data")

	resp, err := client.Post("https://httpbin.org/post", map[string]string{
		"Content-Type": "text/plain",
	}, reader)

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
}

// 自定义配置示例
func configExample() {
	// 创建配置
	config := &clientLib.Config{
		Timeout:            15 * 1000000000, // 15 秒
		InsecureSkipVerify: true,            // 跳过证书验证
		ServerName:         "example.com",   // 自定义 SNI
	}

	// 创建客户端
	client := clientLib.NewClient(&utls.HelloChrome_133, config)
	defer client.Close()

	// 发送请求
	resp, err := client.Get("https://httpbin.org/get", nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
}

// 处理文件上传示例
func uploadExample() {
	client := clientLib.DefaultClient()
	defer client.Close()

	// 模拟文件内容
	fileContent := "This is a test file content"
	reader := strings.NewReader(fileContent)

	resp, err := client.Post("https://httpbin.org/post", map[string]string{
		"Content-Type": "application/octet-stream",
	}, reader)

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("上传完成，状态码: %d\n", resp.StatusCode)
}

// 错误处理示例
func errorHandlingExample() {
	client := clientLib.DefaultClient()
	defer client.Close()

	// 尝试访问不存在的网站
	resp, err := client.Get("https://nonexistent-website-12345.com/", nil)

	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
		return
	}

	fmt.Printf("意外成功: %d\n", resp.StatusCode)
}

// 并发请求示例（注意：当前实现不支持并发，需要为每个goroutine创建独立的客户端）
func concurrentExample() {
	// 使用 channel 控制并发
	const numRequests = 5
	results := make(chan string, numRequests)

	// 启动多个 goroutine
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			// 每个 goroutine 创建独立的客户端
			client := clientLib.DefaultClient()
			defer client.Close()

			resp, err := client.Get("https://httpbin.org/get", nil)
			if err != nil {
				results <- fmt.Sprintf("请求 %d 失败: %v", id, err)
				return
			}

			results <- fmt.Sprintf("请求 %d 成功: %d", id, resp.StatusCode)
		}(i)
	}

	// 收集结果
	for i := 0; i < numRequests; i++ {
		fmt.Println(<-results)
	}
}

// 响应处理示例
func responseHandlingExample() {
	client := clientLib.DefaultClient()
	defer client.Close()

	resp, err := client.Get("https://httpbin.org/json", nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 检查状态码
	if resp.StatusCode != 200 {
		fmt.Printf("非200状态码: %d %s\n", resp.StatusCode, resp.Status)
		return
	}

	// 处理响应头
	for key, value := range resp.Headers {
		fmt.Printf("头部: %s = %s\n", key, value)
	}

	// 处理响应体
	fmt.Printf("响应体长度: %d 字节\n", len(resp.Body))

	// 检查 Content-Type
	if contentType, ok := resp.Headers["Content-Type"]; ok {
		fmt.Printf("Content-Type: %s\n", contentType)
	}
}

// 大文件下载示例
func downloadLargeFileExample() {
	client := clientLib.DefaultClient()
	defer client.Close()

	// 请求一个较大的资源
	resp, err := client.Get("https://httpbin.org/bytes/102400", nil)
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return
	}

	fmt.Printf("下载成功: %d 字节\n", len(resp.Body))
	fmt.Printf("状态码: %d\n", resp.StatusCode)

	// 保存文件（示例）
	// os.WriteFile("downloaded_file.bin", resp.Body, 0644)
}

// 模拟浏览器行为示例
func browserSimulationExample() {
	client := clientLib.NewClient(&utls.HelloChrome_133, nil)
	defer client.Close()

	// 模拟完整的浏览器请求头
	browserHeaders := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"Accept-Encoding":           "gzip, deflate, br",
		"Accept-Language":           "en-US,en;q=0.9",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"DNT":                       "1",
		"Sec-Ch-Ua":                 `"Chromium";v="133", "Not(A:Brand";v="8", "Google Chrome";v="133"`,
		"Sec-Ch-Ua-Mobile":          "?0",
		"Sec-Ch-Ua-Platform":        `"Windows"`,
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	}

	resp, err := client.Get("https://httpbin.org/headers", browserHeaders)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("浏览器模拟请求成功: %d\n", resp.StatusCode)
	if len(resp.Body) < 1000 {
		fmt.Printf("响应内容: %s\n", resp.Body)
	}
}
