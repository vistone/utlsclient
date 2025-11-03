# IP 池库使用指南

## 概述

这是一个独立的 IP 池管理和分析库，用于从 ZeroMaps 的 IP 池 API 获取、同步和分析 IP 地址数据。

### 功能特性

- ✅ **实时同步**：自动从服务器同步最新的 IP 池数据
- ✅ **双格式支持**：支持简化格式（仅 IP 列表）和详细格式（含地理位置信息）
- ✅ **自动同步**：支持定时自动同步更新
- ✅ **数据分析**：提供强大的数据分析功能
- ✅ **多条件搜索**：支持按国家、城市、ISP、数据中心等条件搜索
- ✅ **线程安全**：所有操作都是线程安全的

## 快速开始

### 基础使用

```go
package main

import (
    "fmt"
    "utls_client/ippool"
)

func main() {
    // 创建 IP 池库
    library := ippool.NewIPPoolLibrary("http://tile0.zeromaps.cn:9005")
    defer library.Close()

    // 同步主机列表
    if err := library.SyncHosts(); err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    // 获取所有主机
    hosts := library.GetAllHosts()
    fmt.Printf("找到 %d 个主机\n", len(hosts))

    // 同步指定主机的 IP 池数据
    if len(hosts) > 0 {
        host := hosts[0].Host
        
        // 同步简化格式
        if err := library.SyncIPPool(host); err != nil {
            fmt.Printf("同步失败: %v\n", err)
            return
        }

        // 获取 IP 列表
        pool, _ := library.GetIPPool(host)
        fmt.Printf("IPv4 数量: %d\n", len(pool.IPv4))
        fmt.Printf("IPv6 数量: %d\n", len(pool.IPv6))
    }
}
```

## API 文档

### IPPoolLibrary

#### 创建库实例

```go
library := ippool.NewIPPoolLibrary(baseURL string) *IPPoolLibrary
```

- `baseURL`: API 服务器地址（默认为 `http://tile0.zeromaps.cn:9005`）

#### 同步方法

```go
// 同步所有数据
err := library.SyncAll()

// 同步主机列表
err := library.SyncHosts()

// 同步指定主机的简化格式 IP 池
err := library.SyncIPPool(host string)

// 同步指定主机的详细格式 IP 池
err := library.SyncDetailIPPool(host string)
```

#### 查询方法

```go
// 获取所有主机列表
hosts := library.GetAllHosts() []HostInfo

// 获取指定主机信息
hostInfo, err := library.GetHostInfo(host string) (*HostInfo, error)

// 获取简化格式 IP 池
pool, err := library.GetIPPool(host string) (*IPPoolData, error)

// 获取详细格式 IP 池
detailPool, err := library.GetDetailIPPool(host string) (*DetailIPPoolData, error)

// 获取指定 IP 的详细信息
ipInfo, err := library.GetIPDetail(host, ip string) (*IPDetailInfo, error)
```

#### 自动同步

```go
// 启动自动同步（每 5 分钟一次）
err := library.StartAutoSync(0)

// 启动自动同步（自定义间隔）
err := library.StartAutoSync(10 * time.Minute)

// 停止自动同步
library.StopAutoSync()

// 检查自动同步状态
enabled := library.IsAutoSyncEnabled()
```

### Analyzer

#### 创建分析器

```go
analyzer := ippool.NewAnalyzer(library *IPPoolLibrary) *Analyzer
```

#### 分析方法

```go
// 分析所有数据
stats, err := analyzer.AnalyzeAll() (*AnalyzeStats, error)

// 分析指定主机
stats, err := analyzer.AnalyzeByHost(host string) (*AnalyzeStats, error)

// 按国家分析
ips, err := analyzer.AnalyzeByCountry(country string) ([]*IPDetailInfo, error)

// 按城市分析
ips, err := analyzer.AnalyzeByCity(city string) ([]*IPDetailInfo, error)

// 按 ISP 分析
ips, err := analyzer.AnalyzeByISP(isp string) ([]*IPDetailInfo, error)

// 按数据中心分析
ips, err := analyzer.AnalyzeByDataCenter(dataCenter string) ([]*IPDetailInfo, error)

// 搜索 IP（支持多条件）
ips, err := analyzer.SearchIPs(host, country, city, isp, dataCenter string) ([]*IPDetailInfo, error)
```

