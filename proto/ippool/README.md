# IP 池 gRPC 服务 Proto 定义

## 文件说明

- `ippool.proto` - IP 池服务的 gRPC 接口定义

## 生成 Go 代码

### 1. 安装依赖

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 2. 生成代码

```bash
# 从项目根目录执行
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/ippool/ippool.proto
```

### 3. 安装 gRPC 依赖

```bash
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

## 服务接口说明

### 查询接口

- `GetAllHosts` - 获取所有主机列表
- `GetHostInfo` - 获取指定主机信息
- `GetIPPool` - 获取简化格式 IP 池数据
- `GetDetailIPPool` - 获取详细格式 IP 池数据
- `GetIPDetail` - 获取指定 IP 的详细信息

### 搜索和分析接口

- `SearchIPs` - 多条件搜索 IP
- `GetRandomIP` - 获取随机 IP
- `GetAllIPsByHost` - 获取指定主机的所有 IP
- `AnalyzeAll` - 分析所有数据
- `AnalyzeByHost` - 分析指定主机
- `AnalyzeByCountry` - 按国家分析
- `AnalyzeByCity` - 按城市分析
- `AnalyzeByISP` - 按 ISP 分析
- `AnalyzeByDataCenter` - 按数据中心分析

### 同步接口

- `SyncAll` - 同步所有数据
- `SyncHosts` - 同步主机列表
- `SyncIPPool` - 同步简化格式 IP 池
- `SyncDetailIPPool` - 同步详细格式 IP 池

### 状态接口

- `GetServiceStatus` - 获取服务状态


