package fingerprint

import (
	"fmt"
	"math/rand"
	"time"

	utls "github.com/refraction-networking/utls"
)

// FingerprintProfile 表示一个完整的浏览器指纹配置
type FingerprintProfile struct {
	Name        string             // 指纹名称
	HelloID     utls.ClientHelloID // uTLS 指纹ID
	UserAgent   string             // 对应的 User-Agent
	Description string             // 描述
	Platform    string             // 平台类型
	Browser     string             // 浏览器类型
	Version     string             // 版本号
}

// FingerprintLibrary 指纹库管理器
type FingerprintLibrary struct {
	profiles []FingerprintProfile
	rand     *rand.Rand
}

// NewFingerprintLibrary 创建新的指纹库
func NewFingerprintLibrary() *FingerprintLibrary {
	lib := &FingerprintLibrary{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	lib.initProfiles()
	return lib
}

// initProfiles 初始化所有浏览器指纹配置
func (lib *FingerprintLibrary) initProfiles() {
	lib.profiles = []FingerprintProfile{
		// Chrome 系列
		{
			Name:        "Chrome 133 - Windows",
			HelloID:     utls.HelloChrome_133,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			Description: "Chrome 133 on Windows 10/11",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "133",
		},
		{
			Name:        "Chrome 133 - macOS",
			HelloID:     utls.HelloChrome_133,
			UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			Description: "Chrome 133 on macOS",
			Platform:    "macOS",
			Browser:     "Chrome",
			Version:     "133",
		},
		{
			Name:        "Chrome 131 - Windows",
			HelloID:     utls.HelloChrome_131,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			Description: "Chrome 131 on Windows 10/11",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "131",
		},
		{
			Name:        "Chrome 131 - macOS",
			HelloID:     utls.HelloChrome_131,
			UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			Description: "Chrome 131 on macOS",
			Platform:    "macOS",
			Browser:     "Chrome",
			Version:     "131",
		},
		{
			Name:        "Chrome 120 - Windows",
			HelloID:     utls.HelloChrome_120,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Description: "Chrome 120 on Windows 10/11",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "120",
		},
		{
			Name:        "Chrome 120 - Linux",
			HelloID:     utls.HelloChrome_120,
			UserAgent:   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Description: "Chrome 120 on Linux",
			Platform:    "Linux",
			Browser:     "Chrome",
			Version:     "120",
		},
		{
			Name:        "Chrome 115 PQ - Windows",
			HelloID:     utls.HelloChrome_115_PQ,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
			Description: "Chrome 115 with Post-Quantum on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "115-PQ",
		},
		{
			Name:        "Chrome 114 - Windows",
			HelloID:     utls.HelloChrome_114_Padding_PSK_Shuf,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
			Description: "Chrome 114 with advanced features on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "114",
		},
		{
			Name:        "Chrome 112 - Windows",
			HelloID:     utls.HelloChrome_112_PSK_Shuf,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
			Description: "Chrome 112 with PSK shuffle on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "112",
		},
		{
			Name:        "Chrome 106 Shuffle - Windows",
			HelloID:     utls.HelloChrome_106_Shuffle,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
			Description: "Chrome 106 with shuffle on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "106",
		},
		{
			Name:        "Chrome 102 - Windows",
			HelloID:     utls.HelloChrome_102,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36",
			Description: "Chrome 102 on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "102",
		},
		{
			Name:        "Chrome 100 - Windows",
			HelloID:     utls.HelloChrome_100,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.0.0 Safari/537.36",
			Description: "Chrome 100 on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "100",
		},
		{
			Name:        "Chrome 96 - Windows",
			HelloID:     utls.HelloChrome_96,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.0.0 Safari/537.36",
			Description: "Chrome 96 on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "96",
		},
		{
			Name:        "Chrome 87 - Windows",
			HelloID:     utls.HelloChrome_87,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.0.0 Safari/537.36",
			Description: "Chrome 87 on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "87",
		},
		{
			Name:        "Chrome 83 - Windows",
			HelloID:     utls.HelloChrome_83,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.0.0 Safari/537.36",
			Description: "Chrome 83 on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "83",
		},
		{
			Name:        "Chrome Auto - Windows",
			HelloID:     utls.HelloChrome_Auto,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			Description: "Chrome latest on Windows",
			Platform:    "Windows",
			Browser:     "Chrome",
			Version:     "auto",
		},

		// Firefox 系列
		{
			Name:        "Firefox 120 - Windows",
			HelloID:     utls.HelloFirefox_120,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
			Description: "Firefox 120 on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "120",
		},
		{
			Name:        "Firefox 120 - macOS",
			HelloID:     utls.HelloFirefox_120,
			UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0",
			Description: "Firefox 120 on macOS",
			Platform:    "macOS",
			Browser:     "Firefox",
			Version:     "120",
		},
		{
			Name:        "Firefox 105 - Windows",
			HelloID:     utls.HelloFirefox_105,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:105.0) Gecko/20100101 Firefox/105.0",
			Description: "Firefox 105 on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "105",
		},
		{
			Name:        "Firefox 102 - Windows",
			HelloID:     utls.HelloFirefox_102,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0",
			Description: "Firefox 102 on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "102",
		},
		{
			Name:        "Firefox 99 - Windows",
			HelloID:     utls.HelloFirefox_99,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:99.0) Gecko/20100101 Firefox/99.0",
			Description: "Firefox 99 on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "99",
		},
		{
			Name:        "Firefox 65 - Windows",
			HelloID:     utls.HelloFirefox_65,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:65.0) Gecko/20100101 Firefox/65.0",
			Description: "Firefox 65 on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "65",
		},
		{
			Name:        "Firefox 63 - Windows",
			HelloID:     utls.HelloFirefox_63,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:63.0) Gecko/20100101 Firefox/63.0",
			Description: "Firefox 63 on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "63",
		},
		{
			Name:        "Firefox 56 - Windows",
			HelloID:     utls.HelloFirefox_56,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
			Description: "Firefox 56 on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "56",
		},
		{
			Name:        "Firefox 55 - Windows",
			HelloID:     utls.HelloFirefox_55,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:55.0) Gecko/20100101 Firefox/55.0",
			Description: "Firefox 55 on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "55",
		},
		{
			Name:        "Firefox Auto - Windows",
			HelloID:     utls.HelloFirefox_Auto,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
			Description: "Firefox latest on Windows",
			Platform:    "Windows",
			Browser:     "Firefox",
			Version:     "auto",
		},

		// Edge 系列
		{
			Name:        "Edge 106 - Windows",
			HelloID:     utls.HelloEdge_106,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 Edg/106.0.0.0",
			Description: "Edge 106 on Windows",
			Platform:    "Windows",
			Browser:     "Edge",
			Version:     "106",
		},
		{
			Name:        "Edge 85 - Windows",
			HelloID:     utls.HelloEdge_85,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.0.0 Safari/537.36 Edg/85.0.0.0",
			Description: "Edge 85 on Windows",
			Platform:    "Windows",
			Browser:     "Edge",
			Version:     "85",
		},
		{
			Name:        "Edge Auto - Windows",
			HelloID:     utls.HelloEdge_Auto,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
			Description: "Edge latest on Windows",
			Platform:    "Windows",
			Browser:     "Edge",
			Version:     "auto",
		},

		// Safari 系列 (macOS)
		{
			Name:        "Safari 17 - macOS",
			HelloID:     utls.HelloSafari_Auto,
			UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			Description: "Safari on macOS",
			Platform:    "macOS",
			Browser:     "Safari",
			Version:     "17",
		},

		// iOS Safari
		{
			Name:        "iOS Safari 14 - iPhone",
			HelloID:     utls.HelloIOS_14,
			UserAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.2 Mobile/15E148 Safari/604.1",
			Description: "iOS Safari 14 on iPhone",
			Platform:    "iOS",
			Browser:     "Safari",
			Version:     "14",
		},
		{
			Name:        "iOS Safari 13 - iPhone",
			HelloID:     utls.HelloIOS_13,
			UserAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 13_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.2 Mobile/15E148 Safari/604.1",
			Description: "iOS Safari 13 on iPhone",
			Platform:    "iOS",
			Browser:     "Safari",
			Version:     "13",
		},
		{
			Name:        "iOS Safari 12 - iPhone",
			HelloID:     utls.HelloIOS_12_1,
			UserAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 12_5_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1.2 Mobile/15E148 Safari/604.1",
			Description: "iOS Safari 12 on iPhone",
			Platform:    "iOS",
			Browser:     "Safari",
			Version:     "12",
		},

		// 随机指纹（使用通用 User-Agent）
		{
			Name:        "Randomized - Chrome Like",
			HelloID:     utls.HelloRandomized,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			Description: "Randomized fingerprint with Chrome-like UA",
			Platform:    "Random",
			Browser:     "Random",
			Version:     "random",
		},
		{
			Name:        "Randomized ALPN - Chrome Like",
			HelloID:     utls.HelloRandomizedALPN,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			Description: "Randomized fingerprint with ALPN",
			Platform:    "Random",
			Browser:     "Random",
			Version:     "random",
		},
		{
			Name:        "Randomized No ALPN - Firefox Like",
			HelloID:     utls.HelloRandomizedNoALPN,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
			Description: "Randomized fingerprint without ALPN",
			Platform:    "Random",
			Browser:     "Random",
			Version:     "random",
		},
	}
}

