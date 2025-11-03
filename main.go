package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"utls_client/ippool"
	clientLib "utls_client/lib"

	utls "github.com/refraction-networking/utls"
)

func main() {
	fmt.Println("=== 系统自检：IPv4/IPv6 与地址池能力检查 ===")

	ipv4OK, ipv4Ifs := detectIPv4()
	ipv6OK, ipv6Ifs := detectIPv6()
	hasV6Default, v6DefaultDev := detectIPv6DefaultRoute()
	hasSIT, sitInfo := detectSIT("ipv6net")

	fmt.Printf("IPv4 支持: %v (接口: %s)\n", ipv4OK, strings.Join(ipv4Ifs, ", "))
	fmt.Printf("IPv6 支持: %v (接口: %s)\n", ipv6OK, strings.Join(ipv6Ifs, ", "))
	fmt.Printf("IPv6 默认路由(::/0): %v", hasV6Default)
	if v6DefaultDev != "" {
		fmt.Printf(" (dev: %s)", v6DefaultDev)
	}
	fmt.Println()
	fmt.Printf("SIT 隧道 ipv6net: %v", hasSIT)
	if sitInfo != "" {
		fmt.Printf(" (%s)", sitInfo)
	}
	fmt.Println()

	// 地址池能力：加载本地数据并统计 IPv6 可用性
	lib := ippool.NewIPPoolLibrary("", "./ippool_data")
	defer lib.Close()
	hosts := lib.GetAllHosts()
	totalHosts := len(hosts)
	hostsWithV6 := 0
	for _, h := range hosts {
		pool, err := lib.GetIPPool(h.Host)
		if err == nil && len(pool.IPv6) > 0 {
			hostsWithV6++
		}
	}
	fmt.Printf("已加载主机数: %d，其中支持 IPv6 的主机: %d\n", totalHosts, hostsWithV6)

	fmt.Println()
	// 构建全局连接池管理器（长连常驻）
	connMgr = clientLib.NewConnPoolManager(utls.HelloChrome_133, &clientLib.Config{Timeout: 30 * time.Second, ServerName: "kh.google.com"})
	// 测试：从环境变量注入黑名单
	seedBlacklistFromEnv(lib, "kh.google.com")
	fmt.Println("=== 首次预热 kh.google.com IPv6 长连接（仅白名单） ===")
	initialWarmup(lib, "kh.google.com", true)
	fmt.Println("=== 首次预热 kh.google.com IPv4 长连接（仅白名单） ===")
	initialWarmup(lib, "kh.google.com", false)

	fmt.Println()
	fmt.Println("=== 操作建议 ===")
	if !ipv6OK || !hasV6Default {
		fmt.Println("未检测到稳定的公网 IPv6 通路或默认路由，可按需配置 SIT 隧道 (示例)：")
		fmt.Println("  ip tunnel add ipv6net mode sit local <LOCAL_IPV4> remote <REMOTE_IPV4> ttl 255")
		fmt.Println("  ip link set ipv6net up")
		fmt.Println("  ip addr add 2607:8700:5500:2943::2/64 dev ipv6net")
		fmt.Println("  ip route add ::/0 dev ipv6net")
	} else {
		fmt.Println("IPv6 默认路由存在，建议直接进行 IPv6 地址池测试与管理。")
	}

	// 仅黑名单健康检查（每20分钟）
	go func() {
		ticker := time.NewTicker(20 * time.Minute)
		for range ticker.C {
			fmt.Println("\n=== 黑名单健康检查：kh.google.com ===")
			checkBlacklisted(lib, "kh.google.com", true)
			checkBlacklisted(lib, "kh.google.com", false)
		}
	}()

	fmt.Println("\n✅ 自检完成，长连接常驻：仅在请求403时移出池；黑名单每20分钟健康检查，200后再加入池。按 Ctrl+C 退出。")
	select {}
}

// 全局长连接池管理器
var connMgr *clientLib.ConnPoolManager

