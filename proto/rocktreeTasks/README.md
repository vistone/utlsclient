# RockTree 任务服务 Proto 定义

## 服务定义

### RockTreeTaskService

提供 RockTree 任务处理服务，使用 uTLS 客户端访问 Google Tile 服务。

#### RPC 方法

##### ProcessTask

处理任务请求，支持批量元数据和节点数据两种类型。

**请求消息**: `TaskRequest`
- `client_id` (string): 客户端 ID
- `type` (Type): 任务类型（BULK_METADATA 或 NODE_DATA）
- `tilekey` (string): 瓦片键
- `epoch` (int32): 纪元
- `imagery_epoch` (int32): 图像纪元（可选，允许为空）

**响应消息**: `TaskResponse`
- `client_id` (string): 客户端 ID（回显）
- `type` (Type): 任务类型（回显）
- `tilekey` (string): 瓦片键（回显）
- `epoch` (int32): 纪元（回显）
- `imagery_epoch` (int32): 图像纪元（回显，允许为空）
- `body` (bytes): 响应体（仅状态码 200 时返回，其他状态码为空）
- `status_code` (int32): HTTP 状态码

## 任务类型

### Type 枚举

- `BULK_METADATA = 0`: 批量元数据
- `NODE_DATA = 1`: 节点数据

## URL 构建规则

### 批量元数据 (BULK_METADATA)
```
https://tile.googleapis.com/tile/v1/bulkmetadata?tilekey={tilekey}&epoch={epoch}[&imagery_epoch={imagery_epoch}]
```

### 节点数据 (NODE_DATA)
```
https://tile.googleapis.com/tile/v1/nodedata?tilekey={tilekey}&epoch={epoch}[&imagery_epoch={imagery_epoch}]
```

## 使用示例

### 启动服务器

```bash
go run examples/rocktree_tasks_server_example.go -port=50053
```

### 客户端调用

#### 批量元数据请求
```bash
go run examples/rocktree_tasks_client_example.go \
  -server=localhost:50053 \
  -client-id=my-client-001 \
  -type=BULK_METADATA \
  -tilekey=t:0:0:0 \
  -epoch=1
```

#### 节点数据请求（带图像纪元）
```bash
go run examples/rocktree_tasks_client_example.go \
  -server=localhost:50053 \
  -client-id=my-client-001 \
  -type=NODE_DATA \
  -tilekey=t:1:2:3 \
  -epoch=1 \
  -imagery-epoch=123
```

## 特性

- ✅ 使用 uTLS 客户端进行指纹伪装
- ✅ 支持 HTTP/2 和 HTTP/1.1
- ✅ 自动处理 TLS 握手
- ✅ 流量优化：仅返回状态码 200 的响应体
- ✅ 支持两种任务类型：批量元数据和节点数据
- ✅ 可选图像纪元参数


