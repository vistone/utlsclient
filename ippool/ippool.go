package ippool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	clientLib "utls_client/lib"
)

// IPPoolLibrary IP 池库管理器
type IPPoolLibrary struct {
	// API 基础地址
	baseURL string

	// uTLS 客户端（使用项目的客户端）
	client *clientLib.Client

	// 本地存储目录
	dataDir string

	// 是否启用离线模式（只从本地文件读取，不联网）
	offlineMode   bool
	offlineModeMu sync.RWMutex

	// 主机列表缓存
	hosts   []HostInfo
	hostsMu sync.RWMutex

	// IP 池数据缓存 (host -> IPPoolData)
	ipPools   map[string]*IPPoolData
	ipPoolsMu sync.RWMutex

	// 详细 IP 数据缓存 (host -> DetailIPPoolData)
	detailPools   map[string]*DetailIPPoolData
	detailPoolsMu sync.RWMutex

	// 最后同步时间
	lastSyncTime   time.Time
	lastSyncTimeMu sync.RWMutex

	// 各主机数据的最后更新时间 (host -> last_updated from server)
	hostLastUpdated   map[string]time.Time
	hostLastUpdatedMu sync.RWMutex

	// 同步间隔
	syncInterval time.Duration

	// 自动同步控制
	autoSyncEnabled bool
	autoSyncMu      sync.RWMutex
	syncTicker      *time.Ticker
	syncStopCh      chan struct{}

	// 白名单/黑名单（内存态）
	// 默认所有IP视为白名单；仅在收到403时将该IP加入黑名单；收到200时将其移出黑名单
	banMu        sync.RWMutex
	blacklistIPs map[string]map[string]struct{} // host -> set(ip)
}

// HostInfo 主机信息
type HostInfo struct {
	Host         string `json:"host"`
	FileName     string `json:"file_name"`
	DetailFile   string `json:"detail_file"`
	URL          string `json:"url"`
	DetailURL    string `json:"detail_url"`
	Exists       bool   `json:"exists"`
	DetailExists bool   `json:"detail_exists"`
}

// IPPoolResponse API 响应结构
type IPPoolResponse struct {
	Hosts []HostInfo `json:"hosts"`
	Usage string     `json:"usage"`
}

// IPPoolData 简化格式的 IP 池数据
type IPPoolData struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6,omitempty"`
}

// IPLocationInfo IP 地理位置信息
type IPLocationInfo struct {
	Country    string `json:"country"`
	Region     string `json:"region"`
	City       string `json:"city"`
	ISP        string `json:"isp"`
	Org        string `json:"org"`
	DataCenter string `json:"data_center"`
	IPType     string `json:"ip_type"`
}

// IPDetailInfo 详细 IP 信息
type IPDetailInfo struct {
	IP       string         `json:"ip"`
	Location IPLocationInfo `json:"location"`
}

// DetailIPPoolData 详细格式的 IP 池数据
type DetailIPPoolData struct {
	IPs   map[string]*IPDetailInfo `json:"-"` // IP -> 详细信息
	Stats PoolStats                `json:"stats"`
}

// PoolStats IP 池统计信息
type PoolStats struct {
	IPv4Count   int       `json:"ipv4_count"`
	IPv6Count   int       `json:"ipv6_count"`
	LastUpdated time.Time `json:"last_updated"`
}

