//go:build ignore

package main

import (
	"fmt"
	"sort"
	"utls_client/ippool"
)

func main() {
	fmt.Println("=== 按国家和城市分析 IP 池 ===")
	fmt.Println()

	// 创建 IP 池库（使用离线模式）
	library := ippool.NewIPPoolLibrary("", "./ippool_data")
	library.SetOfflineMode(true)
	defer library.Close()

	// 加载本地数据
	library.LoadFromLocal()

	// 创建分析器
	analyzer := ippool.NewAnalyzer(library)

	// 1. 获取所有国家列表
	fmt.Println("1. 所有国家列表（按 IP 数量排序）：")
	countries, err := analyzer.GetCountriesList()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 排序：按 IP 数量降序
	type countryStat struct {
		country string
		count   int
	}
	var countryList []countryStat
	for country, count := range countries {
		countryList = append(countryList, countryStat{country, count})
	}
	sort.Slice(countryList, func(i, j int) bool {
		return countryList[i].count > countryList[j].count
	})

	for i, cs := range countryList {
		if i >= 10 { // 只显示前 10 个
			fmt.Printf("   ... 还有 %d 个国家\n", len(countryList)-10)
			break
		}
		fmt.Printf("   %-30s: %4d 个 IP\n", cs.country, cs.count)
	}
	fmt.Printf("   总计: %d 个国家\n\n", len(countryList))

	// 2. 选择一个国家，查看其城市列表
	if len(countryList) > 0 {
		testCountry := countryList[0].country
		fmt.Printf("2. 国家 '%s' 的所有城市列表：\n", testCountry)

		cities, err := analyzer.GetCitiesByCountry(testCountry)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			return
		}

		// 排序城市列表
		var cityList []countryStat
		for city, count := range cities {
			cityList = append(cityList, countryStat{city, count})
		}
		sort.Slice(cityList, func(i, j int) bool {
			return cityList[i].count > cityList[j].count
		})

		for i, cs := range cityList {
			if i >= 15 { // 只显示前 15 个
				fmt.Printf("   ... 还有 %d 个城市\n", len(cityList)-15)
				break
			}
			fmt.Printf("   %-30s: %4d 个 IP\n", cs.country, cs.count)
		}
		fmt.Printf("   总计: %d 个城市\n\n", len(cityList))

		// 3. 选择一个城市，查看该城市的所有 IP
		if len(cityList) > 0 {
			testCity := cityList[0].country
			fmt.Printf("3. 国家 '%s' / 城市 '%s' 的所有 IP（前 10 个）：\n", testCountry, testCity)

			ips, err := analyzer.GetIPsByCountryAndCity(testCountry, testCity)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				return
			}

			fmt.Printf("   总计: %d 个 IP\n", len(ips))
			for i, ipInfo := range ips {
				if i >= 10 {
					fmt.Printf("   ... 还有 %d 个 IP\n", len(ips)-10)
					break
				}
				fmt.Printf("   %-20s | %s, %s\n", ipInfo.IP, ipInfo.Location.Country, ipInfo.Location.City)
			}
		}
	}

	fmt.Println()
	fmt.Println("✅ 分析完成")
}