// 保留：旧版全量预热函数（不再在主流程使用）
func warmupKHIPv6() {
	lib := ippool.NewIPPoolLibrary("", "./ippool_data")
	defer lib.Close()
	host := "kh.google.com"
	analyzer := ippool.NewAnalyzer(lib)
	_, v6s, err := analyzer.GetAllIPsByHost(host)
	if err != nil || len(v6s) == 0 {
		fmt.Println("未找到 kh.google.com 的 IPv6 列表，跳过预热")
		return
	}
	// 使用地址池内的全部 IPv6

	// 实际发起一次 GET 以建立 TLS+HTTP/2 长连接
	fp := utls.HelloChrome_133
	conf := &clientLib.Config{Timeout: 30 * time.Second, ServerName: host}

	sem := make(chan struct{}, 64)
	var wg sync.WaitGroup
	for _, ip := range v6s {
		wg.Add(1)
		sem <- struct{}{}
		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-sem }()
			c := clientLib.NewClient(&fp, conf)
			url := fmt.Sprintf("https://[%s]/rt/earth/PlanetoidMetadata", targetIP)
			_, _ = c.Do("GET", url, &clientLib.RequestConfig{
				Method: "GET",
				Headers: map[string]string{
					"Host":            host,
					"Accept-Encoding": "gzip",
					"User-Agent":      "Mozilla/5.0",
				},
				Host: host,
			})
			_ = c.Close()
		}(ip)
	}
	wg.Wait()
	fmt.Printf("已预热 kh.google.com 的 IPv6 连接数: %d\n", len(v6s))
}

// 保留：旧版全量预热函数（不再在主流程使用）
func warmupKHIPv4() {
	lib := ippool.NewIPPoolLibrary("", "./ippool_data")
	defer lib.Close()
	host := "kh.google.com"
	analyzer := ippool.NewAnalyzer(lib)
	v4s, _, err := analyzer.GetAllIPsByHost(host)
	if err != nil || len(v4s) == 0 {
		fmt.Println("未找到 kh.google.com 的 IPv4 列表，跳过预热")
		return
	}
	// 使用地址池内的全部 IPv4
	fp := utls.HelloChrome_133
	conf := &clientLib.Config{Timeout: 30 * time.Second, ServerName: host}
	sem := make(chan struct{}, 64)
	var wg sync.WaitGroup
	for _, ip := range v4s {
		wg.Add(1)
		sem <- struct{}{}
		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-sem }()
			c := clientLib.NewClient(&fp, conf)
			url := fmt.Sprintf("https://%s/rt/earth/PlanetoidMetadata", targetIP)
			_, _ = c.Do("GET", url, &clientLib.RequestConfig{
				Method: "GET",
				Headers: map[string]string{
					"Host":            host,
					"Accept-Encoding": "gzip",
					"User-Agent":      "Mozilla/5.0",
				},
				Host: host,
			})
			_ = c.Close()
		}(ip)
	}
	wg.Wait()
	fmt.Printf("已预热 kh.google.com 的 IPv4 连接数: %d\n", len(v4s))
}

func detectIPv4() (bool, []string) {
	ifs, _ := net.Interfaces()
	names := []string{}
	ok := false
	for _, iface := range ifs {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			if ipnet, okAddr := a.(*net.IPNet); okAddr {
				ip := ipnet.IP
				if ip.To4() != nil && ip.IsGlobalUnicast() {
					names = append(names, iface.Name)
					ok = true
					break
				}
			}
		}
	}
	if len(names) == 0 {
		names = []string{"(none)"}
	}
	return ok, unique(names)
}

func detectIPv6() (bool, []string) {
	ifs, _ := net.Interfaces()
	names := []string{}
	ok := false
	for _, iface := range ifs {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			if ipnet, okAddr := a.(*net.IPNet); okAddr {
				ip := ipnet.IP
				if ip.To16() != nil && ip.To4() == nil && ip.IsGlobalUnicast() {
					names = append(names, iface.Name)
					ok = true
					break
				}
			}
		}
	}
	if len(names) == 0 {
		names = []string{"(none)"}
	}
	return ok, unique(names)
}

func detectIPv6DefaultRoute() (bool, string) {
	// 优先尝试 ip 命令
	if path, _ := exec.LookPath("ip"); path != "" {
		out, err := exec.Command(path, "-6", "route", "show", "default").Output()
		if err == nil {
			s := strings.TrimSpace(string(out))
			if s != "" {
				// 可能包含 "default via ... dev <dev>"
				dev := parseDevFromIPRoute(s)
				return true, dev
			}
		}
	}
	// 备用：/proc/net/ipv6_route 默认路由项目的前缀全 0
	f, err := os.Open("/proc/net/ipv6_route")
	if err != nil {
		return false, ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// 字段：dest(32) dest_prefixlen src(32) src_prefixlen ... dev
		parts := strings.Fields(line)
		if len(parts) >= 10 {
			if parts[0] == strings.Repeat("0", 32) && parts[1] == "00000000" {
				// 默认路由
				dev := parts[len(parts)-1]
				return true, dev
			}
		}
	}
	return false, ""
}

