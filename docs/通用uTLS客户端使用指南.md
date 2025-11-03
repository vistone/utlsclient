# 通用 uTLS 客户端使用指南

## 📋 目录
1. [概述](#概述)
2. [快速开始](#快速开始)
3. [核心功能](#核心功能)
4. [详细示例](#详细示例)
5. [高级功能](#高级功能)
6. [常见问题](#常见问题)

---

## 概述

通用 uTLS 客户端是一个完全灵活的 HTTP/HTTPS 请求库，支持：
- ✅ **完整自定义头部**：支持所有 HTTP 头部
- ✅ **域名和 IP 请求**：直接支持 IP 地址访问
- ✅ **HTTP 和 HTTPS**：自动协议识别
- ✅ **自定义 TLS 指纹**：支持所有 uTLS 指纹
- ✅ **流式传输**：支持大数据流传输
- ✅ **分块编码**：完整支持 HTTP 分块传输
- ✅ **零依赖**：只依赖 uTLS 库

---

## 快速开始

### 安装

```bash
go get github.com/refraction-networking/utls@v1.8.1
```

### 基础示例

```go
package main

import (
    "fmt"
    "utls_client"
)

func main() {
    // 创建默认客户端
    client := utls_client.DefaultClient()
    defer client.Close()
    
    // 发送 GET 请求
    resp, err := client.Get("https://httpbin.org/get", nil)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("状态码: %d\n", resp.StatusCode)
    fmt.Printf("响应: %s\n", resp.Body)
}
```

---

## 核心功能

### 1. 创建客户端

#### 默认客户端
```go
// 使用 Chrome 120 指纹
client := utls_client.DefaultClient()
```

#### 自定义指纹
```go
import utls "github.com/refraction-networking/utls"

// Chrome 133
client := utls_client.NewClient(&utls.HelloChrome_133, nil)

// Firefox 120
client := utls_client.NewClient(&utls.HelloFirefox_120, nil)

// Edge 106
client := utls_client.NewClient(&utls.HelloEdge_106, nil)

// Safari
client := utls_client.NewClient(&utls.HelloSafari_Auto, nil)

// 随机指纹
client := utls_client.NewClient(&utls.HelloRandomized, nil)
```

#### 自定义配置
```go
config := &utls_client.Config{
    Timeout:            30 * time.Second,  // 超时时间
    InsecureSkipVerify: false,             // 是否跳过证书验证
    ServerName:         "example.com",     // SNI 设置
}

client := utls_client.NewClient(&utls.HelloChrome_120, config)
```

### 2. 发送请求

#### GET 请求
```go
// 基本 GET
resp, err := client.Get("https://example.com", nil)

// 带头部
headers := map[string]string{
    "User-Agent": "MyClient/1.0",
    "Accept":     "application/json",
}
resp, err := client.Get("https://example.com", headers)
```

#### POST 请求
```go
import "strings"

// 字符串数据
body := strings.NewReader("data=value")
resp, err := client.Post("https://example.com", map[string]string{
    "Content-Type": "application/x-www-form-urlencoded",
}, body)

// JSON 数据
jsonBody := strings.NewReader(`{"key": "value"}`)
resp, err := client.Post("https://example.com", map[string]string{
    "Content-Type": "application/json",
}, jsonBody)
```

#### 通用请求
```go
resp, err := client.Do("https://example.com", &utls_client.RequestConfig{
    Method: "PUT",
    Path:   "/api/resource",
    Headers: map[string]string{
        "Authorization": "Bearer token",
        "Content-Type":  "application/json",
    },
    Body: strings.NewReader(`{"data": "value"}`),
})
```

### 3. 处理响应

```go
resp, err := client.Get("https://example.com", nil)
if err != nil {
    panic(err)
}

// 状态码
fmt.Printf("状态码: %d %s\n", resp.StatusCode, resp.Status)

// HTTP 版本
fmt.Printf("HTTP 版本: %s\n", resp.HTTPVersion)

// 响应头
for key, value := range resp.Headers {
    fmt.Printf("%s: %s\n", key, value)
}

// 响应体
fmt.Printf("响应: %s\n", resp.Body)

// 二进制数据
os.WriteFile("output.bin", resp.Body, 0644)
```

### 4. IP 地址访问

```go
// 直接使用 IP
resp, err := client.Get("https://1.1.1.1/", map[string]string{
    "Host": "cloudflare-dns.com",  // 设置 Host 头
})

// 使用请求配置
resp, err := client.Do("https://1.2.3.4/", &utls_client.RequestConfig{
    Method: "GET",
    Host:   "example.com",  // Host 头
    Headers: map[string]string{
        "User-Agent": "Mozilla/5.0",
    },
})
```

---

## 详细示例

### 示例 1: 模拟真实浏览器

```go
package main

import (
    "fmt"
    "utls_client"
    utls "github.com/refraction-networking/utls"
)

func main() {
    // Chrome 133 指纹
    client := utls_client.NewClient(&utls.HelloChrome_133, nil)
    defer client.Close()
    
    // 完整的浏览器头部
    browserHeaders := map[string]string{
        "Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language":           "en-US,en;q=0.9",
        "Accept-Encoding":           "gzip, deflate, br",
        "Connection":                "keep-alive",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
        "Sec-Ch-Ua":                 `"Chromium";v="133", "Not(A:Brand";v="8", "Google Chrome";v="133"`,
        "Sec-Ch-Ua-Mobile":          "?0",
        "Sec-Fetch-Dest":            "document",
        "Sec-Fetch-Mode":            "navigate",
        "Sec-Fetch-Site":            "none",
        "Cache-Control":             "max-age=0",
    }
    
    resp, err := client.Get("https://www.example.com", browserHeaders)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("成功: %d\n", resp.StatusCode)
}
```

### 示例 2: 结合指纹库

```go
package main

import (
    "fmt"
    "utls_client"
    "utls_client/fingerprint"
)

func main() {
    // 从指纹库获取随机指纹
    lib := fingerprint.NewFingerprintLibrary()
    profile := lib.GetRandomProfile()
    
    fmt.Printf("使用指纹: %s\n", profile.Name)
    
    // 创建客户端
    client := utls_client.NewClient(&profile.HelloID, nil)
    defer client.Close()
    
    // 使用匹配的 User-Agent
    resp, err := client.Get("https://httpbin.org/get", map[string]string{
        "User-Agent": profile.UserAgent,
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("状态码: %d\n", resp.StatusCode)
}
```

### 示例 3: API 客户端

```go
package main

import (
    "encoding/json"
    "fmt"
    "strings"
    "utls_client"
    utls "github.com/refraction-networking/utls"
)

type APIResponse struct {
    Status string      `json:"status"`
    Data   interface{} `json:"data"`
}

func main() {
    client := utls_client.NewClient(&utls.HelloChrome_133, nil)
    defer client.Close()
    
    // 发送 JSON 请求
    jsonData := map[string]interface{}{
        "username": "user",
        "password": "pass",
    }
    
    jsonBytes, _ := json.Marshal(jsonData)
    body := strings.NewReader(string(jsonBytes))
    
    resp, err := client.Post("https://api.example.com/login", map[string]string{
        "Content-Type":  "application/json",
        "Accept":        "application/json",
        "Authorization": "Bearer token",
    }, body)
    
    if err != nil {
        panic(err)
    }
    
    // 解析 JSON 响应
    var apiResp APIResponse
    if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
        panic(err)
    }
    
    fmt.Printf("API 状态: %s\n", apiResp.Status)
}
```

### 示例 4: 文件上传

```go
package main

import (
    "fmt"
    "os"
    "utls_client"
    utls "github.com/refraction-networking/utls"
)

func main() {
    client := utls_client.NewClient(&utls.HelloChrome_133, nil)
    defer client.Close()
    
    // 读取文件
    fileContent, err := os.ReadFile("test.txt")
    if err != nil {
        panic(err)
    }
    
    // 上传
    body := strings.NewReader(string(fileContent))
    resp, err := client.Post("https://upload.example.com/api/files", map[string]string{
        "Content-Type": "application/octet-stream",
    }, body)
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("上传成功: %d\n", resp.StatusCode)
}
```

### 示例 5: 下载文件

```go
package main

import (
    "fmt"
    "os"
    "utls_client"
    utls "github.com/refraction-networking/utls"
)

func main() {
    client := utls_client.NewClient(&utls.HelloFirefox_120, nil)
    defer client.Close()
    
    // 下载文件
    resp, err := client.Get("https://example.com/file.zip", nil)
    if err != nil {
        panic(err)
    }
    
    // 保存到本地
    if err := os.WriteFile("downloaded.zip", resp.Body, 0644); err != nil {
        panic(err)
    }
    
    fmt.Printf("下载完成: %d 字节\n", len(resp.Body))
}
```

---

## 高级功能

### 1. 动态配置

```go
client := utls_client.DefaultClient()

// 修改超时
client.SetTimeout(60 * time.Second)

// 修改 SNI
client.SetServerName("custom.example.com")

// 跳过证书验证
client.SetInsecureSkipVerify(true)

// 切换指纹
client.SetFingerprint(utls.HelloChrome_133)
```

### 2. 连接管理

```go
// 手动连接
client.Connect("https://example.com")

// 发送请求（使用已存在的连接）
resp, err := client.Get("https://example.com/path1", nil)
resp, err := client.Get("https://example.com/path2", nil)

// 重新连接
client.Reconnect("https://example.com")

// 关闭连接
client.Close()
```

### 3. 自定义路径和 Host

```go
resp, err := client.Do("https://example.com", &utls_client.RequestConfig{
    Method: "GET",
    Path:   "/custom/path?param=value",
    Host:   "different-domain.com",  // Host 头
    Headers: map[string]string{
        "User-Agent": "Custom-Client",
    },
})
```

### 4. 处理分块传输

```go
// 客户端自动处理 HTTP 分块传输
resp, err := client.Get("https://example.com/stream", nil)

// 检查是否使用分块
if resp.Headers["Transfer-Encoding"] == "chunked" {
    fmt.Println("使用分块传输")
}
```

---

## API 参考

### Client 结构

```go
type Client struct {
    config      *Config
    fingerprint *utls.ClientHelloID
    conn        net.Conn
}
```

### Config 结构

```go
type Config struct {
    Timeout             time.Duration  // 超时时间
    InsecureSkipVerify  bool           // 跳过证书验证
    ServerName          string         // SNI
    TLSConfig           *tls.Config    // 自定义 TLS 配置
}
```

### RequestConfig 结构

```go
type RequestConfig struct {
    Method      string            // HTTP 方法
    Path        string            // 请求路径
    HTTPVersion string            // HTTP 版本 (1.1 或 2)
    Headers     map[string]string // 自定义头部
    Body        io.Reader         // 请求体
    Host        string            // Host 头
}
```

### Response 结构

```go
type Response struct {
    StatusCode  int               // 状态码
    Status      string            // 状态文本
    Headers     map[string]string // 响应头
    Body        []byte            // 响应体
    HTTPVersion string            // HTTP 版本
}
```

### 主要方法

```go
// 创建客户端
func NewClient(fingerprint *utls.ClientHelloID, config *Config) *Client
func DefaultClient() *Client

// 连接管理
func (c *Client) Connect(target string) error
func (c *Client) Close() error
func (c *Client) Reconnect(target string) error

// 发送请求
func (c *Client) Do(target string, req *RequestConfig) (*Response, error)
func (c *Client) Get(target string, headers map[string]string) (*Response, error)
func (c *Client) Post(target string, headers map[string]string, body io.Reader) (*Response, error)

// 配置修改
func (c *Client) SetTimeout(timeout time.Duration)
func (c *Client) SetServerName(serverName string)
func (c *Client) SetInsecureSkipVerify(skip bool)
func (c *Client) SetFingerprint(fingerprint utls.ClientHelloID)
```

---

## 常见问题

### Q1: 如何处理 HTTPS 证书错误？

```go
config := &utls_client.Config{
    InsecureSkipVerify: true,  // 跳过证书验证
}
client := utls_client.NewClient(&utls.HelloChrome_120, config)
```

或者：

```go
client := utls_client.DefaultClient()
client.SetInsecureSkipVerify(true)
```

### Q2: 如何使用 IP 地址访问特定域名？

```go
// 方法 1: 设置 Host 头
resp, err := client.Get("https://1.2.3.4/", map[string]string{
    "Host": "example.com",
})

// 方法 2: 使用请求配置
resp, err := client.Do("https://1.2.3.4/", &utls_client.RequestConfig{
    Host: "example.com",
    Headers: map[string]string{
        "User-Agent": "Mozilla/5.0",
    },
})
```

### Q3: 如何并发请求？

**注意**：当前实现不支持并发，需要为每个 goroutine 创建独立的客户端。

```go
const numWorkers = 10

for i := 0; i < numWorkers; i++ {
    go func(id int) {
        // 每个 goroutine 创建独立客户端
        client := utls_client.NewClient(&utls.HelloChrome_120, nil)
        defer client.Close()
        
        resp, err := client.Get("https://example.com", nil)
        if err != nil {
            fmt.Printf("Worker %d 失败: %v\n", id, err)
            return
        }
        
        fmt.Printf("Worker %d 成功: %d\n", id, resp.StatusCode)
    }(i)
}
```

### Q4: 如何处理大数据传输？

```go
// 使用流式读取器
client := utls_client.DefaultClient()
defer client.Close()

resp, err := client.Post("https://upload.example.com", map[string]string{
    "Content-Type": "application/octet-stream",
}, largeReader)

if err != nil {
    panic(err)
}

// 响应体会自动缓存
fmt.Printf("传输完成: %d 字节\n", len(resp.Body))
```

### Q5: 如何自定义 SNI？

```go
config := &utls_client.Config{
    ServerName: "custom.example.com",
}
client := utls_client.NewClient(&utls.HelloChrome_133, config)
```

或者：

```go
client := utls_client.DefaultClient()
client.SetServerName("custom.example.com")
```

### Q6: 如何切换指纹？

```go
// 创建时指定
client := utls_client.NewClient(&utls.HelloFirefox_120, nil)

// 动态切换
client.SetFingerprint(utls.HelloChrome_133)
```

### Q7: 为什么建议使用指纹库？

使用指纹库可以确保 TLS 指纹和 HTTP 头部的一致性：

```go
import "utls_client/fingerprint"

lib := fingerprint.NewFingerprintLibrary()
profile := lib.GetRandomProfile()

client := utls_client.NewClient(&profile.HelloID, nil)

// 使用匹配的 User-Agent
resp, err := client.Get("https://example.com", map[string]string{
    "User-Agent": profile.UserAgent,
})
```

---

## 最佳实践

### 1. 指纹和头部一致性

```go
// ❌ 错误：混用不同的指纹和 User-Agent
client := utls_client.NewClient(&utls.HelloFirefox_120, nil)
resp, err := client.Get("https://example.com", map[string]string{
    "User-Agent": "Chrome User-Agent",  // 不匹配！
})

// ✅ 正确：使用指纹库确保一致性
lib := fingerprint.NewFingerprintLibrary()
profile := lib.GetRandomProfile()
client := utls_client.NewClient(&profile.HelloID, nil)
resp, err := client.Get("https://example.com", map[string]string{
    "User-Agent": profile.UserAgent,
})
```

### 2. 超时设置

```go
config := &utls_client.Config{
    Timeout: 30 * time.Second,  // 设置合理的超时
}
client := utls_client.NewClient(&utls.HelloChrome_120, config)
```

### 3. 错误处理

```go
resp, err := client.Get("https://example.com", nil)
if err != nil {
    // 检查错误类型
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        fmt.Println("超时错误")
    } else {
        fmt.Printf("其他错误: %v\n", err)
    }
    return
}
```

### 4. 资源清理

```go
client := utls_client.DefaultClient()
defer client.Close()  // 确保关闭连接

resp, err := client.Get("https://example.com", nil)
```

### 5. 日志记录

```go
resp, err := client.Get("https://example.com", nil)
if err != nil {
    log.Printf("请求失败: %v\n", err)
    return
}

log.Printf("请求成功: %d\n", resp.StatusCode)
```

---

## 限制说明

1. **不支持并发**：每个 goroutine 需要独立的客户端实例
2. **HTTP/2 支持有限**：主要针对 HTTP/1.1 优化
3. **连接池**：当前使用单连接，不支持自动连接池
4. **Cookie 管理**：需要手动管理 Cookie

---

## 完整示例

参考 `examples/client_example.go` 文件查看更多详细示例。

---

**最后更新**：2025-01-10  
**版本**：1.0.0  
**说明**：通用的 uTLS HTTP/HTTPS 客户端库


