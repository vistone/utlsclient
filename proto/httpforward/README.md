# HTTP 转发服务 Proto 定义

## 服务定义

### HTTPForwardService

提供 HTTP 请求转发功能，使用 uTLS 客户端访问远程资源。

#### RPC 方法

##### Handshake

客户端握手（首次连接，获取编码）。

**请求消息**: `HandshakeRequest`
- `client_ip` (string): 客户端 IP 地址（仅首次需要）

**响应消息**: `HandshakeResponse`
- `client_code` (int32): 分配的客户端编码（1,2,3,4...）

##### ForwardRequest

转发 HTTP 请求（使用 uTLS 客户端）。

**请求消息**: `ForwardRequestRequest`
- `client_id` (oneof):
  - `client_ip` (string): 客户端 IP（首次使用）
  - `client_code` (int32): 客户端编码（后续使用，节省流量）
- `hostname` (oneof):
  - `hostname_raw` (string): 原始主机名（首次使用）
  - `hostname_code` (int32): 主机名编码（后续使用，节省流量）
- `path` (string): 请求路径

**响应消息**: `ForwardRequestResponse`
- `client_code` (int32): 客户端编码（回显）
- `hostname_code` (int32): 主机名编码（回显）
- `path` (string): 路径（回显）
- `body` (bytes): 响应体（仅状态码 200 时返回，其他状态码为空）
- `status_code` (int32): HTTP 状态码

## 流量优化机制

### 1. 客户端编码机制
- **首次连接**: 客户端发送 IP 地址（例如：`"192.168.1.100"`）
- **服务器响应**: 分配编码（例如：`1`, `2`, `3`, `4`...）
- **后续请求**: 只需发送编码数字，无需传输完整的 IP 字符串

**流量节省**: 
- IP 地址: `192.168.1.100` = 15 字节
- 编码: `1` = 1 字节
- **节省**: ~93% 流量

### 2. 主机名编码机制
- **首次使用**: 发送完整主机名（例如：`"www.example.com"`）
- **服务器响应**: 返回主机名编码（例如：`1`, `2`, `3`...）
- **后续使用相同主机名**: 只需发送编码，无需重复传输主机名字符串

**流量节省**:
- 主机名: `www.example.com` = 17 字节
- 编码: `1` = 1 字节
- **节省**: ~94% 流量

### 3. Body 优化
- 状态码 **200**: 返回完整响应体
- 状态码 **非 200**: 返回空 body，节省流量

### 4. 响应头优化
- **已移除**: 不再传输响应头，进一步节省流量

### 使用流程示例

```
1. 客户端握手
   请求: { client_ip: "192.168.1.100" }
   响应: { client_code: 1 }

2. 首次请求（新主机名）
   请求: { client_code: 1, hostname_raw: "www.example.com", path: "/" }
   响应: { client_code: 1, hostname_code: 1, path: "/", status_code: 200, body: [...] }

3. 后续请求（相同主机名）
   请求: { client_code: 1, hostname_code: 1, path: "/api/data" }
   响应: { client_code: 1, hostname_code: 1, path: "/api/data", status_code: 200, body: [...] }
```

## 使用场景

1. **HTTP 代理转发**: 通过 gRPC 服务转发 HTTP 请求到远程服务器
2. **TLS 指纹模拟**: 使用 uTLS 模拟不同浏览器的 TLS 指纹
3. **流量优化**: 
   - 使用编码机制减少重复传输
   - 仅返回状态码 200 的响应体
   - 不传输响应头，节省带宽

## 使用示例

### 启动服务器

```bash
go run examples/httpforward_server_example.go -port=50052
```

### 客户端调用

```bash
go run examples/httpforward_client_example.go \
  -server=localhost:50052 \
  -client-ip=192.168.1.100 \
  -hostname=www.google.com \
  -path=/
```

## 特性

- ✅ 使用 uTLS 客户端进行指纹伪装
- ✅ 支持 HTTP/2 和 HTTP/1.1
- ✅ 自动处理 TLS 握手
- ✅ **流量优化**: 使用编码机制大幅减少传输数据量
- ✅ **智能编码**: 客户端 IP 和主机名使用简单数字编码（1,2,3,4...）
- ✅ **Body 优化**: 仅返回状态码 200 的响应体