// NewIPPoolLibrary 创建新的 IP 池库
// baseURL: API 服务器地址，如果为空则使用默认值
// dataDir: 本地数据存储目录，如果为空则使用默认值 "./ippool_data"
func NewIPPoolLibrary(baseURL, dataDir string) *IPPoolLibrary {
	if baseURL == "" {
		baseURL = "http://tile0.zeromaps.cn:9005"
	}
	if dataDir == "" {
		dataDir = "./ippool_data"
	}

	// 确保数据目录存在
	os.MkdirAll(dataDir, 0755)

	// 创建客户端，设置更长的超时时间
	config := &clientLib.Config{
		Timeout: 60 * time.Second, // 60秒超时
	}
	client := clientLib.NewClient(nil, config)

	lib := &IPPoolLibrary{
		baseURL:         baseURL,
		client:          client,
		dataDir:         dataDir,
		hosts:           make([]HostInfo, 0),
		ipPools:         make(map[string]*IPPoolData),
		detailPools:     make(map[string]*DetailIPPoolData),
		hostLastUpdated: make(map[string]time.Time),
		syncInterval:    5 * time.Minute, // 默认5分钟同步一次
		autoSyncEnabled: false,
		syncStopCh:      make(chan struct{}),
		blacklistIPs:    make(map[string]map[string]struct{}),
	}

	// 1. 先从本地加载数据（快速启动，不依赖网络）
	lib.LoadFromLocal()

	// 2. 不再在启动时自动同步，改为由用户通过定时任务或手动调用 SyncAll() 来控制
	// 如果需要定时同步，请调用：library.StartAutoSync(5 * time.Minute)

	return lib
}

// SetOfflineMode 设置离线模式
func (lib *IPPoolLibrary) SetOfflineMode(offline bool) {
	lib.offlineModeMu.Lock()
	lib.offlineMode = offline
	lib.offlineModeMu.Unlock()
}

// IsOfflineMode 检查是否处于离线模式
func (lib *IPPoolLibrary) IsOfflineMode() bool {
	lib.offlineModeMu.RLock()
	defer lib.offlineModeMu.RUnlock()
	return lib.offlineMode
}

// StartAutoSync 启动自动同步
func (lib *IPPoolLibrary) StartAutoSync(interval time.Duration) error {
	if interval <= 0 {
		interval = lib.syncInterval
	}

	lib.autoSyncMu.Lock()
	defer lib.autoSyncMu.Unlock()

	if lib.autoSyncEnabled {
		return fmt.Errorf("自动同步已启用")
	}

	lib.syncInterval = interval
	lib.autoSyncEnabled = true
	lib.syncTicker = time.NewTicker(interval)

	go lib.autoSyncLoop()

	return nil
}

// StopAutoSync 停止自动同步
func (lib *IPPoolLibrary) StopAutoSync() {
	lib.autoSyncMu.Lock()
	defer lib.autoSyncMu.Unlock()

	if !lib.autoSyncEnabled {
		return
	}

	lib.autoSyncEnabled = false
	if lib.syncTicker != nil {
		lib.syncTicker.Stop()
	}
	close(lib.syncStopCh)
	lib.syncStopCh = make(chan struct{})
}

// autoSyncLoop 自动同步循环（定时热更新）
func (lib *IPPoolLibrary) autoSyncLoop() {
	// 定时执行同步（热更新：服务器有新数据时自动更新到内存和本地文件）
	for {
		select {
		case <-lib.syncTicker.C:
			// 智能同步：只更新服务器数据比本地新的
			lib.SyncAll()
		case <-lib.syncStopCh:
			return
		}
	}
}

// SyncAll 同步所有数据（智能同步：服务器数据更新时才更新本地）
func (lib *IPPoolLibrary) SyncAll() error {
	// 如果处于离线模式，跳过同步
	if lib.IsOfflineMode() {
		return nil
	}

	// 1. 同步主机列表（如果失败，使用本地数据）
	if err := lib.SyncHosts(); err != nil {
		// 网络不通，使用本地数据（已经在 LoadFromLocal 中加载）
		return nil
	}

	// 2. 同步所有主机的 IP 池数据
	hosts := lib.GetAllHosts()

	// 使用 WaitGroup 确保所有同步完成，但设置超时避免长时间阻塞
	var wg sync.WaitGroup
	done := make(chan struct{})

	// 限制并发数，避免同时发起过多请求导致资源耗尽和网络拥塞
	const maxSyncConcurrency = 10 // 最多同时同步 10 个主机
	semaphore := make(chan struct{}, maxSyncConcurrency)

	for _, host := range hosts {
		// 同步简化格式
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量
		go func(h HostInfo) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量
			_ = lib.SyncIPPool(h.Host)
		}(host)

		// 同步详细格式（智能更新：只更新服务器数据比本地新的）
		if host.DetailExists {
			wg.Add(1)
			semaphore <- struct{}{} // 获取信号量
			go func(h HostInfo) {
				defer wg.Done()
				defer func() { <-semaphore }()          // 释放信号量
				_ = lib.SyncDetailIPPool(h.Host, false) // 使用智能更新判断
			}(host)
		}
	}

	// 等待所有同步完成，但设置超时
	go func() {
		wg.Wait()
		close(done)
	}()

	// 设置最大等待时间（60秒），避免长时间阻塞
	select {
	case <-done:
		// 所有同步完成
	case <-time.After(60 * time.Second):
		// 超时，不再等待（后台继续执行）
	}

	// 更新最后同步时间
	lib.lastSyncTimeMu.Lock()
	lib.lastSyncTime = time.Now()
	lib.lastSyncTimeMu.Unlock()

	return nil
}

