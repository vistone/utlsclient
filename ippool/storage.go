package ippool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LoadFromLocal 从本地文件加载所有数据（网络不通时使用本地数据）
func (lib *IPPoolLibrary) LoadFromLocal() error {
	// 加载主机列表
	if err := lib.loadHostsFromLocal(); err != nil {
		// 如果本地文件不存在，静默失败（可能是第一次运行，等待网络同步）
		return nil
	}

	// 加载所有主机的 IP 池数据到内存
	hosts := lib.GetAllHosts()
	for _, host := range hosts {
		// 加载简化格式
		_ = lib.loadIPPoolFromLocal(host.Host)

		// 加载详细格式
		if host.DetailExists {
			_ = lib.loadDetailIPPoolFromLocal(host.Host)
		}
	}

	return nil
}

// loadHostsFromLocal 从本地文件加载主机列表
func (lib *IPPoolLibrary) loadHostsFromLocal() error {
	filePath := filepath.Join(lib.dataDir, "hosts.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var apiResp IPPoolResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return err
	}

	lib.hostsMu.Lock()
	lib.hosts = apiResp.Hosts
	lib.hostsMu.Unlock()

	return nil
}

// saveHostsToLocal 保存主机列表到本地文件
func (lib *IPPoolLibrary) saveHostsToLocal() error {
	hosts := lib.GetAllHosts()
	apiResp := IPPoolResponse{
		Hosts: hosts,
	}

	data, err := json.MarshalIndent(apiResp, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(lib.dataDir, "hosts.json")
	return os.WriteFile(filePath, data, 0644)
}

// loadIPPoolFromLocal 从本地文件加载 IP 池数据（简化格式）
func (lib *IPPoolLibrary) loadIPPoolFromLocal(host string) error {
	// 将 host 转换为文件名（替换特殊字符）
	fileName := sanitizeFileName(host) + ".json"
	filePath := filepath.Join(lib.dataDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var poolData IPPoolData
	if err := json.Unmarshal(data, &poolData); err != nil {
		return err
	}

	lib.ipPoolsMu.Lock()
	lib.ipPools[host] = &poolData
	lib.ipPoolsMu.Unlock()

	return nil
}

// loadDetailIPPoolFromLocal 从本地文件加载详细 IP 池数据
// 直接从本地保存的原始 JSON 文件加载
func (lib *IPPoolLibrary) loadDetailIPPoolFromLocal(host string) error {
	fileName := sanitizeFileName(host) + "_detail.json"
	filePath := filepath.Join(lib.dataDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 解析 JSON（动态结构，保持服务器原始格式）
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}

	// 构建详细数据
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
	// JSON 结构: {ipv4_detailed: {IP地址: {ip: "...", location: {...}}}, ipv6_detailed: {...}}
	// 需要从 ipv4_detailed 和 ipv6_detailed 中提取数据
	for _, detailKey := range []string{"ipv4_detailed", "ipv6_detailed"} {
		if detailObj, ok := rawData[detailKey].(map[string]interface{}); ok {
			for ipAddr, ipDataRaw := range detailObj {
				if ipData, ok := ipDataRaw.(map[string]interface{}); ok {
					ipInfo := &IPDetailInfo{}

					// IP 地址就是 key，但也从数据中获取（如果存在）
					ipInfo.IP = ipAddr
					if ip, ok := ipData["ip"].(string); ok && ip != "" {
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

					// 使用 IP 地址作为 key
					detailData.IPs[ipInfo.IP] = ipInfo
				}
			}
		}
	}

	lib.detailPoolsMu.Lock()
	lib.detailPools[host] = detailData
	lib.detailPoolsMu.Unlock()

	// 更新最后更新时间
	if !serverLastUpdated.IsZero() {
		lib.hostLastUpdatedMu.Lock()
		lib.hostLastUpdated[host] = serverLastUpdated
		lib.hostLastUpdatedMu.Unlock()
	}

	return nil
}

// sanitizeFileName 将主机名转换为安全的文件名
func sanitizeFileName(host string) string {
	result := ""
	for _, char := range host {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' {
			result += string(char)
		} else {
			result += "_"
		}
	}
	return result
}

// GetLocalDataInfo 获取本地数据信息
func (lib *IPPoolLibrary) GetLocalDataInfo() map[string]interface{} {
	info := make(map[string]interface{})
	info["data_dir"] = lib.dataDir

	// 检查文件是否存在
	hostsFile := filepath.Join(lib.dataDir, "hosts.json")
	if _, err := os.Stat(hostsFile); err == nil {
		info["hosts_file_exists"] = true
		if stat, err := os.Stat(hostsFile); err == nil {
			info["hosts_file_modified"] = stat.ModTime().Format(time.RFC3339)
		}
	} else {
		info["hosts_file_exists"] = false
	}

	// 统计已保存的主机数据文件
	hosts := lib.GetAllHosts()
	fileCount := 0
	for _, host := range hosts {
		fileName := sanitizeFileName(host.Host) + ".json"
		filePath := filepath.Join(lib.dataDir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			fileCount++
		}
	}
	info["ip_pool_files_count"] = fileCount

	return info
}
