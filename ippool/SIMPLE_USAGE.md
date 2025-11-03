# IP 池库 - 简化使用指南

## 核心功能

这个库非常简单易用，只需要一行代码创建，然后就可以使用了：

```go
// 创建库（自动从本地加载，后台与服务器同步）
library := ippool.NewIPPoolLibrary("http://tile0.zeromaps.cn:9005", "./ippool_data")

// 立即可以使用（使用本地数据，不等待网络）
hosts := library.GetAllHosts()
pool, _ := library.GetIPPool("kh.google.com")

// 启动定时自动同步（热更新）
library.StartAutoSync(5 * time.Minute)
```

## 工作原理

### 1. 初始化（自动加载本地数据）

```go
library := ippool.NewIPPoolLibrary(baseURL, dataDir)
```

**自动执行：**
- ✅ 从本地文件加载所有数据到内存（快速启动）
- ✅ 后台尝试与服务器同步（网络不通也不影响）
- ✅ 如果服务器数据更新，自动更新本地文件和内存

### 2. 智能同步（只更新新数据）

```go
library.SyncAll()
```

**智能判断：**
- ✅ 比较服务器的 `last_updated` 和本地的 `last_updated`
- ✅ 如果服务器数据更新 → 下载并保存到本地，加载到内存
- ✅ 如果服务器数据没有更新 → 跳过（节省带宽）
- ✅ 如果网络不通 → 使用本地数据（不报错）

### 3. 定时自动同步（热更新）

```go
library.StartAutoSync(5 * time.Minute)
```

**自动执行：**
- ✅ 每 5 分钟自动检查服务器是否有新数据
- ✅ 有新数据时自动更新到本地文件和内存
- ✅ 热更新：更新后立即生效，无需重启

### 4. 离线模式（完全不依赖网络）

```go
library.SetOfflineMode(true)
```

**工作方式：**
- ✅ 只使用本地数据
- ✅ 所有网络操作都跳过
- ✅ 完全离线运行

## 完整示例

```go
package main

import (
    "fmt"
    "time"
    "utls_client/ippool"
)

func main() {
    // 创建库（自动加载本地数据，后台同步）
    library := ippool.NewIPPoolLibrary("", "./ippool_data")
    defer library.Close()

    // 立即可以使用（本地数据已加载）
    hosts := library.GetAllHosts()
    fmt.Printf("找到 %d 个主机\n", len(hosts))

    for _, host := range hosts {
        // 获取 IP 池数据（从内存读取，快速）
        pool, err := library.GetIPPool(host.Host)
        if err == nil {
            fmt.Printf("%s: %d 个 IPv4\n", host.Host, len(pool.IPv4))
        }

        // 获取详细数据（从内存读取，快速）
        detail, err := library.GetDetailIPPool(host.Host)
        if err == nil {
            fmt.Printf("%s: 最后更新 %s\n", host.Host, 
                detail.Stats.LastUpdated.Format(time.RFC3339))
        }
    }

    // 启动定时自动同步（热更新）
    library.StartAutoSync(5 * time.Minute)
    
    // 程序运行...
    time.Sleep(1 * time.Hour)
    
    // 停止自动同步
    library.StopAutoSync()
}
```

## 数据流程

```
启动
  ↓
从本地文件加载 → 内存（快速）
  ↓
后台尝试同步服务器
  ├─ 网络正常 → 检查 last_updated
  │   ├─ 服务器更新 → 下载 → 保存本地 → 加载内存（热更新）
  │   └─ 服务器未更新 → 跳过
  └─ 网络不通 → 使用本地数据（无影响）
  ↓
定时自动同步（热更新）
  ↓
持续运行...
```

## 关键特性

1. **快速启动**：从本地加载，不等待网络
2. **智能更新**：只下载新数据，节省带宽
3. **网络容错**：网络不通时使用本地数据
4. **热更新**：新数据自动加载到内存
5. **本地存储**：所有数据保存在本地文件
6. **离线运行**：可以完全离线使用

## 注意事项

- 首次运行需要网络连接（下载初始数据）
- 后续运行可以使用本地数据（快速启动）
- 定时同步会检查服务器是否有新数据
- 所有数据都保存在 `dataDir` 目录下



