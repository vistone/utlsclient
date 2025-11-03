package main

import (
	"fmt"
	"utls_client/fingerprint"
)

func mainFingerprintDemo() {
	// 创建指纹库
	lib := fingerprint.NewFingerprintLibrary()

	// 打印所有指纹
	fmt.Println("========== 所有指纹配置 ==========")
	lib.PrintProfilesByBrowser()

	// 随机获取一个指纹
	fmt.Println("\n========== 随机指纹示例 ==========")
	randomProfile := lib.GetRandomProfile()
	fmt.Printf("随机选择: %s\n", randomProfile.Name)
	fmt.Printf("User-Agent: %s\n", randomProfile.UserAgent)

	// 获取推荐指纹
	fmt.Println("\n========== 推荐指纹 ==========")
	recommended := lib.GetRecommendedProfiles()
	for i, profile := range recommended {
		if i >= 5 {
			break
		}
		fmt.Printf("%s (平台: %s, 版本: %s)\n",
			profile.Name, profile.Platform, profile.Version)
	}

	// 根据浏览器获取
	fmt.Println("\n========== Firefox 指纹 ==========")
	firefoxProfiles := lib.GetProfilesByBrowser("Firefox")
	for _, profile := range firefoxProfiles {
		fmt.Printf("%s\n", profile.Name)
	}

	// 获取安全指纹
	fmt.Println("\n========== 安全指纹推荐 ==========")
	safeProfiles := lib.GetSafeProfiles()
	for i, profile := range safeProfiles {
		if i >= 10 {
			break
		}
		fmt.Printf("%s - %s\n", profile.Name, profile.Browser)
	}

	// 按平台获取
	fmt.Println("\n========== Windows 平台指纹 ==========")
	windowsProfiles := lib.GetProfilesByPlatform("Windows")
	for i, profile := range windowsProfiles {
		if i >= 5 {
			break
		}
		fmt.Printf("%s - %s\n", profile.Name, profile.Browser)
	}

	// 按名称查找
	fmt.Println("\n========== 查找测试 ==========")
	profile, err := lib.GetProfileByName("Chrome 133 - Windows")
	if err == nil {
		fmt.Printf("找到: %s\n", profile.Name)
		fmt.Printf("User-Agent: %s\n", profile.UserAgent)
	} else {
		fmt.Printf("错误: %v\n", err)
	}

	// 随机浏览器指纹
	fmt.Println("\n========== 随机 Chrome 指纹 ==========")
	chromeProfile, err := lib.GetRandomProfileByBrowser("Chrome")
	if err == nil {
		fmt.Printf("随机 Chrome: %s\n", chromeProfile.Name)
	} else {
		fmt.Printf("错误: %v\n", err)
	}
}