// SyncHosts 同步主机列表
func (lib *IPPoolLibrary) SyncHosts() error {
	// 如果处于离线模式，跳过同步
	if lib.IsOfflineMode() {
		return fmt.Errorf("离线模式，无法同步")
	}

	url := fmt.Sprintf("%s/api/ipPool/", lib.baseURL)

	resp, err := lib.client.Get(url, map[string]string{
		"Accept": "application/json",
	})
	if err != nil {
		return fmt.Errorf("获取主机列表失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP 状态码: %d", resp.StatusCode)
	}

	var apiResp IPPoolResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	lib.hostsMu.Lock()
	lib.hosts = apiResp.Hosts
	lib.hostsMu.Unlock()

	// 保存到本地文件
	lib.saveHostsToLocal()

	return nil
}

// SyncIPPool 同步指定主机的 IP 池数据（简化格式）
func (lib *IPPoolLibrary) SyncIPPool(host string) error {
	hostInfo, err := lib.GetHostInfo(host)
	if err != nil {
		return err
	}

	if !hostInfo.Exists {
		return fmt.Errorf("主机 %s 的简化数据不存在", host)
	}

	url := fmt.Sprintf("%s%s", lib.baseURL, hostInfo.URL)

	resp, err := lib.client.Get(url, map[string]string{
		"Accept": "application/json",
	})
	if err != nil {
		return fmt.Errorf("获取 IP 池数据失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP 状态码: %d", resp.StatusCode)
	}

	// 直接保存原始 JSON 到本地文件（保持服务器格式）
	fileName := sanitizeFileName(host) + ".json"
	filePath := filepath.Join(lib.dataDir, fileName)
	if err := os.WriteFile(filePath, resp.Body, 0644); err != nil {
		// 保存失败不影响继续处理
		_ = err
	}

	var poolData IPPoolData
	if err := json.Unmarshal(resp.Body, &poolData); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	lib.ipPoolsMu.Lock()
	lib.ipPools[host] = &poolData
	lib.ipPoolsMu.Unlock()

	// 注意：原始 JSON 已经在上面直接保存了，不需要再次保存

	return nil
}

