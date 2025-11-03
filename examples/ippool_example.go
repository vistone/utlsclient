//go:build ignore

package main

import (
	"fmt"
	"time"
	"utls_client/ippool"
)

func main() {
	fmt.Println("=== IP 池库测试 ===")
	fmt.Println()

	// 创建 IP 池库（指定数据存储目录）
	library := ippool.NewIPPoolLibrary("http://tile0.zeromaps.cn:9005", "./ippool_data")
	defer library.Close()

	// 显示本地数据信息
	fmt.Println("0. 本地数据信息...")
	info := library.GetLocalDataInfo()
	fmt.Printf("   数据目录: %s\n", info["data_dir"])
	fmt.Printf("   主机文件存在: %v\n", info["hosts_file_exists"])
	if modified, ok := info["hosts_file_modified"]; ok {
		fmt.Printf("   主机文件修改时间: %s\n", modified)
	}
	fmt.Printf("   IP池文件数量: %v\n", info["ip_pool_files_count"])
	fmt.Println()

	// 1. 同步主机列表
	fmt.Println("1. 同步主机列表...")
	fmt.Println("   正在连接服务器，请稍候...")
	if err := library.SyncHosts(); err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		fmt.Println("   提示: 如果超时，可能是网络问题，请检查网络连接")
		return
	}
	fmt.Println("✅ 主机列表同步成功")

	// 显示所有主机
	hosts := library.GetAllHosts()
	fmt.Printf("   找到 %d 个主机:\n", len(hosts))
	for _, host := range hosts {
		fmt.Printf("   - %s (简化: %v, 详细: %v)\n", host.Host, host.Exists, host.DetailExists)
	}
	fmt.Println()

	// 2. 同步所有主机的 IP 池数据
	fmt.Println("2. 同步所有主机的数据...")

	// 同步所有主机
	for i, host := range hosts {
		testHost := host.Host
		fmt.Printf("   主机 %d/%d: %s\n", i+1, len(hosts), testHost)

		// 同步简化格式
		fmt.Println("   同步简化格式...")
		if err := library.SyncIPPool(testHost); err != nil {
			fmt.Printf("   ❌ 错误: %v\n", err)
		} else {
			fmt.Println("   ✅ 简化格式同步成功")

			// 获取并显示 IP 池数据
			pool, err := library.GetIPPool(testHost)
			if err == nil {
				fmt.Printf("   IPv4 数量: %d\n", len(pool.IPv4))
				if len(pool.IPv4) > 0 {
					fmt.Printf("   前5个 IPv4: %v\n", pool.IPv4[:min(5, len(pool.IPv4))])
				}
				if len(pool.IPv6) > 0 {
					fmt.Printf("   IPv6 数量: %d\n", len(pool.IPv6))
				}
			}
		}

		// 同步详细格式
		if host.DetailExists {
			fmt.Println("     同步详细格式...")
			// 强制同步，确保所有数据都下载
			if err := library.SyncDetailIPPool(testHost, true); err != nil {
				fmt.Printf("     ❌ 错误: %v\n", err)
			} else {
				fmt.Println("     ✅ 详细格式同步成功")

				// 获取详细数据
				detailPool, err := library.GetDetailIPPool(testHost)
				if err == nil {
					fmt.Printf("     统计: IPv4=%d, IPv6=%d, 最后更新=%s\n",
						detailPool.Stats.IPv4Count,
						detailPool.Stats.IPv6Count,
						detailPool.Stats.LastUpdated.Format(time.RFC3339))
				}
			}
		}
		fmt.Println()
	}

	// 使用 SyncAll 确保所有数据都同步（后台进行，不阻塞）
	fmt.Println("2b. 使用 SyncAll 同步所有数据（后台进行）...")
	fmt.Println("   （注：如果数据已是最新，会快速跳过）")
	if err := library.SyncAll(); err != nil {
		fmt.Printf("   ⚠️ 警告: %v\n", err)
	} else {
		fmt.Println("   ✅ 同步检查完成（如果数据已是最新则已跳过更新）")
	}
	fmt.Println()

	// 3. 使用分析器进行分析
	fmt.Println("3. 使用分析器分析数据...")
	analyzer := ippool.NewAnalyzer(library)

	// 分析指定主机
	if len(hosts) > 0 {
		testHost := hosts[0].Host
		fmt.Printf("   分析主机: %s\n", testHost)

		stats, err := analyzer.AnalyzeByHost(testHost)
		if err == nil {
			fmt.Printf("   IPv4 总数: %d\n", stats.TotalIPv4)
			fmt.Printf("   IPv6 总数: %d\n", stats.TotalIPv6)

			if len(stats.Countries) > 0 {
				fmt.Printf("   国家分布 (前5个):\n")
				count := 0
				for country, num := range stats.Countries {
					if count >= 5 {
						break
					}
					fmt.Printf("     - %s: %d 个 IP\n", country, num)
					count++
				}
			}

			if len(stats.Cities) > 0 {
				fmt.Printf("   城市分布 (前5个):\n")
				count := 0
				for city, num := range stats.Cities {
					if count >= 5 {
						break
					}
					fmt.Printf("     - %s: %d 个 IP\n", city, num)
					count++
				}
			}

			if len(stats.DataCenters) > 0 {
				fmt.Printf("   数据中心分布:\n")
				for dc, num := range stats.DataCenters {
					fmt.Printf("     - %s: %d 个 IP\n", dc, num)
				}
			}
		}
	}
	fmt.Println()

	// 4. 测试搜索功能
	fmt.Println("4. 测试搜索功能...")
	if len(hosts) > 0 {
		testHost := hosts[0].Host
		fmt.Printf("   搜索 %s 中 Google LLC 的 IP...\n", testHost)

		results, err := analyzer.SearchIPs(testHost, "", "", "Google LLC", "")
		if err == nil && len(results) > 0 {
			fmt.Printf("   找到 %d 个匹配的 IP\n", len(results))
			if len(results) > 0 {
				fmt.Printf("   示例 IP: %s (国家: %s, 城市: %s)\n",
					results[0].IP,
					results[0].Location.Country,
					results[0].Location.City)
			}
		}
	}
	fmt.Println()

	// 5. 测试自动同步（可选）
	fmt.Println("5. 测试自动同步功能...")
	fmt.Println("   启动自动同步（每30秒一次，运行10秒后停止）...")

	if err := library.StartAutoSync(30 * time.Second); err != nil {
		fmt.Printf("   ❌ 启动自动同步失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 自动同步已启动")
		fmt.Println("   等待 10 秒...")
		time.Sleep(10 * time.Second)
		library.StopAutoSync()
		fmt.Println("   ✅ 自动同步已停止")
	}
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
	fmt.Printf("最后同步时间: %s\n", library.GetLastSyncTime().Format(time.RFC3339))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
