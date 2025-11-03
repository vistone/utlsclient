package ippool

import (
	"fmt"
)

// Analyzer IP 池分析器
type Analyzer struct {
	library *IPPoolLibrary
}

// NewAnalyzer 创建新的分析器
func NewAnalyzer(library *IPPoolLibrary) *Analyzer {
	return &Analyzer{
		library: library,
	}
}

// AnalyzeStats 分析统计信息
type AnalyzeStats struct {
	// 主机统计
	TotalHosts int

	// IP 统计
	TotalIPv4 int
	TotalIPv6 int

	// 地理位置统计
	Countries map[string]int // 国家 -> IP 数量
	Cities    map[string]int // 城市 -> IP 数量
	Regions   map[string]int // 地区 -> IP 数量

	// ISP 统计
	ISPs map[string]int // ISP -> IP 数量
	Orgs map[string]int // 组织 -> IP 数量

	// 数据中心统计
	DataCenters map[string]int // 数据中心 -> IP 数量

	// IP 类型统计
	IPTypes map[string]int // IP 类型 -> IP 数量
}

// AnalyzeAll 分析所有数据
func (a *Analyzer) AnalyzeAll() (*AnalyzeStats, error) {
	stats := &AnalyzeStats{
		Countries:   make(map[string]int),
		Cities:      make(map[string]int),
		Regions:     make(map[string]int),
		ISPs:        make(map[string]int),
		Orgs:        make(map[string]int),
		DataCenters: make(map[string]int),
		IPTypes:     make(map[string]int),
	}

	hosts := a.library.GetAllHosts()
	stats.TotalHosts = len(hosts)

	// 遍历所有主机
	for _, host := range hosts {
		// 分析简化格式
		pool, err := a.library.GetIPPool(host.Host)
		if err == nil {
			stats.TotalIPv4 += len(pool.IPv4)
			stats.TotalIPv6 += len(pool.IPv6)
		}

		// 分析详细格式
		detailPool, err := a.library.GetDetailIPPool(host.Host)
		if err == nil {
			for _, ipInfo := range detailPool.IPs {
				// 统计地理位置
				if ipInfo.Location.Country != "" {
					stats.Countries[ipInfo.Location.Country]++
				}
				if ipInfo.Location.Region != "" {
					stats.Regions[ipInfo.Location.Region]++
				}
				if ipInfo.Location.City != "" {
					stats.Cities[ipInfo.Location.City]++
				}

				// 统计 ISP 和组织
				if ipInfo.Location.ISP != "" {
					stats.ISPs[ipInfo.Location.ISP]++
				}
				if ipInfo.Location.Org != "" {
					stats.Orgs[ipInfo.Location.Org]++
				}

				// 统计数据中心
				if ipInfo.Location.DataCenter != "" {
					stats.DataCenters[ipInfo.Location.DataCenter]++
				}

				// 统计 IP 类型
				if ipInfo.Location.IPType != "" {
					stats.IPTypes[ipInfo.Location.IPType]++
				}
			}
		}
	}

	return stats, nil
}

// AnalyzeByCountry 按国家分析
func (a *Analyzer) AnalyzeByCountry(country string) ([]*IPDetailInfo, error) {
	var result []*IPDetailInfo

	hosts := a.library.GetAllHosts()
	for _, host := range hosts {
		detailPool, err := a.library.GetDetailIPPool(host.Host)
		if err != nil {
			continue
		}

		for _, ipInfo := range detailPool.IPs {
			if ipInfo.Location.Country == country {
				result = append(result, ipInfo)
			}
		}
	}

	return result, nil
}

// AnalyzeByCity 按城市分析
func (a *Analyzer) AnalyzeByCity(city string) ([]*IPDetailInfo, error) {
	var result []*IPDetailInfo

	hosts := a.library.GetAllHosts()
	for _, host := range hosts {
		detailPool, err := a.library.GetDetailIPPool(host.Host)
		if err != nil {
			continue
		}

		for _, ipInfo := range detailPool.IPs {
			if ipInfo.Location.City == city {
				result = append(result, ipInfo)
			}
		}
	}

	return result, nil
}

// AnalyzeByISP 按 ISP 分析
func (a *Analyzer) AnalyzeByISP(isp string) ([]*IPDetailInfo, error) {
	var result []*IPDetailInfo

	hosts := a.library.GetAllHosts()
	for _, host := range hosts {
		detailPool, err := a.library.GetDetailIPPool(host.Host)
		if err != nil {
			continue
		}

		for _, ipInfo := range detailPool.IPs {
			if ipInfo.Location.ISP == isp {
				result = append(result, ipInfo)
			}
		}
	}

	return result, nil
}

// AnalyzeByDataCenter 按数据中心分析
func (a *Analyzer) AnalyzeByDataCenter(dataCenter string) ([]*IPDetailInfo, error) {
	var result []*IPDetailInfo

	hosts := a.library.GetAllHosts()
	for _, host := range hosts {
		detailPool, err := a.library.GetDetailIPPool(host.Host)
		if err != nil {
			continue
		}

		for _, ipInfo := range detailPool.IPs {
			if ipInfo.Location.DataCenter == dataCenter {
				result = append(result, ipInfo)
			}
		}
	}

	return result, nil
}

// GetCountriesList 获取所有国家列表及其统计信息
// 返回: map[国家名称]IP数量
func (a *Analyzer) GetCountriesList() (map[string]int, error) {
	stats, err := a.AnalyzeAll()
	if err != nil {
		return nil, err
	}
	return stats.Countries, nil
}