#### 实用方法

```go
// 获取随机 IP
ip, err := analyzer.GetRandomIP(host string) (string, error)

// 获取指定主机的所有 IP
ipv4, ipv6, err := analyzer.GetAllIPsByHost(host string) ([]string, []string, error)
```

## 数据结构

### HostInfo

```go
type HostInfo struct {
    Host        string  // 主机名
    FileName    string  // 简化格式文件名
    DetailFile  string  // 详细格式文件名
    URL         string  // 简化格式 URL
    DetailURL   string  // 详细格式 URL
    Exists      bool    // 简化格式是否存在
    DetailExists bool   // 详细格式是否存在
}
```

### IPPoolData

```go
type IPPoolData struct {
    IPv4 []string  // IPv4 地址列表
    IPv6 []string  // IPv6 地址列表（可选）
}
```

### IPDetailInfo

```go
type IPDetailInfo struct {
    IP       string          // IP 地址
    Location IPLocationInfo  // 地理位置信息
}

type IPLocationInfo struct {
    Country    string  // 国家
    Region     string  // 地区
    City       string  // 城市
    ISP        string  // ISP
    Org        string  // 组织
    DataCenter string  // 数据中心
    IPType     string  // IP 类型
}
```

### AnalyzeStats

```go
type AnalyzeStats struct {
    TotalHosts   int            // 主机总数
    TotalIPv4    int             // IPv4 总数
    TotalIPv6    int             // IPv6 总数
    Countries    map[string]int  // 国家分布
    Cities       map[string]int  // 城市分布
    Regions      map[string]int  // 地区分布
    ISPs         map[string]int  // ISP 分布
    Orgs         map[string]int  // 组织分布
    DataCenters  map[string]int  // 数据中心分布
    IPTypes      map[string]int  // IP 类型分布
}
```

## 使用示例

### 示例 1：基本同步和查询

```go
library := ippool.NewIPPoolLibrary("")
library.SyncHosts()
hosts := library.GetAllHosts()

for _, host := range hosts {
    library.SyncIPPool(host.Host)
    pool, _ := library.GetIPPool(host.Host)
    fmt.Printf("%s: %d 个 IPv4\n", host.Host, len(pool.IPv4))
}
```

### 示例 2：使用分析器

```go
library := ippool.NewIPPoolLibrary("")
library.SyncHosts()
library.SyncDetailIPPool("kh.google.com")

analyzer := ippool.NewAnalyzer(library)
stats, _ := analyzer.AnalyzeByHost("kh.google.com")

fmt.Printf("国家分布:\n")
for country, count := range stats.Countries {
    fmt.Printf("  %s: %d\n", country, count)
}
```

### 示例 3：自动同步

```go
library := ippool.NewIPPoolLibrary("")

// 启动自动同步，每 5 分钟同步一次
library.StartAutoSync(5 * time.Minute)

// 程序运行...

// 停止自动同步
defer library.StopAutoSync()
```

### 示例 4：搜索功能

```go
analyzer := ippool.NewAnalyzer(library)

// 搜索所有 Google LLC 的 IP
ips, _ := analyzer.SearchIPs("", "", "", "Google LLC", "")

// 搜索指定主机中特定国家的 IP
ips, _ := analyzer.SearchIPs("kh.google.com", "United States", "", "", "")

for _, ip := range ips {
    fmt.Printf("%s: %s, %s\n", ip.IP, ip.Location.City, ip.Location.Country)
}
```

## 注意事项

1. **网络超时**：默认超时时间为 60 秒，如果网络较慢可能需要等待
2. **数据缓存**：所有数据都在内存中缓存，大量数据可能占用较多内存
3. **线程安全**：所有方法都是线程安全的，可以在多个 goroutine 中使用
4. **自动同步**：自动同步会在后台运行，记得在程序退出时调用 `StopAutoSync()`
5. **错误处理**：建议对每个方法都进行错误检查

## 完整示例

参考 `examples/ippool_example.go` 查看完整的使用示例。