func detectSIT(name string) (bool, string) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return false, ""
	}
	addrs, _ := iface.Addrs()
	hasV6 := false
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip := ipnet.IP; ip.To16() != nil && ip.To4() == nil {
				if ip.IsGlobalUnicast() {
					hasV6 = true
					break
				}
			}
		}
	}
	return true, fmt.Sprintf("up=%v, v6_global=%v", (iface.Flags&net.FlagUp) != 0, hasV6)
}

func parseDevFromIPRoute(out string) string {
	// 解析 "dev <name>" 片段
	fields := strings.Fields(out)
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == "dev" {
			return fields[i+1]
		}
	}
	return ""
}

func unique(in []string) []string {
	if len(in) == 0 {
		return in
	}
	m := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := m[s]; ok {
			continue
		}
		m[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// initialWarmup：仅对白名单进行一次性预热，将连接保留在 connMgr 池中
func initialWarmup(lib *ippool.IPPoolLibrary, host string, ipv6 bool) {
	analyzer := ippool.NewAnalyzer(lib)
	v4s, v6s, err := analyzer.GetAllIPsByHost(host)
	if err != nil {
		fmt.Printf("获取 %s IP 列表失败: %v\n", host, err)
		return
	}
	var ips []string
	var label string
	if ipv6 {
		ips, label = v6s, "IPv6"
	} else {
		ips, label = v4s, "IPv4"
	}
	if len(ips) == 0 {
		fmt.Printf("未找到 %s 的 %s 列表，跳过预热\n", host, label)
		return
	}

	allowed := lib.FilterIPs(host, ips)
	if len(allowed) == 0 {
		fmt.Printf("%s 白名单为空，跳过预热\n", label)
		return
	}
	connMgr.WarmUp(allowed)

	fp := utls.HelloChrome_133
	conf := &clientLib.Config{Timeout: 30 * time.Second, ServerName: host}
	sem := make(chan struct{}, 64)
	var wg sync.WaitGroup
	success := int64(0)
	for _, ip := range allowed {
		wg.Add(1)
		sem <- struct{}{}
		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-sem }()
			c, ok := connMgr.Get(targetIP)
			if !ok {
				c = clientLib.NewClient(&fp, conf)
				connMgr.WarmUp([]string{targetIP})
			}
			var url string
			if ipv6 {
				url = fmt.Sprintf("https://[%s]/rt/earth/PlanetoidMetadata", targetIP)
			} else {
				url = fmt.Sprintf("https://%s/rt/earth/PlanetoidMetadata", targetIP)
			}
			resp, err := c.Do("GET", url, &clientLib.RequestConfig{Method: "GET", Headers: map[string]string{"Host": host, "Accept-Encoding": "gzip", "User-Agent": "Mozilla/5.0"}, Host: host})
			if err == nil && resp != nil && resp.StatusCode == 200 {
				success++
				lib.ReportStatus(host, targetIP, 200)
			}
			if err == nil && resp != nil && resp.StatusCode == 403 {
				lib.ReportStatus(host, targetIP, 403)
			}
		}(ip)
	}
	wg.Wait()
	fmt.Printf("%s 首次预热完成：成功 %d / %d\n", label, success, len(allowed))
}

// checkBlacklisted：仅对黑名单做健康检查，200 则解封并保留长连
func checkBlacklisted(lib *ippool.IPPoolLibrary, host string, ipv6 bool) {
	analyzer := ippool.NewAnalyzer(lib)
	v4s, v6s, err := analyzer.GetAllIPsByHost(host)
	if err != nil {
		return
	}
	var ips []string
	var label string
	if ipv6 {
		ips, label = v6s, "IPv6"
	} else {
		ips, label = v4s, "IPv4"
	}
	if len(ips) == 0 {
		return
	}
	banned := make([]string, 0)
	for _, ip := range ips {
		if !lib.IsAllowed(host, ip) {
			banned = append(banned, ip)
		}
	}
	if len(banned) == 0 {
		fmt.Printf("%s 黑名单为空，跳过\n", label)
		return
	}

	fp := utls.HelloChrome_133
	conf := &clientLib.Config{Timeout: 30 * time.Second, ServerName: host}
	sem := make(chan struct{}, 32)
	var wg sync.WaitGroup
	unbanned := int64(0)
	for _, ip := range banned {
		wg.Add(1)
		sem <- struct{}{}
		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-sem }()
			c, ok := connMgr.Get(targetIP)
			if !ok {
				c = clientLib.NewClient(&fp, conf)
				connMgr.WarmUp([]string{targetIP})
			}
			var url string
			if ipv6 {
				url = fmt.Sprintf("https://[%s]/rt/earth/PlanetoidMetadata", targetIP)
			} else {
				url = fmt.Sprintf("https://%s/rt/earth/PlanetoidMetadata", targetIP)
			}
			resp, err := c.Do("GET", url, &clientLib.RequestConfig{Method: "GET", Headers: map[string]string{"Host": host, "Accept-Encoding": "gzip", "User-Agent": "Mozilla/5.0"}, Host: host})
			if err == nil && resp != nil && resp.StatusCode == 200 {
				lib.ReportStatus(host, targetIP, 200)
				unbanned++
			}
		}(ip)
	}
	wg.Wait()
	fmt.Printf("%s 黑名单健康检查：解封成功 %d / %d\n", label, unbanned, len(banned))
}

