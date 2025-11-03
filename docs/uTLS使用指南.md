# uTLS 全面使用指南

## 目录
1. [什么是 uTLS](#什么是-utls)
2. [安装与导入](#安装与导入)
3. [基础使用](#基础使用)
4. [浏览器指纹模拟](#浏览器指纹模拟)
5. [高级自定义配置](#高级自定义配置)
6. [完整示例代码](#完整示例代码)
7. [注意事项与最佳实践](#注意事项与最佳实践)
8. [常见问题解答](#常见问题解答)
9. [参考资料](#参考资料)
10. [更新日志](#更新日志)
11. [文档更新日志](#文档更新日志)

---

## 什么是 uTLS

uTLS（universal TLS / unconventional TLS）是一个用于 Go 语言的 TLS 库，它是对 Go 标准库 TLS 的分支，旨在通过模拟主流浏览器的 TLS 握手指纹，增强网络流量的隐匿性，减少被检测的风险。

### 主要功能

- **模拟浏览器指纹**：能够模仿不同浏览器的 TLS 握手指纹，如 Chrome、Firefox、Safari、Edge 等
- **自定义 TLS 扩展**：允许开发者自定义 TLS 扩展，如加密套件、压缩方法、客户端随机值等
- **低级别访问**：提供对 ClientHello 消息的读写访问
- **QUIC 集成**：支持与 QUIC 协议的集成，适用于 HTTP/3 等新兴协议的开发
- **会话票据管理**：支持伪造会话票据，以实现更灵活的会话管理

### 应用场景

- 网络代理客户端开发
- 爬虫系统开发
- 隐私保护工具开发
- 需要绕过 TLS 指纹检测的应用

---

## 安装与导入

### 0. 版本要求

- **最低 Go 版本：** Go 1.21+
- **推荐 Go 版本：** Go 1.22+ 
- **最新 uTLS 版本：** v1.8.1（2025-10-14 发布）

### 1. 安装 uTLS 库

```bash
# 安装最新版本（推荐）
go get -u github.com/refraction-networking/utls@v1.8.1

# 或安装到最新版本
go get -u github.com/refraction-networking/utls

# 如果使用 go.mod
go mod edit -require=github.com/refraction-networking/utls@v1.8.1
go mod tidy
```

> **Chrome 120+ 用户注意：** 如果你使用 Chrome 120 及以上版本的指纹，必须使用 v1.8.1 或更高版本！

### 2. 导入 uTLS 包

```go
import (
    "github.com/refraction-networking/utls"
    "crypto/tls"
    "net"
)
```

---

## 基础使用

### 最简单的示例

```go
package main

import (
    "fmt"
    "net"
    "github.com/refraction-networking/utls"
)

func main() {
    // 1. 创建 TCP 连接
    conn, err := net.Dial("tcp", "example.com:443")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // 2. 创建 uTLS 配置
    config := &utls.Config{
        ServerName: "example.com",
    }

    // 3. 创建 uTLS 客户端连接（模拟 Chrome）
    uconn := utls.UClient(conn, config, utls.HelloChrome_Auto)

    // 4. 执行 TLS 握手
    err = uconn.Handshake()
    if err != nil {
        panic(err)
    }

    // 5. 使用连接进行通信
    // 例如：发送 HTTP 请求
    fmt.Println("TLS 握手成功！")
}
```

---

## 浏览器指纹模拟

uTLS 支持模拟多种浏览器的 TLS 握手指纹，这些指纹决定了你的 TLS 流量看起来像哪个浏览器。

### 支持的浏览器指纹（v1.8.1）

> **重要提示：** Chrome 120 及以上版本的用户必须更新到 v1.8.1，v1.8.1 修复了 Chrome≥120 指纹的重要 bug。

```go
// Chrome 系列（最新支持到 Chrome 133）
utls.HelloChrome_Auto         // Chrome 最新版本（自动更新，推荐 v1.7.0+）
utls.HelloChrome_58           // Chrome 58
utls.HelloChrome_62           // Chrome 62
utls.HelloChrome_70           // Chrome 70
utls.HelloChrome_72           // Chrome 72
utls.HelloChrome_83           // Chrome 83
utls.HelloChrome_87           // Chrome 87
utls.HelloChrome_96           // Chrome 96
utls.HelloChrome_100          // Chrome 100
utls.HelloChrome_102          // Chrome 102
utls.HelloChrome_106_Shuffle  // Chrome 106（混淆版本）
utls.HelloChrome_112_PSK_Shuf // Chrome 112（PSK + 混淆）
utls.HelloChrome_114_Padding_PSK_Shuf // Chrome 114（Padding + PSK + 混淆）
utls.HelloChrome_115_PQ       // Chrome 115（后量子加密）
utls.HelloChrome_115_PQ_PSK   // Chrome 115（后量子 + PSK）
utls.HelloChrome_120          // Chrome 120（v1.8.1 已修复）
utls.HelloChrome_120_PQ       // Chrome 120（后量子加密）
utls.HelloChrome_131          // Chrome 131（v1.7.0+）
utls.HelloChrome_133          // Chrome 133（v1.8.0+）

// Firefox 系列
utls.HelloFirefox_Auto        // Firefox 最新版本（自动更新）
utls.HelloFirefox_55          // Firefox 55
utls.HelloFirefox_56          // Firefox 56
utls.HelloFirefox_63          // Firefox 63
utls.HelloFirefox_65          // Firefox 65
utls.HelloFirefox_99          // Firefox 99
utls.HelloFirefox_102         // Firefox 102
utls.HelloFirefox_105         // Firefox 105
utls.HelloFirefox_120         // Firefox 120（v1.7.0+）

// Edge 系列
utls.HelloEdge_Auto           // Edge 最新版本（自动更新）
utls.HelloEdge_85             // Edge 85
utls.HelloEdge_106            // Edge 106（v1.8.0 已修复）

// iOS Safari 系列
utls.HelloIOS_Auto            // iOS Safari 最新版本（自动更新）
utls.HelloIOS_11_1            // iOS Safari 11.1
utls.HelloIOS_12_1            // iOS Safari 12.1
utls.HelloIOS_13              // iOS Safari 13
utls.HelloIOS_14              // iOS Safari 14

// Safari 系列（macOS）
utls.HelloSafari_Auto         // Safari 最新版本（自动更新）

// 随机指纹系列（推荐用于绕过检测）
utls.HelloRandomized          // 随机指纹（可能包含 ALPN）
utls.HelloRandomizedALPN      // 随机指纹（强制包含 ALPN）
utls.HelloRandomizedNoALPN    // 随机指纹（不包含 ALPN）
```

### 重要变更（v1.8.1）

- **Chrome 120+ 必须更新**：v1.8.1 修复了 Chrome≥120 指纹的 critical bug，请立即更新
- **新增 Chrome 133 支持**：v1.8.0 添加
- **新增 Chrome 131 支持**：v1.7.0 添加
- **新增 Firefox 120 支持**：v1.7.0 添加
- **修复 Edge 106 bug**：v1.8.0 修复
- **改善随机指纹**：v1.8.0 添加更多熵值，更不易检测

### 选择指纹的示例

```go
package main

import (
    "fmt"
    "net"
    "github.com/refraction-networking/utls"
)

func main() {
    conn, err := net.Dial("tcp", "example.com:443")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    config := &utls.Config{
        ServerName: "example.com",
    }

    // 方案1：使用 Chrome 指纹
    uconn1 := utls.UClient(conn, config, utls.HelloChrome_Auto)
    err = uconn1.Handshake()
    if err != nil {
        fmt.Println("Chrome 握手失败:", err)
    }

    // 方案2：使用 Firefox 指纹（推荐，更安全）
    uconn2 := utls.UClient(conn, config, utls.HelloFirefox_Auto)
    err = uconn2.Handshake()
    if err != nil {
        fmt.Println("Firefox 握手失败:", err)
    }

    // 方案3：使用随机指纹
    uconn3 := utls.UClient(conn, config, utls.HelloRandomized)
    err = uconn3.Handshake()
    if err != nil {
        fmt.Println("随机指纹握手失败:", err)
    }
}
```

---

## 高级自定义配置

如果你需要完全控制 TLS 握手的各个方面，可以使用自定义配置。

### 自定义 ClientHello

```go
package main

import (
    "net"
    "github.com/refraction-networking/utls"
)

func main() {
    conn, err := net.Dial("tcp", "example.com:443")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    config := &utls.Config{
        ServerName: "example.com",
    }

    // 创建 uTLS 连接
    uconn := utls.UClient(conn, config, utls.HelloCustom)

    // 定义自定义 ClientHello 规格
    spec := utls.ClientHelloSpec{
        // 设置加密套件
        CipherSuites: []uint16{
            utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        },
        
        // 设置 TLS 扩展
        Extensions: []utls.TLSExtension{
            // SNI 扩展
            &utls.SNIExtension{ServerName: "example.com"},
            
            // 支持的椭圆曲线扩展
            &utls.SupportedCurvesExtension{
                Curves: []utls.CurveID{
                    utls.X25519,
                    utls.CurveP256,
                    utls.CurveP384,
                    utls.CurveP521,
                },
            },
            
            // 支持的签名算法扩展
            &utls.SupportedSignatureAlgorithmsExtension{
                SignatureAndHashes: []utls.SignatureScheme{
                    utls.ECDSAWithP256AndSHA256,
                    utls.ECDSAWithP384AndSHA384,
                    utls.ECDSAWithP521AndSHA512,
                    utls.Ed25519,
                    utls.RSAWithPSSAndSHA256,
                    utls.RSAWithPSSAndSHA384,
                    utls.RSAWithPSSAndSHA512,
                    utls.PKCS1WithSHA256,
                    utls.PKCS1WithSHA384,
                    utls.PKCS1WithSHA512,
                },
            },
            
            // ALPN 扩展
            &utls.ALPNExtension{
                AlpnProtocols: []string{"h2", "http/1.1"},
            },
            
            // 压缩方法扩展
            &utls.CompressCertificateExtension{
                Algorithms: []utls.CertCompressionAlgo{
                    utls.CertCompressionZlib,
                },
            },
            
            // Padding 扩展（用于混淆流量大小）
            &utls.UtlsPaddingExtension{
                GetPaddingLen: utls.BoringPaddingStyle,
            },
        },
    }

    // 应用自定义规格
    err = uconn.ApplyPreset(&spec)
    if err != nil {
        panic(err)
    }

    // 执行握手
    err = uconn.Handshake()
    if err != nil {
        panic(err)
    }

    // 使用连接
}
```

### 更多自定义选项

```go
// TLS 版本控制
config := &utls.Config{
    MinVersion: utls.VersionTLS12,  // v1.8.1 推荐最低 TLS 1.2
    MaxVersion: utls.VersionTLS13,
}

// 证书验证跳过（仅用于测试）
config := &utls.Config{
    InsecureSkipVerify: true,
}

// 会话复用
config := &utls.Config{
    SessionTicketsDisabled: false,
    ClientSessionCache: utls.NewLRUClientSessionCache(128),
}

// 自定义套接字选项
uconn.SetReadDeadline(time.Now().Add(30 * time.Second))
uconn.SetWriteDeadline(time.Now().Add(30 * time.Second))
```

### v1.7.0+ 新特性

#### 1. ML-KEM 后量子加密支持

Chrome 115+ 引入了 ML-KEM（后量子加密）支持，uTLS v1.7.0+ 已支持：

```go
// 使用 Chrome 115 后量子加密指纹
uconn := utls.UClient(conn, config, utls.HelloChrome_115_PQ)

// 或带 PSK 的后量子指纹
uconn := utls.UClient(conn, config, utls.HelloChrome_115_PQ_PSK)
```

#### 2. ECH 支持改进（v1.7.0+）

ECH（Encrypted ClientHello）支持已在 v1.7.0 改进，现在在自定义 ClientHello 中使用时更稳定。

#### 3. 使用 Roller 自动切换指纹（v1.8.0+）

推荐的用法，自动在多个最新指纹间切换，直到找到可用的：

```go
package main

import (
    "fmt"
    "github.com/refraction-networking/utls"
)

func main() {
    // 创建 Roller，自动使用最新的多个指纹
    roller, err := utls.NewRoller()
    if err != nil {
        panic(err)
    }

    // Dial 会自动尝试不同的指纹直到成功
    conn, err := roller.Dial("tcp", "example.com:443", "example.com")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // 后续连接会重用成功工作的指纹
    fmt.Println("连接成功！")
}
```

---

## 完整示例代码

### 示例1：简单的 HTTPS 请求

```go
package main

import (
    "fmt"
    "io"
    "net"
    "net/http"
    "github.com/refraction-networking/utls"
)

func main() {
    // 1. 建立 TCP 连接
    conn, err := net.Dial("tcp", "www.example.com:443")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // 2. 创建 uTLS 配置
    config := &utls.Config{
        ServerName: "www.example.com",
    }

    // 3. 创建 uTLS 客户端连接（模拟 Firefox）
    uconn := utls.UClient(conn, config, utls.HelloFirefox_Auto)

    // 4. 执行 TLS 握手
    err = uconn.Handshake()
    if err != nil {
        panic(err)
    }

    // 5. 创建 HTTP 客户端
    client := &http.Client{
        Transport: &http.Transport{
            DialTLS: func(network, addr string) (net.Conn, error) {
                return uconn, nil
            },
        },
    }

    // 6. 发送 HTTPS 请求
    resp, err := client.Get("https://www.example.com")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // 7. 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    fmt.Println("响应状态:", resp.Status)
    fmt.Println("响应长度:", len(body))
}
```

### 示例2：封装成函数

```go
package main

import (
    "fmt"
    "net"
    "github.com/refraction-networking/utls"
)

// 创建 uTLS 连接的通用函数
func createUTLSConnection(addr string, serverName string, fingerprint utls.ClientHelloID) (*utls.UConn, error) {
    // 建立 TCP 连接
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return nil, fmt.Errorf("建立 TCP 连接失败: %w", err)
    }

    // 创建 uTLS 配置
    config := &utls.Config{
        ServerName:         serverName,
        InsecureSkipVerify: false,
    }

    // 创建 uTLS 客户端
    uconn := utls.UClient(conn, config, fingerprint)

    // 执行握手
    err = uconn.Handshake()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("TLS 握手失败: %w", err)
    }

    return uconn, nil
}

func main() {
    // 使用不同的指纹创建连接
    fingerprints := []utls.ClientHelloID{
        utls.HelloChrome_Auto,
        utls.HelloFirefox_Auto,
        utls.HelloSafari_Auto,
        utls.HelloEdge_Auto,
    }

    for _, fingerprint := range fingerprints {
        fmt.Printf("尝试使用指纹: %s\n", fingerprint.Client)
        
        uconn, err := createUTLSConnection(
            "www.example.com:443",
            "www.example.com",
            fingerprint,
        )
        
        if err != nil {
            fmt.Printf("失败: %v\n\n", err)
            continue
        }
        
        fmt.Println("成功建立连接！")
        fmt.Printf("TLS 版本: %s\n", uconn.ConnectionState().Version)
        fmt.Printf("加密套件: %s\n\n", uconn.ConnectionState().CipherSuite)
        
        uconn.Close()
    }
}
```

### 示例3：自定义指纹的高级用法

```go
package main

import (
    "crypto/tls"
    "fmt"
    "net"
    "github.com/refraction-networking/utls"
)

func main() {
    conn, err := net.Dial("tcp", "example.com:443")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    config := &utls.Config{
        ServerName: "example.com",
    }

    // 使用自定义指纹
    uconn := utls.UClient(conn, config, utls.HelloCustom)

    // 基于 Chrome 指纹，但自定义修改
    chromeSpec := utls.ClientHelloSpec{
        TLSVersMin: utls.VersionTLS12,
        TLSVersMax: utls.VersionTLS13,
        CipherSuites: []uint16{
            utls.TLS_AES_128_GCM_SHA256,
            utls.TLS_AES_256_GCM_SHA384,
            utls.TLS_CHACHA20_POLY1305_SHA256,
            utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
            utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
        },
        CompressionMethods: []byte{0},
        Extensions: []utls.TLSExtension{
            &utls.UtlsGREASEExtension{},
            &utls.SNIExtension{ServerName: "example.com"},
            &utls.UtlsExtendedMasterSecretExtension{},
            &utls.RenegotiationInfoExtension{Renegotiation: utls.RenegotiateOnceAsClient},
            &utls.SupportedCurvesExtension{
                Curves: []utls.CurveID{
                    utls.X25519,
                    utls.CurveP256,
                    utls.CurveP384,
                    utls.CurveP521,
                    utls.GREASE_PLACEHOLDER,
                    utls.CurveIDs([]utls.CurveID{utls.X25519})[0],
                },
            },
            &utls.SupportedPointsExtension{
                SupportedPoints: []uint8{0},
            },
            &utls.SessionTicketExtension{},
            &utls.ALPNExtension{
                AlpnProtocols: []string{"h2", "http/1.1"},
            },
            &utls.StatusRequestExtension{},
            &utls.SignatureAlgorithmsExtension{
                SupportedSignatureAlgorithms: []utls.SignatureScheme{
                    utls.ECDSAWithP256AndSHA256,
                    utls.Ed25519,
                    utls.ECDSAWithP384AndSHA384,
                    utls.ECDSAWithP521AndSHA512,
                    utls.RSAWithPSSAndSHA256,
                    utls.RSAWithPSSAndSHA384,
                    utls.RSAWithPSSAndSHA512,
                    utls.PKCS1WithSHA256,
                    utls.PKCS1WithSHA384,
                    utls.PKCS1WithSHA512,
                    utls.RSAWithSHA1,
                    utls.ECDSAWithSHA1,
                },
            },
            &utls.SCTExtension{},
            &utls.KeyShareExtension{
                KeyShares: []utls.KeyShare{
                    {Group: utls.X25519},
                },
            },
            &utls.PSKKeyExchangeModesExtension{
                Modes: []uint8{utls.PskModeDHE},
            },
            &utls.SupportedVersionsExtension{
                Versions: []uint16{
                    utls.VersionTLS13,
                    utls.VersionTLS12,
                },
            },
            &utls.CompressCertificateExtension{
                Algorithms: []utls.CertCompressionAlgo{
                    utls.CertCompressionBrotli,
                },
            },
            &utls.UtlsGREASEExtension{},
            &utls.UtlsPaddingExtension{
                GetPaddingLen: utls.BoringPaddingStyle,
            },
        },
    }

    // 应用规格
    err = uconn.ApplyPreset(&chromeSpec)
    if err != nil {
        panic(err)
    }

    // 执行握手
    err = uconn.Handshake()
    if err != nil {
        panic(err)
    }

    // 获取连接信息
    state := uconn.ConnectionState()
    fmt.Println("TLS 版本:", tlsVersion(state.Version))
    fmt.Println("加密套件:", state.CipherSuite)
    fmt.Println("服务器名称:", state.ServerName)
    fmt.Println("握手完成时间:", state.HandshakeComplete)
}

func tlsVersion(version uint16) string {
    switch version {
    case tls.VersionTLS10:
        return "TLS 1.0"
    case tls.VersionTLS11:
        return "TLS 1.1"
    case tls.VersionTLS12:
        return "TLS 1.2"
    case tls.VersionTLS13:
        return "TLS 1.3"
    default:
        return fmt.Sprintf("Unknown (%x)", version)
    }
}
```

---

## 注意事项与最佳实践

### ⚠️ 重要安全提示

#### 0. 🚨 紧急更新：v1.8.1（2025-10-14 发布）

**Chrome 120+ 用户必须更新到 v1.8.1！**

v1.8.1 修复了一个 critical bug（#375）：Chrome≥120 指纹在 GREASE ECH 扩展中使用了错误的加密算法。这会导致：
- Chrome 120、131、133 等指纹失效
- 连接可能被拒绝或检测
- 潜在的安全风险

**立即行动：**
```bash
go get -u github.com/refraction-networking/utls@v1.8.1
```

如果你使用的是 Chrome 120+ 指纹（`HelloChrome_120`、`HelloChrome_131`、`HelloChrome_133` 等），请务必升级。

#### 1. 指纹泄露漏洞

**历史问题：** uTLS 在较旧版本（2023-2025 早期）曾存在指纹泄露漏洞，某些指纹生成矛盾的 TLS 指纹。

**当前状态：** 最新版本（v1.8.1）已修复相关问题。

**最佳实践：**
- ✅ 使用最新版本的 uTLS 库（v1.8.1 或更高）
- ✅ 优先使用 `HelloFirefox_Auto` 等非 Chrome 指纹
- ✅ 使用 `HelloRandomized` 随机指纹提高隐蔽性
- ✅ 定期更新 uTLS 到最新版本

```go
// ✅ 推荐：使用 Firefox 指纹
uconn := utls.UClient(conn, config, utls.HelloFirefox_Auto)

// ✅ 也推荐：使用随机指纹
uconn := utls.UClient(conn, config, utls.HelloRandomized)

// ⚠️ 谨慎使用：Chrome 指纹可能被检测
uconn := utls.UClient(conn, config, utls.HelloChrome_Auto)
```

#### 2. 传输方式支持

uTLS 仅在部分传输方式中受支持：
- ✅ **支持：** TCP、WebSocket
- ❌ **不支持：** Unix Domain Socket、其他自定义传输方式

在不支持的传输方式中使用 uTLS 可能导致程序异常退出。

#### 3. 不要跳过证书验证（生产环境）

```go
// ❌ 生产环境不要这样做
config := &utls.Config{
    InsecureSkipVerify: true,
}

// ✅ 正确的做法
config := &utls.Config{
    ServerName:         "example.com",
    InsecureSkipVerify: false,
}
```

#### 4. 连接超时设置

```go
// 设置连接超时
conn, err := net.DialTimeout("tcp", "example.com:443", 10*time.Second)
if err != nil {
    panic(err)
}

// 设置读取/写入超时
uconn.SetReadDeadline(time.Now().Add(30 * time.Second))
uconn.SetWriteDeadline(time.Now().Add(30 * time.Second))
```

### 最佳实践

#### 1. 错误处理

```go
func createConnection(addr string, serverName string, fingerprint utls.ClientHelloID) (*utls.UConn, error) {
    conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
    if err != nil {
        return nil, fmt.Errorf("建立连接失败: %w", err)
    }

    config := &utls.Config{
        ServerName:         serverName,
        InsecureSkipVerify: false,
        MinVersion:         utls.VersionTLS12,
        MaxVersion:         utls.VersionTLS13,
    }

    uconn := utls.UClient(conn, config, fingerprint)
    
    err = uconn.Handshake()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("握手失败: %w", err)
    }

    return uconn, nil
}
```

#### 2. 连接池管理

```go
type ConnectionPool struct {
    connections chan *utls.UConn
    maxSize     int
}

func NewConnectionPool(maxSize int) *ConnectionPool {
    return &ConnectionPool{
        connections: make(chan *utls.UConn, maxSize),
        maxSize:     maxSize,
    }
}

func (p *ConnectionPool) Get() *utls.UConn {
    select {
    case conn := <-p.connections:
        return conn
    default:
        return nil
    }
}

func (p *ConnectionPool) Put(conn *utls.UConn) {
    select {
    case p.connections <- conn:
    default:
        conn.Close()
    }
}
```

#### 3. 重试机制

```go
func connectWithRetry(addr string, serverName string, fingerprint utls.ClientHelloID, maxRetries int) (*utls.UConn, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        conn, err := createUTLSConnection(addr, serverName, fingerprint)
        if err == nil {
            return conn, nil
        }
        lastErr = err
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    return nil, fmt.Errorf("重试 %d 次后仍失败: %w", maxRetries, lastErr)
}
```

#### 4. 监控与日志

```go
type ConnLogger struct {
    *utls.UConn
}

func (c *ConnLogger) Write(data []byte) (int, error) {
    n, err := c.UConn.Write(data)
    log.Printf("写入 %d 字节到连接", n)
    return n, err
}

func (c *ConnLogger) Read(data []byte) (int, error) {
    n, err := c.UConn.Read(data)
    log.Printf("从连接读取 %d 字节", n)
    return n, err
}
```

---

## 常见问题解答

### Q1: 我应该使用哪个指纹？

**A:** 推荐使用 `HelloFirefox_Auto` 或 `HelloRandomized`，因为它们更不容易被检测。避免使用 `HelloChrome_Auto` 除非你确定你的环境支持它。

### Q2: uTLS 和标准 TLS 库有什么区别？

**A:** 
- 标准 TLS 库：提供标准的 TLS 实现，所有 Go 程序的指纹都相同
- uTLS：可以模拟不同浏览器的 TLS 指纹，提供更强的隐私保护

### Q3: 使用 uTLS 是否违法？

**A:** uTLS 本身是合法的技术工具，但使用方式决定了合法性。请确保在合法的场景下使用，遵守当地法律法规和网站的使用条款。

### Q4: uTLS 会影响性能吗？

**A:** uTLS 的性能影响微乎其微，TLS 握手阶段可能有 1-5% 的性能开销，但对整体性能影响不大。

### Q5: 如何处理 "bad certificate" 错误？

**A:** 首先检查服务器的证书是否有效。在测试环境下可以使用 `InsecureSkipVerify: true`，但在生产环境中请使用有效的证书和正确的 `ServerName`。

### Q6: 为什么有时候握手会失败？

**A:** 可能的原因：
1. 网络连接问题
2. 服务器不接受该指纹
3. 证书验证失败
4. TLS 版本不兼容

建议添加详细的错误处理和日志记录。

### Q7: 如何选择合适的 TLS 版本？

**A:** 
- TLS 1.3 是最新、最安全的版本
- TLS 1.2 是最广泛支持的版本
- 建议设置 `MinVersion: VersionTLS12, MaxVersion: VersionTLS13`

### Q8: uTLS 支持 HTTP/2 吗？

**A:** 是的，uTLS 支持 HTTP/2。在 ClientHello 中添加 ALPN 扩展即可：

```go
&utls.ALPNExtension{
    AlpnProtocols: []string{"h2", "http/1.1"},
}
```

### Q9: 如何调试 uTLS 连接问题？

**A:** 
1. 启用详细日志
2. 使用 `tls.Config` 的 `GetConfigForClient` 回调
3. 使用网络抓包工具（如 Wireshark）分析握手过程
4. 检查 `uconn.ConnectionState()` 的输出

### Q10: uTLS 可以用于客户端和服务端吗？

**A:** uTLS 主要设计用于客户端，虽然也支持服务端，但服务端使用并不常见。

---

## 参考资料

### 官方资源
- **GitHub 仓库：** https://github.com/refraction-networking/utls
- **官方文档：** https://pkg.go.dev/github.com/refraction-networking/utls

### 相关文章
- **指纹泄露漏洞分析：** https://blog.xiaohack.org/5016.html
- **V2Fly uTLS 配置：** https://www.v2fly.org/v5/config/stream.html
- **uTLS 高级用法：** https://opendeep.wiki/refraction-networking/utls/advanced-usage

### 其他资源
- **Go 标准 TLS 库：** https://pkg.go.dev/crypto/tls
- **TLS 协议规范：** https://tools.ietf.org/html/rfc8446

### 社区支持
- **Issues：** https://github.com/refraction-networking/utls/issues
- **讨论：** https://github.com/refraction-networking/utls/discussions

---

## 更新日志

### 2024 年
- 添加了 Chrome 115、120 及其后量子加密版本支持
- 修复了多个安全漏洞和指纹泄露问题
- 改进了 QUIC 集成

### 2025 年
- **v1.7.0**：添加 Chrome 131、Firefox 120 支持；合并 Go 1.23.4 和 1.24.0 更新；支持 ML-KEM 后量子加密；ECH 支持改进
- **v1.7.1-v1.7.3**：bug 修复，性能优化
- **v1.8.0**：添加 Chrome 133 支持；修复 Edge 106 spec 问题；改善随机指纹生成
- **v1.8.1**：**关键修复**：修复 Chrome≥120 的 GREASE ECH bug；修复 PubServerHelloMsg ServerShare 导出问题

---

## 文档更新日志

本文档从 v1.6.0 更新到 v1.8.1，以下是详细的更新内容。

### 📊 统计信息
- **原始版本**：v1.0.0（基于 uTLS v1.6.0）
- **当前版本**：v2.0.0（基于 uTLS v1.8.1）
- **文件行数**：887 行 → 987 行（+100 行，+11.3%）

### 🚀 主要更新内容

#### 1. 版本要求更新
- ✅ 添加 Go 1.21+ 最低版本要求
- ✅ 添加推荐 Go 1.22+ 说明
- ✅ 明确标注 v1.8.1 最新版本和发布日期

#### 2. 浏览器指纹完整更新
从旧版的 8 个 Chrome 指纹扩展到 15+ 个：

**新增 Chrome 系列指纹**：
- `HelloChrome_100` - Chrome 100
- `HelloChrome_106_Shuffle` - Chrome 106（混淆版本）
- `HelloChrome_112_PSK_Shuf` - Chrome 112（PSK + 混淆）
- `HelloChrome_114_Padding_PSK_Shuf` - Chrome 114（Padding + PSK + 混淆）
- `HelloChrome_115_PQ` - Chrome 115（后量子加密）
- `HelloChrome_115_PQ_PSK` - Chrome 115（后量子 + PSK）
- `HelloChrome_120_PQ` - Chrome 120（后量子加密）
- `HelloChrome_131` - Chrome 131（v1.7.0+）
- `HelloChrome_133` - Chrome 133（v1.8.0+）

**新增 Firefox 系列指纹**：
- `HelloFirefox_120` - Firefox 120（v1.7.0+）

#### 3. 🚨 v1.8.1 紧急更新警告
添加独立的紧急更新章节，详细说明：
- Chrome 120+ 用户必须更新的原因
- GREASE ECH bug（#375）的具体影响
- 立即升级的命令和步骤
- 受影响的具体指纹列表

#### 4. v1.7.0+ 新特性补充
- ✅ **ML-KEM 后量子加密支持**：Chrome 115+ 新增的后量子加密算法
- ✅ **ECH 支持改进**：更稳定的 Encrypted ClientHello 实现
- ✅ **Roller 功能**：详细的自动切换指纹示例代码
- ✅ **Edge 106 修复**：spec 问题的解决方案

#### 5. 更新日志完善
详细记录从 v1.7.0 到 v1.8.1 的每个版本变更，包括：
- 主要功能更新
- Bug 修复
- 性能优化
- 重要 API 变化

#### 6. 安全提示更新
- ✅ 重写指纹泄露漏洞说明，明确已修复状态
- ✅ 添加 v1.8.1 关键更新为第 0 条安全提示
- ✅ 更新最佳实践建议
- ✅ 强调 Chrome 120+ 更新的紧迫性

#### 7. 技术改进
- ✅ 添加 Roller 自动切换指纹的完整示例
- ✅ 补充 ML-KEM 后量子加密的使用说明
- ✅ 改进 ECH 支持的相关文档
- ✅ 优化代码示例的注释和说明

### 🔍 版本对比

#### 旧版本（v1.0.0）问题
1. ❌ 指纹列表过时（仅到 Chrome 120）
2. ❌ 缺少重要版本更新信息（v1.7.x - v1.8.1）
3. ❌ 未说明 Chrome 120+ 的 critical bug
4. ❌ 缺少 ML-KEM 后量子加密说明
5. ❌ 缺少 Roller 自动切换功能
6. ❌ 缺少版本要求说明

#### 新版本（v2.0.0）改进
1. ✅ 完整的指纹列表（到 Chrome 133）
2. ✅ 详细的版本更新历史
3. ✅ 醒目的安全警告和升级提示
4. ✅ 新技术支持说明（ML-KEM、ECH、Roller）
5. ✅ 更多实用工具介绍
6. ✅ 清晰的版本要求

### 📌 关键信息

**必须立即更新的用户**：
- 使用 `HelloChrome_120` 的用户
- 使用 `HelloChrome_131` 的用户
- 使用 `HelloChrome_133` 的用户
- 使用 `HelloChrome_Auto` 且版本为 v1.8.0 或更早的用户

**升级命令**：
```bash
go get -u github.com/refraction-networking/utls@v1.8.1
# 或
go mod edit -require=github.com/refraction-networking/utls@v1.8.1
go mod tidy
```

### 🎓 使用建议

1. **立即升级**：Chrome 120+ 指纹用户请立即升级到 v1.8.1
2. **使用最新指纹**：推荐使用 Chrome 133 或 Chrome 131
3. **考虑后量子**：Chrome 115_PQ 系列提供更好的安全性
4. **使用 Roller**：自动切换指纹，更易绕过检测
5. **定期更新**：保持 uTLS 库始终为最新版本

---

## 贡献

如果你发现本文档有任何问题或有改进建议，欢迎提出 Issue 或 Pull Request。

---

**最后更新：** 2025-01-10  
**文档版本：** 2.0.0  
**uTLS 版本：** v1.8.1（最新版本）
