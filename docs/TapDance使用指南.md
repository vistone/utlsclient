# TapDance / GoTapDance 完整使用指南

## 目录
1. [什么是 TapDance](#什么是-tapdance)
2. [安装与导入](#安装与导入)
3. [快速开始](#快速开始)
4. [命令行工具使用](#命令行工具使用)
5. [库方式集成](#库方式集成)
6. [配置参数详解](#配置参数详解)
7. [Conjure 协议支持](#conjure-协议支持)
8. [高级用法](#高级用法)
9. [注意事项与最佳实践](#注意事项与最佳实践)
10. [常见问题解答](#常见问题解答)
11. [参考资料](#参考资料)

---

## 什么是 TapDance

**TapDance** 是 Refraction Networking 家族的反审查技术之一，通过在 ISP（互联网服务提供商）放置代理服务器，使其难以被封锁。GoTapDance 是 TapDance 的跨平台 Golang 客户端实现。

根据 [Refraction Networking](https://refraction.network) 的资料，TapDance 与 Conjure 都是免费使用的反审查技术。

### 核心概念

**TapDance 隐蔽通道详解**：

TapDance 是一个**端到中间（End-to-Middle）**的反审查系统，其隐蔽通道的工作原理如下：

#### 1. Decoy（诱饵服务器）

- **身份**：公开的、合法的 Web 服务器
- **作用**：看起来像是用户要访问的正常网站
- **特点**：这些服务器本身不知道自己在为 TapDance 工作
- **伪装**：客户端到 Decoy 的连接看起来完全正常

```
客户端 -> [TLS 加密] -> Decoy (例如：site.io)
         ↓
      看起来像访问 site.io 的正常流量
```

#### 2. Station（隐蔽代理服务器）

- **身份**：隐藏在 ISP 网络中的代理服务器
- **部署位置**：互联网服务提供商（ISP）内部
- **隐蔽性**：审查者无法轻易识别和封锁
- **重定向**：Decoy 可以将某些流量转发给 Station

```
Decoy -> Station (隐藏在 ISP)
        ↓
   真正的代理转发
```

#### 3. 隐蔽通道的建立流程

**完整的数据流**：

```
用户客户端 → Decoy (site.io) → Station (ISP内部) → 目标网站
           
   [步骤1]    [步骤2]          [步骤3]          [步骤4]
   客户端      Decoy 被动       Station 代理      真实连接
   连接       重定向流量        转发请求          
```

**为什么叫"隐蔽通道"**：

1. **流量伪装**：客户端到 Decoy 的流量看起来完全正常，就像访问一个普通网站
2. **被动重定向**：Decoy 在不知情的情况下，将特定流量转发给 Station
3. **ISP 部署**：Station 部署在 ISP 内部，审查者难以识别和封锁
4. **端到中间**：不是端到端加密，而是客户端-中间代理-目标服务器的三层结构

#### 4. 传统代理 vs TapDance 隐蔽通道

**传统代理**：
```
客户端 -> 代理服务器 -> 目标网站
         ↑
      容易检测和封锁
      有明确的服务特征
```

**TapDance 隐蔽通道**：
```
客户端 -> Decoy -> Station -> 目标网站
         ↑       ↑
      像正常流量  隐藏在ISP
      难以检测    难以封锁
```

**TapDance 工作流程详解**：

1. **客户端启动**：向 Station 注册，获取 Decoy 列表
2. **连接 Decoy**：客户端连接到 Decoy（如 site.io），流量看起来正常
3. **触发重定向**：Decoy 检测到 TapDance 握手标记，被动转发到 Station
4. **Station 代理**：Station 作为代理，连接真正的目标网站
5. **数据转发**：所有数据经过三层转发，审查者只看到客户端到 Decoy 的正常流量

#### 5. 为什么难以被检测

**优势**：
- **无主动特征**：Station 不主动对外提供服务，只接收来自 Decoy 的转发
- **流量混淆**：所有流量看起来像正常的网站访问
- **分布式部署**：Station 分布在各个 ISP，无法批量封锁
- **动态 Decoy**：客户端定期轮换 Decoy，增加检测难度

### 主要特性

- **端到中间代理**：不是端到端，而是端到中间再到端
- **被动流量伪装**：流量看起来像正常的互联网流量
- **无需客户端配置**：客户端自动配置
- **跨平台支持**：Windows、macOS、Linux、Android

### 应用场景

- 绕过网络审查
- 访问被封锁的网站和服务
- 保护网络隐私
- 研究网络审查绕过技术

### ⚠️ 重要提示

**合法使用**：请确保在合法的场景下使用 TapDance。遵守当地法律法规和网络使用政策。

**研究用途**：该项目可以用于研究目的，了解反审查技术的实现原理。

---

## 安装与导入

### 版本要求

- **Go 版本：** 1.10+（项目使用 Go 1.22+ 进行测试）
- **最新版本：** v1.7.10

### 1. 安装 GoTapDance

```bash
# 方法1：使用 go get
go get -d -u -t github.com/refraction-networking/gotapdance/...

# 方法2：克隆仓库
git clone https://github.com/refraction-networking/gotapdance.git
cd gotapdance

# 更新依赖（如果需要）
go get -u all
```

### 2. 构建可执行文件

```bash
# 构建 CLI 工具
go build ./cli

# 或构建所有组件
go build ./...
```

### 3. 导入库

```go
import (
    "github.com/refraction-networking/gotapdance/tapdance"
    "github.com/refraction-networking/gotapdance/tdproxy"
)
```

---

## 快速开始

### 最基础的用法

```go
package main

import (
    "fmt"
    "github.com/refraction-networking/gotapdance/tapdance"
)

func main() {
    // 设置 assets 目录（包含 ClientConf 和 roots 文件）
    tapdance.AssetsSetDir("./path/to/assets/dir/")

    // 建立 TapDance 连接
    tdConn, err := tapdance.Dial("tcp", "censoredsite.com:80")
    if err != nil {
        fmt.Printf("tapdance.Dial() 失败: %+v\n", err)
        return
    }
    defer tdConn.Close()

    // tdConn 实现了标准的 net.Conn 接口
    // 可以像使用其他 Golang 连接一样使用
    
    // 发送 HTTP 请求
    _, err = tdConn.Write([]byte("GET / HTTP/1.1\r\nHost: censoredsite.com\r\n\r\n"))
    if err != nil {
        fmt.Printf("tdConn.Write() 失败: %+v\n", err)
        return
    }

    // 读取响应
    buf := make([]byte, 16384)
    n, err := tdConn.Read(buf)
    if err != nil {
        fmt.Printf("tdConn.Read() 失败: %+v\n", err)
        return
    }

    fmt.Printf("收到响应: %s\n", string(buf[:n]))
}
```

---

## 命令行工具使用

### 基本命令

```bash
./cli -connect-addr=<目标地址> [OPTIONS]
```

### 常用参数

#### 基础参数

```bash
# 指定监听端口（默认为 10500）
-port=10500

# 设置 assets 目录（默认 ./assets/）
-assetsdir=./assets/

# 启用调试日志
-debug

# 启用跟踪日志
-trace
```

#### 连接参数

```bash
# 指定目标地址（必需）
-connect-addr=example.com:443
# 或
-connect-addr=192.0.2.1:443

# 禁用 IPv6 诱饵（默认启用）
-disable-ipv6

# 发送代理头部
-proxy
```

#### Decoy 配置

```bash
# 使用单个 decoy（不请求 ClientConf）
-decoy="site.io,1.2.3.4"
# 或只指定域名，IP 会自动解析
-decoy="site.io"

# 设置诱饵宽度（每个连接发送的注册数）
-w=5
```

#### TLS/SSL 配置

```bash
# 将 SSL 密钥写入文件（用于 Wireshark 解密）
-tlslog=/path/to/ssl.log
```

### 完整示例

```bash
# 基础连接
./cli -connect-addr=example.com:443 -debug

# 使用指定 decoy
./cli -connect-addr=example.com:443 -decoy="cloudflare.com,104.16.132.229" -w=3

# 启用详细日志并保存 TLS 密钥
./cli -connect-addr=example.com:443 -trace -tlslog=./ssl.log
```

---

## 库方式集成

### 示例 1：基本连接

```go
package main

import (
    "fmt"
    "net"
    "github.com/refraction-networking/gotapdance/tapdance"
)

func main() {
    // 1. 设置 assets 目录
    tapdance.AssetsSetDir("./assets/")
    
    // 2. 建立 TapDance 连接
    conn, err := tapdance.Dial("tcp", "example.com:80")
    if err != nil {
        fmt.Printf("连接失败: %v\n", err)
        return
    }
    defer conn.Close()

    // 3. 发送和接收数据
    // conn 实现了 net.Conn 接口，可以直接使用
}
```

### 示例 2：配合 TLS

```go
package main

import (
    "fmt"
    "net"
    "crypto/tls"
    "github.com/refraction-networking/gotapdance/tapdance"
)

func main() {
    tapdance.AssetsSetDir("./assets/")
    
    // 建立 TapDance 连接
    tdConn, err := tapdance.Dial("tcp", "example.com:443")
    if err != nil {
        fmt.Printf("TapDance 连接失败: %v\n", err)
        return
    }
    defer tdConn.Close()

    // 在 TapDance 连接上建立 TLS 连接
    tlsConn := tls.Client(tdConn, &tls.Config{
        ServerName: "example.com",
    })
    defer tlsConn.Close()

    // 执行 TLS 握手
    err = tlsConn.Handshake()
    if err != nil {
        fmt.Printf("TLS 握手失败: %v\n", err)
        return
    }

    // 现在可以安全地发送和接收数据
    fmt.Println("TLS 握手成功！")
}
```

### 示例 3：HTTP 客户端

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "github.com/refraction-networking/gotapdance/tapdance"
)

func main() {
    tapdance.AssetsSetDir("./assets/")
    
    // 创建 HTTP 客户端，使用 TapDance 作为 Transport
    client := &http.Client{
        Transport: &http.Transport{
            Dial: func(network, addr string) (net.Conn, error) {
                return tapdance.Dial(network, addr)
            },
        },
    }

    // 发送 HTTP 请求
    resp, err := client.Get("https://example.com")
    if err != nil {
        fmt.Printf("HTTP 请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()

    // 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("读取响应失败: %v\n", err)
        return
    }

    fmt.Printf("响应长度: %d 字节\n", len(body))
}
```

---

## 配置参数详解

### Assets 目录

TapDance 需要 assets 目录，包含以下文件：

1. **ClientConf**：客户端配置，包含 decoy 列表等信息
2. **roots**：根证书文件，用于验证站点的 TLS 证书

```go
// 设置 assets 目录
tapdance.AssetsSetDir("./assets/")

// 确保该目录：
// 1. 可被 td 进程读写
// 2. 包含必需的配置文件
// 3. 权限设置正确（仅 td 进程可访问）
```

### 配置文件位置

```
assets/
├── ClientConf
└── roots
```

**注意**：你需要从 TapDance 项目获取这些配置文件。

---

## Conjure 协议支持

TapDance 客户端现已支持 **Conjure** 协议。Conjure 是 Refraction Networking 的另一个反审查技术。

### Conjure vs TapDance

**TapDance**：
- 成熟的协议
- 使用固定的 decoy 列表
- 被动流量重定向

**Conjure**：
- 较新的协议
- 使用 CDN 诱饵
- 更动态的注册机制

### 使用 Conjure

命令行工具支持 Conjure：

```bash
# 使用 Conjure
./cli -connect-addr=example.com:443 \
      -registrar=decoy \
      -transport=min
```

### Conjure 注册方式

```bash
# decoy：使用 decoy 注册
-registrar=decoy

# api：使用 API 注册
-registrar=api

# bdapi：双向 API 注册
-registrar=bdapi

# dns：使用 DNS 注册
-registrar=dns

# bddns：双向 DNS 注册
-registrar=bddns

# amp：使用 AMP 缓存注册
-registrar=amp
```

### Conjure 传输方式

```bash
# min：最小传输
-transport=min

# prefix：前缀传输
-transport=prefix

# obfs4：obfs4 传输
-transport=obfs4
```

### Conjure 完整示例

```bash
# 使用 Conjure 连接
./cli -connect-addr=example.com:443 \
      -registrar=bdapi \
      -transport=min \
      -api-endpoint=https://registration.refraction.network/api/register-bidirectional

# 使用 AMP 缓存注册
./cli -connect-addr=example.com:443 \
      -registrar=amp \
      -ampCacheUrl=https://www-amproject.org/v0/cache \
      -transport=min
```

---

## 高级用法

### 1. 自定义 Decoy

```go
package main

import (
    "fmt"
    "github.com/refraction-networking/gotapdance/tapdance"
)

func main() {
    tapdance.AssetsSetDir("./assets/")
    
    // 指定单个 decoy
    // 格式："SNI,IP" 或 "SNI"
    conn, err := tapdance.Dial("tcp", "example.com:443")
    // 注：Dial 函数会自动处理 decoy 选择
    
    // 或者通过配置文件指定
    // 编辑 assets/ClientConf 文件
}
```

### 2. 错误处理和重试

```go
package main

import (
    "fmt"
    "time"
    "github.com/refraction-networking/gotapdance/tapdance"
)

func connectWithRetry(target string, maxRetries int) (tapdance.Conn, error) {
    tapdance.AssetsSetDir("./assets/")
    
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        conn, err := tapdance.Dial("tcp", target)
        if err == nil {
            return conn, nil
        }
        lastErr = err
        time.Sleep(time.Second * time.Duration(i+1))
    }
    return nil, lastErr
}
```

### 3. 连接池管理

```go
package main

import (
    "sync"
    "github.com/refraction-networking/gotapdance/tapdance"
)

type ConnectionPool struct {
    connections chan tapdance.Conn
    maxSize     int
    mutex       sync.Mutex
}

func NewConnectionPool(maxSize int) *ConnectionPool {
    return &ConnectionPool{
        connections: make(chan tapdance.Conn, maxSize),
        maxSize:     maxSize,
    }
}

func (p *ConnectionPool) Get() tapdance.Conn {
    select {
    case conn := <-p.connections:
        return conn
    default:
        return nil
    }
}

func (p *ConnectionPool) Put(conn tapdance.Conn) {
    select {
    case p.connections <- conn:
    default:
        conn.Close()
    }
}
```

### 4. 流量分析

使用 TLS 日志功能可以帮助分析流量：

```bash
# 保存 TLS 密钥
./cli -connect-addr=example.com:443 -tlslog=./ssl.log

# 然后用 Wireshark 解密流量
# Wireshark -> Preferences -> Protocols -> TLS -> (Pre)-Master-Secret log filename
```

---

## 注意事项与最佳实践

### ⚠️ 重要提示

#### 1. 合法使用

- 遵守当地法律法规
- 遵守网络使用政策
- 不得用于非法目的

#### 2. 性能和限制

- TapDance 可能比直接连接慢
- 可能需要多次尝试连接
- 某些网络环境可能无法使用

#### 3. 安全性

- 验证 Assets 文件的真实性
- 定期更新配置文件
- 保护 Assets 目录的权限

### 最佳实践

#### 1. 错误处理

```go
conn, err := tapdance.Dial("tcp", "example.com:443")
if err != nil {
    // 详细记录错误信息
    log.Printf("TapDance 连接失败: %+v", err)
    
    // 实现重试逻辑
    // 或返回错误给调用者
    return err
}
defer conn.Close()  // 确保关闭连接
```

#### 2. 超时设置

```go
// 设置连接超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// TapDance 的 Dial 函数目前不支持 context
// 建议在应用层实现超时控制
```

#### 3. 日志记录

```bash
# 调试时启用详细日志
./cli -debug
# 或
./cli -trace
```

```go
// 在代码中使用日志
import "github.com/sirupsen/logrus"

logrus.SetLevel(logrus.DebugLevel)
```

#### 4. 配置管理

- 将 Assets 目录放在安全位置
- 使用环境变量管理路径
- 定期更新 ClientConf

---

## 常见问题解答

### Q1: 如何获取 Assets 文件？

**A:** Assets 文件（ClientConf 和 roots）需要从 TapDance 项目获取。请参考项目的相关文档或联系维护者。

### Q2: TapDance 和 Tor 有什么区别？

**A:** 
- **TapDance**：端到中间代理，看起来像正常流量
- **Tor**：端到端加密，多层代理网络

两者设计目标不同，可以互补使用。

### Q3: 为什么连接失败？

**A:** 可能的原因：
1. Assets 文件缺失或过期
2. 网络环境不兼容
3. Decoy 不可用
4. 防火墙规则

建议启用详细日志排查。

### Q4: 如何更新配置？

**A:** 
1. 获取最新的 Assets 文件
2. 替换旧的 ClientConf 和 roots
3. 重启客户端

### Q5: TapDance 支持哪些平台？

**A:** 
- Windows
- macOS
- Linux
- Android（通过 gomobile）

### Q6: 如何禁用 IPv6？

**A:** 
```bash
./cli -disable-ipv6 -connect-addr=example.com:443
```

### Q7: 能否与 uTLS 配合使用？

**A:** GoTapDance 现在使用 Conjure 协议，而 Conjure 客户端可以配合 uTLS 使用。详细参考 Conjure 文档。

### Q8: TapDance 的性能如何？

**A:** 
- 比直接连接慢
- 延迟取决于 decoy 的选择
- 适合对速度要求不高的场景

### Q9: 如何贡献代码？

**A:** 
1. Fork 仓库
2. 创建特性分支
3. 提交 Pull Request

参考项目 GitHub 页面的贡献指南。

### Q10: TapDance 是否免费？

**A:** 是的，TapDance 是开源免费的反审查技术，基于 Apache-2.0 许可证。

---

## 参考资料

### 官方资源

- **GoTapDance GitHub：** https://github.com/refraction-networking/gotapdance
- **TapDance Station GitHub：** https://github.com/refraction-networking/tapdance
- **Refraction Networking：** https://refraction.network
- **GoDoc：** https://godoc.org/github.com/refraction-networking/gotapdance/tapdance

### 学术论文

- **TapDance 2014 论文：** ["TapDance: End-to-Middle Anticensorship without Flow Blocking"](https://ericw.us/trow/tapdance-sec14.pdf)
- **TapDance 2017 论文：** ["An ISP-Scale Deployment of TapDance"](https://www.usenix.org/system/files/conference/foci17/foci17-paper-frolov_0.pdf)

### 相关项目

- **Conjure：** https://github.com/refraction-networking/conjure
- **Psiphon：** https://psiphon.ca/（集成了 TapDance）
- **uTLS：** https://github.com/refraction-networking/utls
- **uQUIC：** https://github.com/refraction-networking/uquic

### 许可证

- **Apache-2.0 License：** https://github.com/refraction-networking/gotapdance/blob/master/LICENSE

---

## 更新日志

### v1.7.10
最新版本，包含 Conjure 协议支持和多项改进。

### 主要变更

**Conjure 协议集成**：
- 支持多种注册方式（decoy, api, bdapi, dns, bddns, amp）
- 支持多种传输方式（min, prefix, obfs4）
- 双向通信支持

**依赖更新**：
- Go 1.22+ 支持
- uTLS v1.6.7
- Conjure v0.9.0

**改进**：
- 更好的错误处理
- 性能优化
- 文档更新

### 历史版本

- v1.7.9, v1.7.8, v1.7.7... 历史版本请参考 [GitHub Releases](https://github.com/refraction-networking/gotapdance/releases)

---

## 贡献

如果你发现本文档有任何问题或有改进建议，欢迎提出 Issue 或 Pull Request。

### 联系方式

- **GitHub Issues：** https://github.com/refraction-networking/gotapdance/issues
- **Pull Requests：** https://github.com/refraction-networking/gotapdance/pulls

---

**最后更新：** 2025-01-10  
**文档版本：** 1.0.0  
**GoTapDance 版本：** v1.7.10（最新版本）  
**Go 版本要求：** 1.10+（推荐 1.22+）  
**许可证：** Apache-2.0