// GetAllProfiles 获取所有指纹配置
func (lib *FingerprintLibrary) GetAllProfiles() []FingerprintProfile {
	return lib.profiles
}

// GetRandomProfile 随机获取一个指纹配置
func (lib *FingerprintLibrary) GetRandomProfile() FingerprintProfile {
	return lib.profiles[lib.rand.Intn(len(lib.profiles))]
}

// GetProfileByName 根据名称获取指纹配置
func (lib *FingerprintLibrary) GetProfileByName(name string) (*FingerprintProfile, error) {
	for _, profile := range lib.profiles {
		if profile.Name == name {
			return &profile, nil
		}
	}
	return nil, fmt.Errorf("未找到指纹: %s", name)
}

// GetProfilesByBrowser 根据浏览器类型获取指纹列表
func (lib *FingerprintLibrary) GetProfilesByBrowser(browser string) []FingerprintProfile {
	result := []FingerprintProfile{}
	for _, profile := range lib.profiles {
		if profile.Browser == browser {
			result = append(result, profile)
		}
	}
	return result
}

// GetProfilesByPlatform 根据平台类型获取指纹列表
func (lib *FingerprintLibrary) GetProfilesByPlatform(platform string) []FingerprintProfile {
	result := []FingerprintProfile{}
	for _, profile := range lib.profiles {
		if profile.Platform == platform {
			result = append(result, profile)
		}
	}
	return result
}