// SyncDetailIPPool 同步指定主机的详细 IP 池数据
// 智能更新：如果服务器数据比本地新，才更新；否则跳过
// 网络不通时返回错误，调用者可以使用本地数据
func (lib *IPPoolLibrary) SyncDetailIPPool(host string, force ...bool) error {
	shouldForce := false
	if len(force) > 0 {
		shouldForce = force[0]
	}

	// 如果处于离线模式，跳过
	if lib.IsOfflineMode() {
		return nil
	}

	hostInfo, err := lib.GetHostInfo(host)
	if err != nil {
		return err
	}

	if !hostInfo.DetailExists {
		return fmt.Errorf("主机 %s 的详细数据不存在", host)
	}

	// 智能更新判断：检查是否需要更新（基于 last_updated）
	if !shouldForce {
		if !lib.shouldUpdateDetailPool(host) {
			// 服务器数据没有本地新，或网络不通，跳过更新（使用本地数据）
			return nil
		}
	}

	url := fmt.Sprintf("%s%s", lib.baseURL, hostInfo.DetailURL)

	resp, err := lib.client.Get(url, map[string]string{
		"Accept": "application/json",
	})
	if err != nil {
		return fmt.Errorf("获取详细 IP 池数据失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP 状态码: %d", resp.StatusCode)
	}

	// 直接保存原始 JSON 到本地文件（保持服务器格式）
	fileName := sanitizeFileName(host) + "_detail.json"
	filePath := filepath.Join(lib.dataDir, fileName)
	if err := os.WriteFile(filePath, resp.Body, 0644); err != nil {
		// 保存失败不影响继续处理
		_ = err
	}

	// 解析 JSON（动态结构）
	var rawData map[string]interface{}
	if err := json.Unmarshal(resp.Body, &rawData); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 构建详细数据（用于内存缓存）
	detailData := &DetailIPPoolData{
		IPs: make(map[string]*IPDetailInfo),
	}

	// 提取统计信息
	var serverLastUpdated time.Time
	if stats, ok := rawData["stats"].(map[string]interface{}); ok {
		if count, ok := stats["ipv4_count"].(float64); ok {
			detailData.Stats.IPv4Count = int(count)
		}
		if count, ok := stats["ipv6_count"].(float64); ok {
			detailData.Stats.IPv6Count = int(count)
		}
		if updated, ok := stats["last_updated"].(string); ok {
			if t, err := time.Parse(time.RFC3339, updated); err == nil {
				detailData.Stats.LastUpdated = t
				serverLastUpdated = t
			}
		}
	}

	// 提取 IP 详细信息
	for key, value := range rawData {
		if key == "stats" {
			continue
		}

		if ipData, ok := value.(map[string]interface{}); ok {
			ipInfo := &IPDetailInfo{}

			if ip, ok := ipData["ip"].(string); ok {
				ipInfo.IP = ip
			}

			if locData, ok := ipData["location"].(map[string]interface{}); ok {
				if country, ok := locData["country"].(string); ok {
					ipInfo.Location.Country = country
				}
				if region, ok := locData["region"].(string); ok {
					ipInfo.Location.Region = region
				}
				if city, ok := locData["city"].(string); ok {
					ipInfo.Location.City = city
				}
				if isp, ok := locData["isp"].(string); ok {
					ipInfo.Location.ISP = isp
				}
				if org, ok := locData["org"].(string); ok {
					ipInfo.Location.Org = org
				}
				if dc, ok := locData["data_center"].(string); ok {
					ipInfo.Location.DataCenter = dc
				}
				if ipType, ok := locData["ip_type"].(string); ok {
					ipInfo.Location.IPType = ipType
				}
			}

			detailData.IPs[key] = ipInfo
		}
	}

	// 加载到内存（热更新）
	lib.detailPoolsMu.Lock()
	lib.detailPools[host] = detailData
	lib.detailPoolsMu.Unlock()

	// 更新该主机的最后更新时间（使用从服务器获取的 last_updated）
	if !serverLastUpdated.IsZero() {
		lib.hostLastUpdatedMu.Lock()
		lib.hostLastUpdated[host] = serverLastUpdated
		lib.hostLastUpdatedMu.Unlock()
	} else if !detailData.Stats.LastUpdated.IsZero() {
		// 如果 serverLastUpdated 为空，使用 detailData 中的时间
		lib.hostLastUpdatedMu.Lock()
		lib.hostLastUpdated[host] = detailData.Stats.LastUpdated
		lib.hostLastUpdatedMu.Unlock()
	}

	// 注意：原始 JSON 已经在上面直接保存到本地文件了

	return nil
}

