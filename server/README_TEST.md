# IP 池 gRPC 服务器测试说明

## 测试文件

### 单元测试

`ippool_server_test.go` - 包含所有 gRPC 服务器方法的单元测试

#### 测试用例

- `TestGetAllHosts` - 测试获取所有主机列表
- `TestGetHostInfo` - 测试获取指定主机信息
- `TestGetHostInfoNotFound` - 测试获取不存在的主机（错误处理）
- `TestGetIPPool` - 测试获取简化格式 IP 池
- `TestGetDetailIPPool` - 测试获取详细格式 IP 池
- `TestGetIPDetail` - 测试获取指定 IP 的详细信息
- `TestSearchIPs` - 测试搜索 IP（多条件）
- `TestGetRandomIP` - 测试获取随机 IP
- `TestAnalyzeAll` - 测试分析所有数据
- `TestGetServiceStatus` - 测试获取服务状态
- `TestSyncHosts` - 测试同步主机列表（需要网络连接）

#### 性能测试

- `BenchmarkGetAllHosts` - 性能测试：获取所有主机
- `BenchmarkGetHostInfo` - 性能测试：获取主机信息

### 集成测试

`integration_test.go` - 集成测试，需要构建标签 `integration`

#### 运行集成测试

```bash
# 运行集成测试（需要设置环境变量）
INTEGRATION_TEST=true go test -tags=integration ./server

# 或使用 go test 命令
go test -tags=integration -run TestIntegrationServer ./server
```

#### 测试用例

- `TestIntegrationServer` - 启动真实的 gRPC 服务器并测试客户端连接
- `TestConcurrentRequests` - 测试并发请求处理

## 运行测试

### 运行所有单元测试

```bash
go test ./server -v
```

### 运行特定测试

```bash
# 运行单个测试
go test ./server -v -run TestGetAllHosts

# 运行性能测试
go test ./server -bench=.
```

### 跳过需要网络的测试

```bash
# 使用 -short 标志跳过需要网络的测试
go test ./server -short -v
```

### 运行集成测试

```bash
# 需要设置环境变量
INTEGRATION_TEST=true go test -tags=integration ./server

# 或者不设置环境变量（会被跳过）
go test -tags=integration ./server
```

### 测试覆盖率

```bash
# 生成覆盖率报告
go test ./server -cover

# 生成详细的覆盖率报告
go test ./server -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 测试注意事项

1. **离线模式**: 单元测试使用离线模式，只从本地文件读取数据，不进行网络请求
2. **测试数据**: 测试需要本地有 `ippool_data` 目录和相应的 JSON 文件
3. **集成测试**: 集成测试会启动真实的 gRPC 服务器，需要设置 `INTEGRATION_TEST=true`
4. **网络测试**: 某些测试（如 `TestSyncHosts`）需要网络连接，如果没有网络会被跳过

## 测试数据准备

测试会自动使用 `./testdata/ippool_test` 目录。如果有本地数据文件，测试会更完整。

```bash
# 确保有测试数据目录
mkdir -p server/testdata/ippool_test

# 如果有本地数据，可以复制过来
cp -r ippool_data/* server/testdata/ippool_test/ 2>/dev/null || true
```