// GetCitiesByCountry 获取指定国家的所有城市列表及其统计信息
// 返回: map[城市名称]IP数量
func (a *Analyzer) GetCitiesByCountry(country string) (map[string]int, error) {
	cities := make(map[string]int)

	hosts := a.library.GetAllHosts()
	for _, host := range hosts {
		detailPool, err := a.library.GetDetailIPPool(host.Host)
		if err != nil {
			continue
		}

		for _, ipInfo := range detailPool.IPs {
			if ipInfo.Location.Country == country && ipInfo.Location.City != "" {
				cities[ipInfo.Location.City]++
			}
		}
	}

	return cities, nil
}

// GetIPsByCountryAndCity 获取指定国家和城市的所有 IP
func (a *Analyzer) GetIPsByCountryAndCity(country, city string) ([]*IPDetailInfo, error) {
	var result []*IPDetailInfo

	hosts := a.library.GetAllHosts()
	for _, host := range hosts {
		detailPool, err := a.library.GetDetailIPPool(host.Host)
		if err != nil {
			continue
		}

		for _, ipInfo := range detailPool.IPs {
			matchCountry := country == "" || ipInfo.Location.Country == country
			matchCity := city == "" || ipInfo.Location.City == city

			if matchCountry && matchCity {
				result = append(result, ipInfo)
			}
		}
	}

	return result, nil
}

// AnalyzeByHost 分析指定主机的所有 IP
func (a *Analyzer) AnalyzeByHost(host string) (*AnalyzeStats, error) {
	stats := &AnalyzeStats{
		TotalHosts:  1,
		Countries:   make(map[string]int),
		Cities:      make(map[string]int),
		Regions:     make(map[string]int),
		ISPs:        make(map[string]int),
		Orgs:        make(map[string]int),
		DataCenters: make(map[string]int),
		IPTypes:     make(map[string]int),
	}

	// 分析简化格式
	pool, err := a.library.GetIPPool(host)
	if err == nil {
		stats.TotalIPv4 = len(pool.IPv4)
		stats.TotalIPv6 = len(pool.IPv6)
	}

	// 分析详细格式
	detailPool, err := a.library.GetDetailIPPool(host)
	if err != nil {
		return nil, fmt.Errorf("获取详细数据失败: %w", err)
	}

	for _, ipInfo := range detailPool.IPs {
		// 统计地理位置
		if ipInfo.Location.Country != "" {
			stats.Countries[ipInfo.Location.Country]++
		}
		if ipInfo.Location.Region != "" {
			stats.Regions[ipInfo.Location.Region]++
		}
		if ipInfo.Location.City != "" {
			stats.Cities[ipInfo.Location.City]++
		}

		// 统计 ISP 和组织
		if ipInfo.Location.ISP != "" {
			stats.ISPs[ipInfo.Location.ISP]++
		}
		if ipInfo.Location.Org != "" {
			stats.Orgs[ipInfo.Location.Org]++
		}

		// 统计数据中心
		if ipInfo.Location.DataCenter != "" {
			stats.DataCenters[ipInfo.Location.DataCenter]++
		}

		// 统计 IP 类型
		if ipInfo.Location.IPType != "" {
			stats.IPTypes[ipInfo.Location.IPType]++
		}
	}

	return stats, nil
}

// GetRandomIP 从指定主机随机获取一个 IP（简化格式）
func (a *Analyzer) GetRandomIP(host string) (string, error) {
	pool, err := a.library.GetIPPool(host)
	if err != nil {
		return "", err
	}

	if len(pool.IPv4) == 0 && len(pool.IPv6) == 0 {
		return "", fmt.Errorf("主机 %s 没有可用的 IP", host)
	}

	// 优先返回 IPv4
	if len(pool.IPv4) > 0 {
		return pool.IPv4[0], nil // 简化实现，实际可以使用随机选择
	}

	return pool.IPv6[0], nil
}

// GetAllIPsByHost 获取指定主机的所有 IP（简化格式）
func (a *Analyzer) GetAllIPsByHost(host string) ([]string, []string, error) {
	pool, err := a.library.GetIPPool(host)
	if err != nil {
		return nil, nil, err
	}

	ipv4 := make([]string, len(pool.IPv4))
	copy(ipv4, pool.IPv4)

	ipv6 := make([]string, len(pool.IPv6))
	copy(ipv6, pool.IPv6)

	return ipv4, ipv6, nil
}

// SearchIPs 搜索 IP（支持按多个条件搜索）
func (a *Analyzer) SearchIPs(host, country, city, isp, dataCenter string) ([]*IPDetailInfo, error) {
	var result []*IPDetailInfo

	var hosts []string
	if host != "" {
		hosts = []string{host}
	} else {
		allHosts := a.library.GetAllHosts()
		hosts = make([]string, len(allHosts))
		for i, h := range allHosts {
			hosts[i] = h.Host
		}
	}

	for _, h := range hosts {
		detailPool, err := a.library.GetDetailIPPool(h)
		if err != nil {
			continue
		}

		for _, ipInfo := range detailPool.IPs {
			match := true

			if country != "" && ipInfo.Location.Country != country {
				match = false
			}
			if city != "" && ipInfo.Location.City != city {
				match = false
			}
			if isp != "" && ipInfo.Location.ISP != isp {
				match = false
			}
			if dataCenter != "" && ipInfo.Location.DataCenter != dataCenter {
				match = false
			}

			if match {
				result = append(result, ipInfo)
			}
		}
	}

	return result, nil
}