// shouldUpdateDetailPool 判断是否需要更新详细 IP 池数据
// 优化：如果本地数据很新（1小时内），直接跳过检查（避免慢速网络请求）
func (lib *IPPoolLibrary) shouldUpdateDetailPool(host string) bool {
	// 检查本地缓存的详细数据是否存在
	lib.detailPoolsMu.RLock()
	detailPool, hasDetailData := lib.detailPools[host]
	lib.detailPoolsMu.RUnlock()

	// 如果没有详细数据，需要更新
	if !hasDetailData {
		return true
	}

	// 获取本地缓存的 last_updated
	localLastUpdated := detailPool.Stats.LastUpdated

	// 如果本地时间无效，需要更新
	if localLastUpdated.IsZero() {
		return true
	}

	// 优化：如果本地数据很新（6小时内），直接跳过检查（避免慢速网络请求）
	// 这样可以大幅减少网络请求，提高同步速度
	if time.Since(localLastUpdated) < 6*time.Hour {
		return false
	}

	// 本地数据较旧，检查服务器是否有更新
	// 使用带超时的上下文，避免长时间阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 注意：这个方法需要下载完整 JSON，如果网络慢会比较慢
	// 但通过超时控制，避免单个请求阻塞太久
	done := make(chan bool, 1)
	var serverLastUpdated time.Time
	var err error

	go func() {
		serverLastUpdated, err = lib.getServerLastUpdated(host)
		done <- true
	}()

	select {
	case <-done:
		// 完成检查
	case <-ctx.Done():
		// 超时，使用本地数据，不更新
		return false
	}

	if err != nil {
		// 如果获取失败（网络不通或超时），使用本地数据，不更新
		return false
	}

	// 如果服务器时间无效，不需要更新
	if serverLastUpdated.IsZero() {
		return false
	}

	// 比较服务器和本地的 last_updated，如果服务器更新，则需要更新
	return serverLastUpdated.After(localLastUpdated)
}