// GetRecommendedProfiles 获取推荐的指纹列表（最新版本）
func (lib *FingerprintLibrary) GetRecommendedProfiles() []FingerprintProfile {
	recommended := []FingerprintProfile{}

	// 推荐最新版本的指纹
	for _, profile := range lib.profiles {
		if profile.Version == "133" || profile.Version == "131" ||
			profile.Version == "120" || profile.Version == "auto" {
			recommended = append(recommended, profile)
		}
	}

	return recommended
}

// GetRandomProfileByBrowser 根据浏览器类型随机获取一个指纹
func (lib *FingerprintLibrary) GetRandomProfileByBrowser(browser string) (*FingerprintProfile, error) {
	profiles := lib.GetProfilesByBrowser(browser)
	if len(profiles) == 0 {
		return nil, fmt.Errorf("未找到浏览器类型: %s", browser)
	}
	profile := profiles[lib.rand.Intn(len(profiles))]
	return &profile, nil
}

// GetRandomProfileByPlatform 根据平台类型随机获取一个指纹
func (lib *FingerprintLibrary) GetRandomProfileByPlatform(platform string) (*FingerprintProfile, error) {
	profiles := lib.GetProfilesByPlatform(platform)
	if len(profiles) == 0 {
		return nil, fmt.Errorf("未找到平台类型: %s", platform)
	}
	profile := profiles[lib.rand.Intn(len(profiles))]
	return &profile, nil
}

// GetSafeProfiles 获取安全的指纹列表（避免被检测）
func (lib *FingerprintLibrary) GetSafeProfiles() []FingerprintProfile {
	safeProfiles := []FingerprintProfile{}

	// 优先选择最新版本和随机指纹
	for _, profile := range lib.profiles {
		if profile.Browser == "Firefox" ||
			profile.Browser == "Random" ||
			profile.Version == "133" ||
			profile.Version == "131" {
			safeProfiles = append(safeProfiles, profile)
		}
	}

	return safeProfiles
}

// PrintAllProfiles 打印所有指纹配置
func (lib *FingerprintLibrary) PrintAllProfiles() {
	fmt.Println("=== 可用的浏览器指纹配置 ===")
	for i, profile := range lib.profiles {
		fmt.Printf("\n[%d] %s\n", i+1, profile.Name)
		fmt.Printf("    描述: %s\n", profile.Description)
		fmt.Printf("    平台: %s | 浏览器: %s | 版本: %s\n",
			profile.Platform, profile.Browser, profile.Version)
		fmt.Printf("    User-Agent: %s\n", profile.UserAgent)
		fmt.Printf("    指纹ID: %s\n", profile.HelloID.Client)
	}
}

// PrintProfilesByBrowser 按浏览器分类打印指纹
func (lib *FingerprintLibrary) PrintProfilesByBrowser() {
	browsers := map[string][]FingerprintProfile{}

	for _, profile := range lib.profiles {
		browsers[profile.Browser] = append(browsers[profile.Browser], profile)
	}

	fmt.Println("=== 按浏览器分类的指纹 ===")
	for browser, profiles := range browsers {
		fmt.Printf("\n## %s (%d 个)\n", browser, len(profiles))
		for i, profile := range profiles {
			fmt.Printf("  [%d] %s - %s\n", i+1, profile.Name, profile.Version)
		}
	}
}
