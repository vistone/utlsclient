# IP 池 gRPC 服务器

## 概述

这是 IP 池服务的 gRPC 服务器实现，提供了完整的 IP 池数据查询、分析和同步功能。

## 快速开始

### 启动服务器

```bash
cd examples
go run grpc_server_example.go
```

### 命令行参数

```bash
go run grpc_server_example.go \
  -port=50051 \                    # gRPC 服务端口（默认：50051）
  -base-url=http://tile0.zeromaps.cn:9005 \  # API 基础地址
  -data-dir=./ippool_data \        # 本地数据目录
  -auto-sync=true \                # 启用自动同步
  -sync-interval=5m                # 同步间隔（默认：5分钟）
```

## 服务接口

### 查询接口

- `GetAllHosts` - 获取所有主机列表
- `GetHostInfo` - 获取指定主机信息
- `GetIPPool` - 获取简化格式 IP 池数据
- `GetDetailIPPool` - 获取详细格式 IP 池数据
- `GetIPDetail` - 获取指定 IP 的详细信息

### 搜索和分析接口

- `SearchIPs` - 多条件搜索 IP（支持按国家、城市、ISP、数据中心）
- `GetRandomIP` - 获取随机 IP
- `GetAllIPsByHost` - 获取指定主机的所有 IP
- `AnalyzeAll` - 分析所有数据
- `AnalyzeByHost` - 分析指定主机
- `AnalyzeByCountry/City/ISP/DataCenter` - 按条件分析

### 同步接口

- `SyncAll` - 同步所有数据
- `SyncHosts` - 同步主机列表
- `SyncIPPool` - 同步简化格式 IP 池
- `SyncDetailIPPool` - 同步详细格式 IP 池（支持强制更新）

### 状态接口

- `GetServiceStatus` - 获取服务状态（离线模式、自动同步状态、最后同步时间等）

## 使用客户端

### Go 客户端示例

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    pb "utls_client/proto/ippool"
)

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer conn.Close()
    
    client := pb.NewIPPoolServiceClient(conn)
    
    // 获取所有主机
    hosts, err := client.GetAllHosts(context.Background(), &pb.GetAllHostsRequest{})
    if err != nil {
        log.Fatalf("获取主机列表失败: %v", err)
    }
    
    log.Printf("找到 %d 个主机", len(hosts.Hosts))
}
```

## 特性

- ✅ 完整的 gRPC 接口实现
- ✅ 自动数据同步
- ✅ 离线模式支持
- ✅ 本地数据缓存
- ✅ 线程安全
- ✅ 优雅关闭