// getServerLastUpdated 获取服务器上指定主机的 last_updated 时间
// 这个方法只读取 JSON 的 stats 部分来判断，避免下载完整数据
// 注意：这个方法会发起网络请求，如果网络慢可能会阻塞
func (lib *IPPoolLibrary) getServerLastUpdated(host string) (time.Time, error) {
	// 如果处于离线模式，从本地文件读取
	if lib.IsOfflineMode() {
		lib.detailPoolsMu.RLock()
		detailPool, ok := lib.detailPools[host]
		lib.detailPoolsMu.RUnlock()
		if ok && !detailPool.Stats.LastUpdated.IsZero() {
			return detailPool.Stats.LastUpdated, nil
		}
		return time.Time{}, fmt.Errorf("离线模式且本地无数据")
	}

	hostInfo, err := lib.GetHostInfo(host)
	if err != nil {
		return time.Time{}, err
	}

	if !hostInfo.DetailExists {
		return time.Time{}, fmt.Errorf("主机 %s 的详细数据不存在", host)
	}

	url := fmt.Sprintf("%s%s", lib.baseURL, hostInfo.DetailURL)

	// 使用带超时的请求（缩短超时时间，快速失败）
	resp, err := lib.client.Get(url, map[string]string{
		"Accept": "application/json",
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("获取数据失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return time.Time{}, fmt.Errorf("HTTP 状态码: %d", resp.StatusCode)
	}

	// 解析 JSON，只读取 stats 部分的 last_updated
	var rawData map[string]interface{}
	if err := json.Unmarshal(resp.Body, &rawData); err != nil {
		return time.Time{}, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 提取 stats 中的 last_updated
	if stats, ok := rawData["stats"].(map[string]interface{}); ok {
		if updated, ok := stats["last_updated"].(string); ok {
			if t, err := time.Parse(time.RFC3339, updated); err == nil {
				return t, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("无法解析 last_updated")
}

// GetAllHosts 获取所有主机列表
func (lib *IPPoolLibrary) GetAllHosts() []HostInfo {
	lib.hostsMu.RLock()
	defer lib.hostsMu.RUnlock()

	result := make([]HostInfo, len(lib.hosts))
	copy(result, lib.hosts)
	return result
}

// GetHostInfo 获取指定主机信息
func (lib *IPPoolLibrary) GetHostInfo(host string) (*HostInfo, error) {
	lib.hostsMu.RLock()
	defer lib.hostsMu.RUnlock()

	for _, h := range lib.hosts {
		if h.Host == host {
			return &h, nil
		}
	}

	return nil, fmt.Errorf("未找到主机: %s", host)
}

// GetIPPool 获取指定主机的 IP 池数据（简化格式）
func (lib *IPPoolLibrary) GetIPPool(host string) (*IPPoolData, error) {
	lib.ipPoolsMu.RLock()
	defer lib.ipPoolsMu.RUnlock()

	pool, ok := lib.ipPools[host]
	if !ok {
		return nil, fmt.Errorf("未找到主机 %s 的 IP 池数据，请先调用 SyncIPPool", host)
	}

	return pool, nil
}

// GetDetailIPPool 获取指定主机的详细 IP 池数据
func (lib *IPPoolLibrary) GetDetailIPPool(host string) (*DetailIPPoolData, error) {
	lib.detailPoolsMu.RLock()
	defer lib.detailPoolsMu.RUnlock()

	pool, ok := lib.detailPools[host]
	if !ok {
		return nil, fmt.Errorf("未找到主机 %s 的详细 IP 池数据，请先调用 SyncDetailIPPool", host)
	}

	return pool, nil
}

// GetIPDetail 获取指定 IP 的详细信息
func (lib *IPPoolLibrary) GetIPDetail(host, ip string) (*IPDetailInfo, error) {
	detailPool, err := lib.GetDetailIPPool(host)
	if err != nil {
		return nil, err
	}

	ipInfo, ok := detailPool.IPs[ip]
	if !ok {
		return nil, fmt.Errorf("未找到 IP: %s", ip)
	}

	return ipInfo, nil
}

// ===== 白名单/黑名单：内存管理 =====

// IsAllowed 判断某个 IP（在指定 host 下）是否允许（不在黑名单）
func (lib *IPPoolLibrary) IsAllowed(host, ip string) bool {
	lib.banMu.RLock()
	defer lib.banMu.RUnlock()
	if m, ok := lib.blacklistIPs[host]; ok {
		if _, banned := m[ip]; banned {
			return false
		}
	}
	return true // 默认白名单
}

// ReportStatus 根据请求返回码更新白/黑名单：403 -> 加入黑名单；200 -> 移出黑名单
func (lib *IPPoolLibrary) ReportStatus(host, ip string, statusCode int) {
	lib.banMu.Lock()
	defer lib.banMu.Unlock()
	switch statusCode {
	case 403:
		if lib.blacklistIPs[host] == nil {
			lib.blacklistIPs[host] = make(map[string]struct{})
		}
		lib.blacklistIPs[host][ip] = struct{}{}
	case 200:
		if m := lib.blacklistIPs[host]; m != nil {
			delete(m, ip)
			if len(m) == 0 {
				delete(lib.blacklistIPs, host)
			}
		}
	}
}

// FilterIPs 过滤掉黑名单中的 IP（保留顺序）
func (lib *IPPoolLibrary) FilterIPs(host string, ips []string) []string {
	lib.banMu.RLock()
	defer lib.banMu.RUnlock()
	if len(ips) == 0 {
		return ips
	}
	banned := lib.blacklistIPs[host]
	if len(banned) == 0 {
		return ips
	}
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		if _, isBanned := banned[ip]; !isBanned {
			out = append(out, ip)
		}
	}
	return out
}

// GetLastSyncTime 获取最后同步时间
func (lib *IPPoolLibrary) GetLastSyncTime() time.Time {
	lib.lastSyncTimeMu.RLock()
	defer lib.lastSyncTimeMu.RUnlock()

	return lib.lastSyncTime
}

// IsAutoSyncEnabled 检查自动同步是否启用
func (lib *IPPoolLibrary) IsAutoSyncEnabled() bool {
	lib.autoSyncMu.RLock()
	defer lib.autoSyncMu.RUnlock()

	return lib.autoSyncEnabled
}

// Close 关闭库（停止自动同步）
func (lib *IPPoolLibrary) Close() {
	lib.StopAutoSync()
	if lib.client != nil {
		lib.client.Close()
	}
}
