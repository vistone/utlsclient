//go:build ignore

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"utls_client/fingerprint"
	"utls_client/ippool"
	clientLib "utls_client/lib"
)

type RequestResult struct {
	IP         string
	StatusCode int
	Success    bool
	Error      error
	Duration   time.Duration
	Country    string // 国家
	City       string // 城市
}

// Task 表示一次要完成的请求任务（同一目标资源，不绑定具体 IP）
type Task struct {
	ID          int
	Attempts    int
	LastTriedIP string
}

func main() {
	fmt.Println("=== IP 池并发测试 ===")
	fmt.Println("向所有 IP 同时发送请求（每个 IP 只请求一次）")
	fmt.Println("目标地址: https://kh.google.com/rt/earth/BulkMetadata/pb=!1m2!1s!2u1002")
	fmt.Println()

	// 1. 初始化 IP 池库
	dataDir := "./ippool_data"
	library := ippool.NewIPPoolLibrary("http://tile0.zeromaps.cn:9005", dataDir)
	defer library.Close()

	// 直接从本地加载数据（不进行网络同步）
	// 如果需要更新数据，可以手动调用 library.SyncAll() 或启动定时同步
	fmt.Println("✅ 已从本地加载 IP 池数据")
	fmt.Println("   提示：如果需要同步最新数据，请调用 library.SyncAll()")
	fmt.Println()

	// 2. 查找 kh.google.com 主机的 IP
	hostName := "kh.google.com"
	hosts := library.GetAllHosts()
	var targetHost *ippool.HostInfo
	for _, host := range hosts {
		if host.Host == hostName {
			targetHost = &host
			break
		}
	}

	if targetHost == nil {
		fmt.Printf("❌ 未找到主机: %s\n", hostName)
		fmt.Println("可用主机列表:")
		for _, h := range hosts {
			fmt.Printf("  - %s\n", h.Host)
		}
		return
	}

	// 3. 获取所有 IP 地址
	analyzer := ippool.NewAnalyzer(library)
	ipv4List, ipv6List, err := analyzer.GetAllIPsByHost(hostName)
	if err != nil {
		fmt.Printf("❌ 获取 IP 列表失败: %v\n", err)
		return
	}

	// 合并 IPv4 和 IPv6，优先使用 IPv4
	allIPs := append(ipv4List, ipv6List...)
	totalIPs := len(allIPs)

	if totalIPs == 0 {
		fmt.Printf("❌ 主机 %s 没有可用的 IP\n", hostName)
		return
	}

	fmt.Printf("✅ 找到 %d 个 IP (IPv4: %d, IPv6: %d)\n", totalIPs, len(ipv4List), len(ipv6List))
	fmt.Println()

	// 4. 任务队列 + 工作者模型
	// 生成 1000 个任务，每个任务请求同一目标地址
	// 由 totalIPs 个工作者（每个对应一个 IP）并发执行
	taskCount := 1000
	target := "https://kh.google.com/rt/earth/BulkMetadata/pb=!1m2!1s!2u1002"
	results := make([]*RequestResult, taskCount)
	var successCount int64
	var failCount int64

	// 从指纹库随机获取指纹
	lib := fingerprint.NewFingerprintLibrary()
	profile := lib.GetRandomProfile()

	fmt.Printf("📊 开始任务队列 + 工作者模型...\n")
	fmt.Printf("   目标地址: %s\n", target)
	fmt.Printf("   任务数量: %d\n", taskCount)
	fmt.Printf("   工作者数量（IP数）: %d\n", totalIPs)
	// 读取环境变量 REQ_TIMEOUT_MS，默认 2000ms
	reqTimeoutMs := 2000
	if v := os.Getenv("REQ_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			reqTimeoutMs = n
		}
	}
	reqTimeout := time.Duration(reqTimeoutMs) * time.Millisecond
	fmt.Printf("   超时策略: %dms 超时，失败任务由其他空闲IP继续执行\n", reqTimeoutMs)
	fmt.Printf("   使用指纹: %s\n", profile.Name)
	if os.Getenv("PREWARM") == "1" {
		fmt.Println("   预热: 已启用（每个 IP 在计时前先发起一次预热请求）")
	} else {
		fmt.Println("   预热: 未启用（设置 PREWARM=1 可启用）")
	}
	fmt.Println()

	startTime := time.Now()

	fmt.Println("开始发送请求（任务队列 + 工作者模型，超时任务由其他IP继续）...")
	fmt.Println()

	// 5. 初始化任务队列（1000 个任务）
	tasks := make(chan *Task, taskCount)
	var wg sync.WaitGroup
	var printMu sync.Mutex // 保护打印

	// 初始化任务（每个任务代表一次对同一目标的请求）
	for i := 0; i < taskCount; i++ {
		tasks <- &Task{ID: i}
	}

	// 结果存储（按任务ID）
	results = make([]*RequestResult, taskCount)
	var remaining int32 = int32(taskCount)
	maxAttempts := totalIPs // 每个任务最多尝试所有 IP 的数量

	// 启动每个IP的工作者
	for _, ip := range allIPs {
		// 地理信息（打印用）
		geoCountry := ""
		geoCity := ""
		if ipDetail, err := library.GetIPDetail(hostName, ip); err == nil {
			geoCountry = ipDetail.Location.Country
			geoCity = ipDetail.Location.City
		}

		wg.Add(1)
		go func(workerIP, country, city string) {
			defer wg.Done()
			// 每个工人一个客户端
			clientConfig := &clientLib.Config{Timeout: reqTimeout, ServerName: "kh.google.com"}
			ipClient := clientLib.NewClient(&profile.HelloID, clientConfig)
			defer ipClient.Close()

			// URL 与请求头
			var ipURL string
			if strings.Contains(workerIP, ":") && !strings.HasPrefix(workerIP, "[") {
				ipURL = fmt.Sprintf("https://[%s]/rt/earth/BulkMetadata/pb=!1m2!1s!2u1002", workerIP)
			} else {
				ipURL = fmt.Sprintf("https://%s/rt/earth/BulkMetadata/pb=!1m2!1s!2u1002", workerIP)
			}
			headers := map[string]string{
				"Accept":          "application/json, text/javascript, */*; q=0.01",
				"User-Agent":      profile.UserAgent,
				"Origin":          "https://earth.google.com",
				"Referer":         "https://earth.google.com/",
				"Accept-Encoding": "gzip",
			}
			if os.Getenv("PREWARM") == "1" {
				_, _ = ipClient.Do("GET", ipURL, &clientLib.RequestConfig{Method: "GET", Headers: headers, Host: "kh.google.com"})
			}

			for {
				// 检查是否所有任务已完成
				if atomic.LoadInt32(&remaining) == 0 {
					return
				}

				select {
				case task := <-tasks:
					if task == nil {
						return
					}
					// 再次检查（避免重复处理已完成的任务）
					if atomic.LoadInt32(&remaining) == 0 {
						tasks <- task
						return
					}

					if task.LastTriedIP == workerIP {
						// 避免同一IP立即再次尝试，放回队列
						tasks <- task
						time.Sleep(1 * time.Millisecond) // 短暂延迟避免忙等待
						continue
					}

					start := time.Now()
					resp, err := ipClient.Do("GET", ipURL, &clientLib.RequestConfig{Method: "GET", Headers: headers, Host: "kh.google.com"})
					dur := time.Since(start)

					if err == nil && resp != nil && resp.StatusCode == 200 {
						// 成功：标记任务完成
						if results[task.ID] == nil {
							results[task.ID] = &RequestResult{IP: workerIP, StatusCode: 200, Success: true, Duration: dur, Country: country, City: city}
							atomic.AddInt64(&successCount, 1)
							atomic.AddInt32(&remaining, -1)
							printMu.Lock()
							geoInfo := country
							if city != "" {
								geoInfo = fmt.Sprintf("%s/%s", country, city)
							}
							if geoInfo == "" {
								geoInfo = "-"
							}
							fmt.Printf("任务 %d ✅ IP: %-15s | %-30s | 状态码: %3d | 耗时: %v\n", task.ID+1, workerIP, geoInfo, 200, dur)
							printMu.Unlock()
						}
						continue
					}

					// 失败/超时：交由其他IP继续
					task.Attempts++
					task.LastTriedIP = workerIP
					if task.Attempts < maxAttempts {
						// 打印失败（含超时标记）
						isTimeout := err != nil && (strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "deadline"))
						printMu.Lock()
						geoInfo := country
						if city != "" {
							geoInfo = fmt.Sprintf("%s/%s", country, city)
						}
						if geoInfo == "" {
							geoInfo = "-"
						}
						if isTimeout {
							fmt.Printf("任务 %d ⏱️ 超时(>2s) IP: %-15s | %-30s | 已尝试: %d | 耗时: %v -> 交由其他IP继续\n", task.ID+1, workerIP, geoInfo, task.Attempts, dur)
						} else if err != nil {
							fmt.Printf("任务 %d ❌ IP: %-15s | %-30s | 已尝试: %d | 耗时: %v | 错误: %v -> 交由其他IP继续\n", task.ID+1, workerIP, geoInfo, task.Attempts, dur, err)
						} else if resp != nil {
							fmt.Printf("任务 %d ⚠️  IP: %-15s | %-30s | 状态码: %3d | 已尝试: %d | 耗时: %v -> 交由其他IP继续\n", task.ID+1, workerIP, geoInfo, resp.StatusCode, task.Attempts, dur)
						}
						printMu.Unlock()
						tasks <- task
					} else {
						// 最终失败（已尝试所有IP）
						if results[task.ID] == nil {
							results[task.ID] = &RequestResult{IP: workerIP, Success: false, Error: err, Duration: dur, Country: country, City: city}
							atomic.AddInt64(&failCount, 1)
							atomic.AddInt32(&remaining, -1)
							printMu.Lock()
							fmt.Printf("任务 %d ❌ 最终失败（已尝试 %d 个IP）\n", task.ID+1, task.Attempts)
							printMu.Unlock()
						}
					}
				case <-time.After(100 * time.Millisecond):
					// 定期检查是否所有任务已完成，避免永久阻塞
					continue
				}
			}
		}(ip, geoCountry, geoCity)
	}

	// 等待所有任务完成
	for {
		if atomic.LoadInt32(&remaining) == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	// 关闭任务通道，通知工作者退出
	close(tasks)
	wg.Wait()

	totalDuration := time.Since(startTime)

	// 6. 统计结果
	fmt.Println("=== 测试结果 ===")
	fmt.Printf("总任务数: %d\n", taskCount)
	fmt.Printf("工作者数（IP数）: %d\n", totalIPs)
	fmt.Printf("成功任务: %d (%.2f%%)\n", successCount, float64(successCount)/float64(taskCount)*100)
	fmt.Printf("失败任务: %d (%.2f%%)\n", failCount, float64(failCount)/float64(taskCount)*100)
	fmt.Printf("总耗时: %v\n", totalDuration)
	fmt.Printf("平均每个任务耗时: %v\n", totalDuration/time.Duration(taskCount))
	if successCount > 0 {
		fmt.Printf("成功任务平均耗时: %v\n", calculateAvgDuration(results, true))
	}
	fmt.Println()

	// 7. 显示状态码分布
	statusCodeCount := make(map[int]int)
	for _, result := range results {
		if result != nil {
			statusCodeCount[result.StatusCode]++
		}
	}

	if len(statusCodeCount) > 0 {
		fmt.Println("状态码分布:")
		for code, count := range statusCodeCount {
			fmt.Printf("  %d: %d 个请求\n", code, count)
		}
		fmt.Println()
	}

	// 8. 按国家和城市统计平均速度
	fmt.Println("按国家/城市统计（成功请求的平均耗时）:")
	countryStats := make(map[string][]time.Duration) // 国家 -> 耗时列表
	cityStats := make(map[string][]time.Duration)    // 城市 -> 耗时列表

	for _, result := range results {
		if result != nil && result.Success {
			if result.Country != "" {
				countryStats[result.Country] = append(countryStats[result.Country], result.Duration)
			}
			if result.City != "" {
				cityKey := fmt.Sprintf("%s/%s", result.Country, result.City)
				cityStats[cityKey] = append(cityStats[cityKey], result.Duration)
			}
		}
	}

	// 按平均耗时排序国家
	type CountryStat struct {
		Country     string
		AvgDuration time.Duration
		Count       int
	}
	var countryList []CountryStat
	for country, durations := range countryStats {
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		countryList = append(countryList, CountryStat{
			Country:     country,
			AvgDuration: total / time.Duration(len(durations)),
			Count:       len(durations),
		})
	}

	// 排序（按平均耗时升序，最快的在前）
	for i := 0; i < len(countryList)-1; i++ {
		for j := i + 1; j < len(countryList); j++ {
			if countryList[i].AvgDuration > countryList[j].AvgDuration {
				countryList[i], countryList[j] = countryList[j], countryList[i]
			}
		}
	}

	if len(countryList) > 0 {
		fmt.Printf("  国家排名（按平均速度，共 %d 个国家）:\n", len(countryList))
		for i, stat := range countryList {
			fmt.Printf("    %2d. %-30s | 平均耗时: %v | 成功数: %d\n",
				i+1, stat.Country, stat.AvgDuration, stat.Count)
		}
		fmt.Println()
	}

	// 按平均耗时排序城市
	type CityStat struct {
		City        string
		AvgDuration time.Duration
		Count       int
	}
	var cityList []CityStat
	for city, durations := range cityStats {
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		cityList = append(cityList, CityStat{
			City:        city,
			AvgDuration: total / time.Duration(len(durations)),
			Count:       len(durations),
		})
	}

	// 排序（按平均耗时升序，最快的在前）
	for i := 0; i < len(cityList)-1; i++ {
		for j := i + 1; j < len(cityList); j++ {
			if cityList[i].AvgDuration > cityList[j].AvgDuration {
				cityList[i], cityList[j] = cityList[j], cityList[i]
			}
		}
	}

	if len(cityList) > 0 {
		fmt.Printf("  城市排名（按平均速度，共 %d 个城市）:\n", len(cityList))
		for i, stat := range cityList {
			fmt.Printf("    %2d. %-30s | 平均耗时: %v | 成功数: %d\n",
				i+1, stat.City, stat.AvgDuration, stat.Count)
		}
		fmt.Println()
	}

	// 9. 显示所有请求的完整列表（按耗时从快到慢排序）
	fmt.Printf("所有请求详情（按耗时从快到慢排序，共 %d 个请求）:\n", totalIPs)
	sortedResults := sortByDuration(results, true)

	for i, result := range sortedResults {
		if result == nil {
			continue
		}
		status := "❌"
		if result.Success {
			status = "✅"
		}

		countryCity := ""
		if result.Country != "" {
			countryCity = fmt.Sprintf("%s/%s", result.Country, result.City)
			if result.City == "" {
				countryCity = result.Country
			}
		}

		if countryCity != "" {
			if result.Error != nil {
				fmt.Printf("  %s [%3d/%3d] IP: %-15s | %-30s | 耗时: %v | 错误: %v\n",
					status, i+1, len(sortedResults), result.IP, countryCity, result.Duration, result.Error)
			} else {
				fmt.Printf("  %s [%3d/%3d] IP: %-15s | %-30s | 状态码: %3d | 耗时: %v\n",
					status, i+1, len(sortedResults), result.IP, countryCity, result.StatusCode, result.Duration)
			}
		} else {
			if result.Error != nil {
				fmt.Printf("  %s [%3d/%3d] IP: %-15s | 耗时: %v | 错误: %v\n",
					status, i+1, len(sortedResults), result.IP, result.Duration, result.Error)
			} else {
				fmt.Printf("  %s [%3d/%3d] IP: %-15s | 状态码: %3d | 耗时: %v\n",
					status, i+1, len(sortedResults), result.IP, result.StatusCode, result.Duration)
			}
		}
	}

	fmt.Println("\n✅ 测试完成")
}

// calculateAvgDuration 计算平均耗时（仅统计成功的请求）
func calculateAvgDuration(results []*RequestResult, successOnly bool) time.Duration {
	var total time.Duration
	var count int

	for _, result := range results {
		if result != nil && (!successOnly || result.Success) {
			total += result.Duration
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return total / time.Duration(count)
}

// sortByDuration 按耗时排序
func sortByDuration(results []*RequestResult, ascending bool) []*RequestResult {
	sorted := make([]*RequestResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			sorted = append(sorted, result)
		}
	}

	// 简单排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if (ascending && sorted[i].Duration > sorted[j].Duration) ||
				(!ascending && sorted[i].Duration < sorted[j].Duration) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
