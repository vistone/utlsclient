# uQUIC 全面使用指南

## 目录
1. [什么是 uQUIC](#什么是-uquic)
2. [安装与导入](#安装与导入)
3. [快速开始](#快速开始)
4. [预设浏览器指纹](#预设浏览器指纹)
5. [自定义 QUIC Spec](#自定义-quic-spec)
6. [HTTP/3 客户端示例](#http3-客户端示例)
7. [高级功能](#高级功能)
8. [注意事项与最佳实践](#注意事项与最佳实践)
9. [常见问题解答](#常见问题解答)
10. [参考资料](#参考资料)

---

## 什么是 uQUIC

uQUIC 是 [quic-go](https://github.com/quic-go/quic-go) 的一个分支，旨在提供对 QUIC Initial Packet（初始数据包）的低级访问，以实现指纹对抗和模拟目的。虽然握手仍然由 quic-go 执行，但 uQUIC 提供了自定义未加密 Initial Packet 的接口，该数据包可能泄露可被指纹识别的信息。

### 主要功能

- **Initial Packet 指纹对抗**：通过自定义未加密的 Initial Packet，减少被指纹识别的风险
- **基于 quic-go**：继承 quic-go 的稳定性和性能
- **预设浏览器指纹**：提供 Chrome、Firefox 等浏览器的预设指纹
- **HTTP/3 支持**：与 uTLS 配合，提供完整的 HTTP/3 客户端支持
- **TLS ClientHello 集成**：通过 uTLS 自定义 TLS 握手

### 应用场景

- HTTP/3 客户端开发
- 隐私保护工具开发
- 网络代理客户端开发
- 需要绕过 QUIC 指纹检测的应用
- 反审查工具开发

### ⚠️ 重要声明

**生产环境警告：** 该项目仍处于概念验证阶段，**不推荐在生产环境使用**。

**研究项目：** 本项目属于大型研究项目的一部分，研究内容是如何指纹识别 QUIC 客户端以及如何缓解这种指纹识别。我们的研究论文尚未发表，因此此仓库既未准备好用于生产，也未经过同行评议。

**风险提示：** 用于反审查目的的开发者请务必理解本库的风险和局限性。某些误用可能导致更容易被指纹识别。

---

## 安装与导入

### 版本要求

- **Go 版本：** 1.20+（最低要求）
- **最新 uQUIC 版本：** v0.0.6（2024-07-19 发布）

### 1. 安装 uQUIC

```bash
# 安装最新版本
go get -u github.com/refraction-networking/uquic

# 安装特定版本
go get github.com/refraction-networking/uquic@v0.0.6

# 使用 go.mod
go mod edit -require=github.com/refraction-networking/uquic@v0.0.6
go mod tidy
```

### 2. 安装依赖

uQUIC 与 uTLS 配合使用，因此需要同时安装：

```bash
go get -u github.com/refraction-networking/utls
```

### 3. 导入包

```go
import (
    "github.com/refraction-networking/uquic"
    "github.com/refraction-networking/uquic/http3"
    tls "github.com/refraction-networking/utls"
)
```

---

## 快速开始

### 最简单的示例：使用 HTTP/3

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "bytes"
    "io"

    tls "github.com/refraction-networking/utls"
    "github.com/refraction-networking/uquic"
    "github.com/refraction-networking/uquic/http3"
)

func main() {
    // 1. 创建标准的 HTTP/3 RoundTripper
    roundTripper := &http3.RoundTripper{
        TLSClientConfig: &tls.Config{},
        QuicConfig:      &uquic.Config{},
    }

    // 2. 获取预设的 QUIC Spec（使用 Firefox 116 指纹）
    quicSpec, err := uquic.QUICID2Spec(uquic.QUICFirefox_116)
    if err != nil {
        log.Fatal(err)
    }

    // 3. 转换为 uQUIC RoundTripper
    uRoundTripper := http3.GetURoundTripper(
        roundTripper,
        &quicSpec,
        nil,
    )
    defer uRoundTripper.Close()

    // 4. 创建 HTTP 客户端
    client := &http.Client{
        Transport: uRoundTripper,
    }

    // 5. 发送请求
    resp, err := client.Get("https://quic.tlsfingerprint.io/qfp/?beautify=true")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    // 6. 读取响应
    body := &bytes.Buffer{}
    _, err = io.Copy(body, resp.Body)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("状态码: %d\n", resp.StatusCode)
    fmt.Printf("响应体: %s\n", body.String())
}
```

---

## 预设浏览器指纹

uQUIC 提供了多个预设的浏览器指纹，可以在 `QUICID2Spec` 中使用。

### 支持的预设指纹

根据 GitHub 仓库信息，目前 uQUIC 支持以下预设指纹：

```go
// Chrome 系列
uquic.QUICChrome_115    // Chrome 115 QUIC fingerprint
// 注：更多 Chrome 版本指纹正在开发中

// Firefox 系列
uquic.QUICFirefox_116   // Firefox 116 QUIC fingerprint
// 注：更多 Firefox 版本指纹正在开发中

// 未来将支持
// - Apple Safari parrot
// - Microsoft Edge parrot
```

### 使用预设指纹

```go
// 使用 Firefox 116 指纹
quicSpec, err := uquic.QUICID2Spec(uquic.QUICFirefox_116)

// 使用 Chrome 115 指纹
quicSpec, err := uquic.QUICID2Spec(uquic.QUICChrome_115)
```

### 完整示例

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    tls "github.com/refraction-networking/utls"
    "github.com/refraction-networking/uquic"
    "github.com/refraction-networking/uquic/http3"
)

func main() {
    // 配置标准 RoundTripper
    roundTripper := &http3.RoundTripper{
        TLSClientConfig: &tls.Config{},
        QuicConfig:      &uquic.Config{},
    }

    // 选择指纹
    var quicSpec uquic.QUICSpec
    var err error
    
    // 选项1：使用 Firefox
    quicSpec, err = uquic.QUICID2Spec(uquic.QUICFirefox_116)
    if err != nil {
        log.Fatal(err)
    }

    // 选项2：使用 Chrome（注释掉上面的 Firefox，使用这个）
    // quicSpec, err = uquic.QUICID2Spec(uquic.QUICChrome_115)
    // if err != nil {
    //     log.Fatal(err)
    // }

    // 转换为 uQUIC RoundTripper
    uRoundTripper := http3.GetURoundTripper(
        roundTripper,
        &quicSpec,
        nil,
    )
    defer uRoundTripper.Close()

    // 创建 HTTP 客户端
    client := &http.Client{
        Transport: uRoundTripper,
    }

    // 发送请求
    resp, err := client.Get("https://example.com")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    fmt.Printf("状态码: %d\n", resp.StatusCode)
}
```

---

## 自定义 QUIC Spec

如果需要完全控制 QUIC Initial Packet，可以自定义 QUIC Spec。

### 查看示例

查看 `u_parrots.go` 文件中的示例，了解如何构建自定义 QUIC Spec：

```go
// 从源代码中学习
// 文件路径：uquic/u_parrots.go
```

### QUIC Spec 结构

QUIC Spec 定义了 uQUIC 建立 QUIC 连接的参数和策略，包括：

- **QUIC Header**：连接 ID、版本号等
- **QUIC Frames**：Crypto Frame、Padding Frame、Ping Frame 等
- **TLS ClientHello**：通过 uTLS 自定义
- **QUIC Transport Parameters**：流控、拥塞控制等参数

### 自定义示例

```go
package main

import (
    "github.com/refraction-networking/uquic"
    tls "github.com/refraction-networking/utls"
)

func createCustomQUICSpec() uquic.QUICSpec {
    // 这里只是示例结构，具体实现需要参考源代码
    spec := uquic.QUICSpec{
        // 设置各种参数
        // Header: ...
        // Frames: ...
        // ClientHelloID: ...
        // TransportParams: ...
    }
    
    return spec
}
```

---

## HTTP/3 客户端示例

### 示例 1：基础 HTTP/3 GET 请求

```go
package main

import (
    "fmt"
    "io"
    "log"
    "net/http"

    tls "github.com/refraction-networking/utls"
    "github.com/refraction-networking/uquic"
    "github.com/refraction-networking/uquic/http3"
)

func main() {
    roundTripper := &http3.RoundTripper{
        TLSClientConfig: &tls.Config{},
        QuicConfig:      &uquic.Config{},
    }

    quicSpec, err := uquic.QUICID2Spec(uquic.QUICFirefox_116)
    if err != nil {
        log.Fatal(err)
    }

    uRoundTripper := http3.GetURoundTripper(
        roundTripper,
        &quicSpec,
        nil,
    )
    defer uRoundTripper.Close()

    client := &http.Client{
        Transport: uRoundTripper,
    }

    resp, err := client.Get("https://example.com")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("状态码: %d\n", resp.StatusCode)
    fmt.Printf("响应长度: %d\n", len(body))
}
```

### 示例 2：带自定义头部的请求

```go
package main

import (
    "fmt"
    "io"
    "log"
    "net/http"

    tls "github.com/refraction-networking/utls"
    "github.com/refraction-networking/uquic"
    "github.com/refraction-networking/uquic/http3"
)

func main() {
    roundTripper := &http3.RoundTripper{
        TLSClientConfig: &tls.Config{},
        QuicConfig:      &uquic.Config{},
    }

    quicSpec, err := uquic.QUICID2Spec(uquic.QUICChrome_115)
    if err != nil {
        log.Fatal(err)
    }

    uRoundTripper := http3.GetURoundTripper(
        roundTripper,
        &quicSpec,
        nil,
    )
    defer uRoundTripper.Close()

    client := &http.Client{
        Transport: uRoundTripper,
    }

    // 创建请求
    req, err := http.NewRequest("GET", "https://api.example.com/data", nil)
    if err != nil {
        log.Fatal(err)
    }

    // 添加自定义头部
    req.Header.Set("User-Agent", "MyApp/1.0")
    req.Header.Set("Accept", "application/json")

    // 发送请求
    resp, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("响应: %s\n", string(body))
}
```

### 示例 3：POST 请求

```go
package main

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"

    tls "github.com/refraction-networking/utls"
    "github.com/refraction-networking/uquic"
    "github.com/refraction-networking/uquic/http3"
)

func main() {
    roundTripper := &http3.RoundTripper{
        TLSClientConfig: &tls.Config{},
        QuicConfig:      &uquic.Config{},
    }

    quicSpec, err := uquic.QUICID2Spec(uquic.QUICFirefox_116)
    if err != nil {
        log.Fatal(err)
    }

    uRoundTripper := http3.GetURoundTripper(
        roundTripper,
        &quicSpec,
        nil,
    )
    defer uRoundTripper.Close()

    client := &http.Client{
        Transport: uRoundTripper,
    }

    // 准备 POST 数据
    data := strings.NewReader(`{"key": "value"}`)

    // 创建 POST 请求
    req, err := http.NewRequest("POST", "https://api.example.com/submit", data)
    if err != nil {
        log.Fatal(err)
    }

    req.Header.Set("Content-Type", "application/json")

    // 发送请求
    resp, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("POST 响应: %s\n", string(body))
}
```

---

## 高级功能

### 1. 与 uTLS 配合使用

uQUIC 通过 uTLS 自定义 TLS ClientHello。可以结合 uTLS 的浏览器指纹：

```go
package main

import (
    "net/http"
    "github.com/refraction-networking/uquic"
    "github.com/refraction-networking/uquic/http3"
    tls "github.com/refraction-networking/utls"
)

func main() {
    // 配置 uTLS
    tlsConfig := &tls.Config{
        ServerName: "example.com",
    }

    roundTripper := &http3.RoundTripper{
        TLSClientConfig: tlsConfig,
        QuicConfig:      &uquic.Config{},
    }

    // 使用 Firefox QUIC + TLS 指纹
    quicSpec, _ := uquic.QUICID2Spec(uquic.QUICFirefox_116)
    
    // 注意：QUIC Spec 中包含 TLS ClientHelloID
    // 可以在自定义 Spec 中指定 uTLS 指纹

    uRoundTripper := http3.GetURoundTripper(
        roundTripper,
        &quicSpec,
        nil,
    )
    defer uRoundTripper.Close()

    client := &http.Client{Transport: uRoundTripper}
    // 使用客户端...
}
```

### 2. 封装成可复用函数

```go
package main

import (
    "fmt"
    "io"
    "log"
    "net/http"

    tls "github.com/refraction-networking/utls"
    "github.com/refraction-networking/uquic"
    "github.com/refraction-networking/uquic/http3"
)

// 创建 uQUIC HTTP/3 客户端
func createH3Client(quicID uquic.QUICID) (*http.Client, error) {
    roundTripper := &http3.RoundTripper{
        TLSClientConfig: &tls.Config{},
        QuicConfig:      &uquic.Config{},
    }

    quicSpec, err := uquic.QUICID2Spec(quicID)
    if err != nil {
        return nil, fmt.Errorf("获取 QUIC Spec 失败: %w", err)
    }

    uRoundTripper := http3.GetURoundTripper(
        roundTripper,
        &quicSpec,
        nil,
    )

    client := &http.Client{
        Transport: uRoundTripper,
    }

    return client, nil
}

func main() {
    // 使用 Firefox 指纹
    client, err := createH3Client(uquic.QUICFirefox_116)
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.Get("https://example.com")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("响应长度: %d\n", len(body))
}
```

---

## 注意事项与最佳实践

### ⚠️ 重要警告

#### 1. 生产环境警告

**不要在生产环境使用 uQUIC！**

- 该项目仍处于概念验证阶段
- 研究论文尚未发表
- 可能未经过充分测试

#### 2. 指纹识别风险

- 某些误用可能导致更容易被指纹识别
- 模拟可能无法完全与现实 QUIC 客户端区分
- 理解库的风险和局限性

#### 3. 开发路线图

根据 GitHub 仓库，以下功能正在开发中：

**已完成**：
- ✅ QUIC Header 自定义
- ✅ QUIC Frame 自定义（Crypto、Padding、Ping）
- ✅ TLS ClientHello 消息自定义（通过 uTLS）
- ✅ QUIC Transport Parameters 自定义
- ✅ Chrome 和 Firefox 预设指纹

**进行中**：
- 🚧 QUIC ACK Frame 自定义
- 🚧 Initial ACK 行为自定义
- 🚧 Initial Retry 行为自定义
- 🚧 Safari 和 Edge 预设指纹

### 最佳实践

#### 1. 错误处理

```go
quicSpec, err := uquic.QUICID2Spec(uquic.QUICFirefox_116)
if err != nil {
    log.Fatalf("获取 QUIC Spec 失败: %v", err)
    // 或者返回错误给调用者
}

uRoundTripper := http3.GetURoundTripper(
    roundTripper,
    &quicSpec,
    nil,
)
defer uRoundTripper.Close()  // 确保关闭
```

#### 2. 超时设置

```go
client := &http.Client{
    Transport: uRoundTripper,
    Timeout:   30 * time.Second,  // 设置超时
}

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    log.Fatal(err)
}
```

#### 3. 连接重用

```go
// 在循环中使用同一个 client
client := &http.Client{
    Transport: uRoundTripper,
}

for _, url := range urls {
    resp, err := client.Get(url)
    if err != nil {
        log.Printf("请求 %s 失败: %v", url, err)
        continue
    }
    // 处理响应...
    resp.Body.Close()
}

// 最后关闭 RoundTripper
defer uRoundTripper.Close()
```

#### 4. 日志和调试

```go
import "log"

// 启用详细日志
roundTripper := &http3.RoundTripper{
    TLSClientConfig: &tls.Config{},
    QuicConfig:      &uquic.Config{
        // 配置日志级别
    },
}
```

---

## 常见问题解答

### Q1: uQUIC 和 quic-go 有什么区别？

**A:** uQUIC 是 quic-go 的分支，主要区别是：
- uQUIC 提供对 Initial Packet 的低级访问
- 可以自定义未加密的 Initial Packet
- 提供预设浏览器指纹
- 与 uTLS 集成使用

### Q2: 为什么说 uQUIC 不适用于生产环境？

**A:** 
- 项目仍处于概念验证阶段
- 研究论文尚未发表和同行评议
- 可能存在未知的 bug 和安全隐患
- 功能仍在积极开发中

### Q3: 如何使用 uQUIC 进行 HTTP/3 请求？

**A:** 
1. 创建 HTTP/3 RoundTripper
2. 获取 QUIC Spec
3. 转换为 uQUIC RoundTripper
4. 创建 HTTP 客户端
5. 发送请求

详见 [HTTP/3 客户端示例](#http3-客户端示例)。

### Q4: 支持哪些浏览器指纹？

**A:** 目前支持：
- Chrome 115
- Firefox 116

Safari 和 Edge 指纹正在开发中。

### Q5: 如何自定义 QUIC Initial Packet？

**A:** 
1. 查看 `u_parrots.go` 中的示例
2. 实现自定义 QUIC Spec
3. 指定 QUIC Header、Frames、TLS ClientHello 等

### Q6: uQUIC 和 uTLS 的关系是什么？

**A:** 
- uQUIC 通过 uTLS 自定义 TLS ClientHello
- 两者配合提供完整的 HTTP/3 客户端指纹对抗
- uQUIC 处理 QUIC 层，uTLS 处理 TLS 层

### Q7: 是否有性能开销？

**A:** 
- uQUIC 继承自 quic-go，性能影响很小
- Initial Packet 自定义可能在握手阶段有轻微开销
- 建议进行实际性能测试

### Q8: 如何处理连接错误？

**A:** 
```go
resp, err := client.Get(url)
if err != nil {
    // 检查错误类型
    if err.Error() == "EOF" {
        // 连接关闭
    } else {
        // 其他错误
    }
    log.Printf("错误: %v", err)
}
```

### Q9: 可以用于反审查吗？

**A:** 
- 可以，但需要理解风险和局限性
- 建议测试多种指纹
- 注意库的开发状态
- 遵守当地法律法规

### Q10: 如何贡献代码？

**A:** 
- 在 GitHub 上提交 Issue
- 发送 Pull Request
- 联系维护者：gaukas.wang@colorado.edu
- 贡献新的浏览器指纹

---

## 参考资料

### 官方资源

- **uQUIC GitHub 仓库：** https://github.com/refraction-networking/uquic
- **官方文档：** https://godoc.org/github.com/refraction-networking/uquic
- **最新发布：** https://github.com/refraction-networking/uquic/releases/latest

### 相关项目

- **quic-go：** https://github.com/quic-go/quic-go（上游项目）
- **uTLS：** https://github.com/refraction-networking/utls（TLS 指纹对抗）
- **clienthellod：** https://github.com/gaukas/clienthellod（ClientHello 数据库）

### 协议规范

- **QUIC 协议：** https://tools.ietf.org/html/rfc9000
- **HTTP/3 协议：** https://tools.ietf.org/html/rfc9114
- **TLS 1.3 协议：** https://tools.ietf.org/html/rfc8446

### 指纹测试工具

- **TLS Fingerprint：** https://quic.tlsfingerprint.io/qfp/
- **其他指纹工具：** https://tlsfingerprint.io/

---

## 更新日志

### v0.0.6（2024-07-19）

最新稳定版本

**更新内容**：
- 同步上游 quic-go v0.42.0
- 修复 CipherSuitesTLS13 链接问题
- 修复 MaybePackProbePacket 使用 QUIC spec 的问题
- 依赖更新到最新版本
- 性能优化和 bug 修复

**依赖更新**：
- golang.org/x/crypto: 0.14.0 → 0.17.0
- golang.org/x/net: 0.17.0 → 0.23.0
- github.com/quic-go/quic-go: 0.39.2 → 0.42.0
- github.com/cloudflare/circl: 1.3.5 → 1.3.7

### v0.0.5

- 初始浏览器指纹支持
- 基础功能实现

### 未来规划

根据开发路线图，未来将添加：
- Safari 和 Edge 预设指纹
- ACK Frame 自定义
- Initial ACK/Retry 行为自定义
- 更多性能和稳定性改进

---

## 贡献

如果你发现本文档有任何问题或有改进建议，欢迎提出 Issue 或 Pull Request。

### 联系方式

- **GitHub：** https://github.com/refraction-networking/uquic
- **维护者邮箱：** gaukas.wang@colorado.edu
- **Issues：** https://github.com/refraction-networking/uquic/issues
- **Discussions：** https://github.com/refraction-networking/uquic/discussions

---

**最后更新：** 2025-01-10  
**文档版本：** 1.0.0  
**uQUIC 版本：** v0.0.6（最新版本）  
**Go 版本要求：** 1.20+  
**项目状态：** ⚠️ 概念验证阶段，不推荐生产使用