// warmupHost 使用黑/白名单策略进行IPv4/IPv6预热
func warmupHost(lib *ippool.IPPoolLibrary, host string, ipv6 bool) {
	analyzer := ippool.NewAnalyzer(lib)
	v4s, v6s, err := analyzer.GetAllIPsByHost(host)
	if err != nil {
		fmt.Printf("获取 %s IP 列表失败: %v\n", host, err)
		return
	}
	var ips []string
	var label string
	if ipv6 {
		ips, label = v6s, "IPv6"
	} else {
		ips, label = v4s, "IPv4"
	}
	if len(ips) == 0 {
		fmt.Printf("未找到 %s 的 %s 列表，跳过预热\n", host, label)
		return
	}

	allowed := lib.FilterIPs(host, ips)
	bannedSet := map[string]struct{}{}
	for _, ip := range ips {
		if !lib.IsAllowed(host, ip) {
			bannedSet[ip] = struct{}{}
		}
	}

	fp := utls.HelloChrome_133
	conf := &clientLib.Config{Timeout: 30 * time.Second, ServerName: host}
	sem := make(chan struct{}, 64)
	var wg sync.WaitGroup

	successAllowed := int64(0)
	for _, ip := range allowed {
		wg.Add(1)
		sem <- struct{}{}
		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-sem }()
			c := clientLib.NewClient(&fp, conf)
			var url string
			if ipv6 {
				url = fmt.Sprintf("https://[%s]/rt/earth/PlanetoidMetadata", targetIP)
			} else {
				url = fmt.Sprintf("https://%s/rt/earth/PlanetoidMetadata", targetIP)
			}
			resp, err := c.Do("GET", url, &clientLib.RequestConfig{
				Method: "GET",
				Headers: map[string]string{
					"Host":            host,
					"Accept-Encoding": "gzip",
					"User-Agent":      "Mozilla/5.0",
				},
				Host: host,
			})
			if err == nil && resp != nil {
				lib.ReportStatus(host, targetIP, resp.StatusCode)
				if resp.StatusCode == 200 {
					successAllowed++
				}
			}
			_ = c.Close()
		}(ip)
	}

	successUnban := int64(0)
	for bannedIP := range bannedSet {
		wg.Add(1)
		sem <- struct{}{}
		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-sem }()
			c := clientLib.NewClient(&fp, conf)
			var url string
			if ipv6 {
				url = fmt.Sprintf("https://[%s]/rt/earth/PlanetoidMetadata", targetIP)
			} else {
				url = fmt.Sprintf("https://%s/rt/earth/PlanetoidMetadata", targetIP)
			}
			resp, err := c.Do("GET", url, &clientLib.RequestConfig{
				Method: "GET",
				Headers: map[string]string{
					"Host":            host,
					"Accept-Encoding": "gzip",
					"User-Agent":      "Mozilla/5.0",
				},
				Host: host,
			})
			if err == nil && resp != nil {
				lib.ReportStatus(host, targetIP, resp.StatusCode)
				if resp.StatusCode == 200 {
					successUnban++
				}
			}
			_ = c.Close()
		}(bannedIP)
	}

	wg.Wait()
	fmt.Printf("%s 预热完成：允许列表成功 %d / %d，黑名单解封成功 %d / %d\n",
		label, successAllowed, len(allowed), successUnban, len(bannedSet))
}

// seedBlacklistFromEnv: 通过环境变量 BLACKLIST_TEST_IPS 预置黑名单，逗号分隔
func seedBlacklistFromEnv(lib *ippool.IPPoolLibrary, host string) {
	env := strings.TrimSpace(os.Getenv("BLACKLIST_TEST_IPS"))
	if env == "" {
		return
	}
	ips := strings.Split(env, ",")
	count := 0
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		lib.ReportStatus(host, ip, 403)
		count++
	}
	if count > 0 {
		fmt.Printf("测试：已将 %d 个 IP 加入黑名单（BLACKLIST_TEST_IPS）\n", count)
	}
}
