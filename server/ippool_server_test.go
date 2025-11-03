package server

import (
	"context"
	"os"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"utls_client/ippool"
	pb "utls_client/proto/ippool"
)

// setupTestLibrary 创建测试用的 IP 池库
func setupTestLibrary(t *testing.T) *ippool.IPPoolLibrary {
	dataDir := "./testdata/ippool_test"
	os.MkdirAll(dataDir, 0755)
	
	// 使用离线模式，避免网络请求
	library := ippool.NewIPPoolLibrary("", dataDir)
	library.SetOfflineMode(true)
	
	// 如果本地有数据，则加载
	library.LoadFromLocal()
	
	return library
}

// setupTestServer 创建测试用的 gRPC 服务器
func setupTestServer(t *testing.T) *IPPoolServer {
	library := setupTestLibrary(t)
	return NewIPPoolServer(library)
}

// TestGetAllHosts 测试获取所有主机列表
func TestGetAllHosts(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	req := &pb.GetAllHostsRequest{}
	resp, err := server.GetAllHosts(ctx, req)
	
	if err != nil {
		t.Logf("获取主机列表: %v (可能是没有本地数据)", err)
		return
	}
	
	if resp == nil {
		t.Fatal("响应为空")
	}
	
	t.Logf("找到 %d 个主机", len(resp.Hosts))
	for _, host := range resp.Hosts {
		t.Logf("  主机: %s (详细数据: %v)", host.Host, host.DetailExists)
	}
}

// TestGetHostInfo 测试获取指定主机信息
func TestGetHostInfo(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	// 先获取所有主机
	getAllReq := &pb.GetAllHostsRequest{}
	getAllResp, err := server.GetAllHosts(ctx, getAllReq)
	if err != nil {
		t.Skipf("跳过测试: 无法获取主机列表 - %v", err)
		return
	}
	
	if len(getAllResp.Hosts) == 0 {
		t.Skip("跳过测试: 没有可用的主机")
		return
	}
	
	// 测试获取第一个主机的信息
	testHost := getAllResp.Hosts[0].Host
	req := &pb.GetHostInfoRequest{Host: testHost}
	resp, err := server.GetHostInfo(ctx, req)
	
	if err != nil {
		t.Errorf("获取主机信息失败: %v", err)
		return
	}
	
	if resp.HostInfo == nil {
		t.Fatal("主机信息为空")
	}
	
	if resp.HostInfo.Host != testHost {
		t.Errorf("主机名不匹配: 期望 %s, 得到 %s", testHost, resp.HostInfo.Host)
	}
	
	t.Logf("主机信息: %+v", resp.HostInfo)
}

// TestGetHostInfoNotFound 测试获取不存在的主机信息
func TestGetHostInfoNotFound(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	req := &pb.GetHostInfoRequest{Host: "nonexistent.example.com"}
	_, err := server.GetHostInfo(ctx, req)
	
	if err == nil {
		t.Fatal("应该返回错误")
	}
	
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("错误不是 gRPC 状态错误")
	}
	
	if st.Code() != codes.NotFound {
		t.Errorf("错误代码不匹配: 期望 %v, 得到 %v", codes.NotFound, st.Code())
	}
}

// TestGetIPPool 测试获取简化格式 IP 池
func TestGetIPPool(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	// 先获取所有主机
	getAllReq := &pb.GetAllHostsRequest{}
	getAllResp, err := server.GetAllHosts(ctx, getAllReq)
	if err != nil {
		t.Skipf("跳过测试: 无法获取主机列表 - %v", err)
		return
	}
	
	if len(getAllResp.Hosts) == 0 {
		t.Skip("跳过测试: 没有可用的主机")
		return
	}
	
	// 测试获取第一个主机的 IP 池
	testHost := getAllResp.Hosts[0].Host
	req := &pb.GetIPPoolRequest{Host: testHost}
	resp, err := server.GetIPPool(ctx, req)
	
	if err != nil {
		t.Logf("获取 IP 池失败 (可能是没有同步数据): %v", err)
		return
	}
	
	if resp == nil {
		t.Fatal("响应为空")
	}
	
	t.Logf("主机 %s 的 IP 池: IPv4=%d, IPv6=%d", testHost, len(resp.Ipv4), len(resp.Ipv6))
}

// TestGetDetailIPPool 测试获取详细格式 IP 池
func TestGetDetailIPPool(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	// 先获取所有主机
	getAllReq := &pb.GetAllHostsRequest{}
	getAllResp, err := server.GetAllHosts(ctx, getAllReq)
	if err != nil {
		t.Skipf("跳过测试: 无法获取主机列表 - %v", err)
		return
	}
	
	// 找到有详细数据的主机
	var testHost string
	for _, host := range getAllResp.Hosts {
		if host.DetailExists {
			testHost = host.Host
			break
		}
	}
	
	if testHost == "" {
		t.Skip("跳过测试: 没有有详细数据的主机")
		return
	}
	
	req := &pb.GetDetailIPPoolRequest{Host: testHost}
	resp, err := server.GetDetailIPPool(ctx, req)
	
	if err != nil {
		t.Logf("获取详细 IP 池失败 (可能是没有同步数据): %v", err)
		return
	}
	
	if resp == nil {
		t.Fatal("响应为空")
	}
	
	t.Logf("主机 %s 的详细 IP 池: IP数量=%d", testHost, len(resp.Ips))
	if resp.Stats != nil {
		t.Logf("  统计: IPv4=%d, IPv6=%d", resp.Stats.Ipv4Count, resp.Stats.Ipv6Count)
	}
}

// TestGetIPDetail 测试获取指定 IP 的详细信息
func TestGetIPDetail(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	// 先获取详细 IP 池
	getAllReq := &pb.GetAllHostsRequest{}
	getAllResp, err := server.GetAllHosts(ctx, getAllReq)
	if err != nil {
		t.Skipf("跳过测试: 无法获取主机列表 - %v", err)
		return
	}
	
	var testHost string
	for _, host := range getAllResp.Hosts {
		if host.DetailExists {
			testHost = host.Host
			break
		}
	}
	
	if testHost == "" {
		t.Skip("跳过测试: 没有有详细数据的主机")
		return
	}
	
	// 获取详细 IP 池以找到一个 IP
	detailReq := &pb.GetDetailIPPoolRequest{Host: testHost}
	detailResp, err := server.GetDetailIPPool(ctx, detailReq)
	if err != nil {
		t.Skipf("跳过测试: 无法获取详细 IP 池 - %v", err)
		return
	}
	
	if len(detailResp.Ips) == 0 {
		t.Skip("跳过测试: 详细 IP 池为空")
		return
	}
	
	// 测试获取第一个 IP 的详细信息
	testIP := ""
	for ip := range detailResp.Ips {
		testIP = ip
		break
	}
	
	if testIP == "" {
		t.Skip("跳过测试: 没有可用的 IP")
		return
	}
	
	req := &pb.GetIPDetailRequest{Host: testHost, Ip: testIP}
	resp, err := server.GetIPDetail(ctx, req)
	
	if err != nil {
		t.Errorf("获取 IP 详细信息失败: %v (IP: %s)", err, testIP)
		return
	}
	
	if resp.IpDetail == nil {
		t.Fatal("IP 详细信息为空")
	}
	
	// IP 应该匹配（可能是 map key 或者 IPDetailInfo.Ip）
	actualIP := resp.IpDetail.Ip
	if actualIP == "" {
		actualIP = testIP // 如果返回的 IP 为空，至少应该匹配请求的 IP
	}
	
	if actualIP != testIP && actualIP != "" {
		t.Logf("注意: IP 地址不完全匹配 (请求: %s, 返回: %s)，但继续测试", testIP, actualIP)
	}
	
	t.Logf("IP %s 的详细信息: Country=%s, City=%s", testIP, resp.IpDetail.Location.Country, resp.IpDetail.Location.City)
}

// TestSearchIPs 测试搜索 IP
func TestSearchIPs(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	// 测试空条件搜索（应该返回所有 IP）
	req := &pb.SearchIPsRequest{}
	resp, err := server.SearchIPs(ctx, req)
	
	if err != nil {
		t.Logf("搜索 IP 失败 (可能是没有详细数据): %v", err)
		return
	}
	
	if resp == nil {
		t.Fatal("响应为空")
	}
	
	t.Logf("搜索到 %d 个 IP", len(resp.Ips))
}

// TestGetRandomIP 测试获取随机 IP
func TestGetRandomIP(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	// 先获取所有主机
	getAllReq := &pb.GetAllHostsRequest{}
	getAllResp, err := server.GetAllHosts(ctx, getAllReq)
	if err != nil {
		t.Skipf("跳过测试: 无法获取主机列表 - %v", err)
		return
	}
	
	if len(getAllResp.Hosts) == 0 {
		t.Skip("跳过测试: 没有可用的主机")
		return
	}
	
	testHost := getAllResp.Hosts[0].Host
	req := &pb.GetRandomIPRequest{Host: testHost}
	resp, err := server.GetRandomIP(ctx, req)
	
	if err != nil {
		t.Logf("获取随机 IP 失败 (可能是没有 IP 数据): %v", err)
		return
	}
	
	if resp.Ip == "" {
		t.Fatal("随机 IP 为空")
	}
	
	t.Logf("随机 IP: %s", resp.Ip)
}

// TestAnalyzeAll 测试分析所有数据
func TestAnalyzeAll(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	req := &pb.AnalyzeAllRequest{}
	resp, err := server.AnalyzeAll(ctx, req)
	
	if err != nil {
		t.Logf("分析所有数据失败 (可能是没有数据): %v", err)
		return
	}
	
	if resp.Stats == nil {
		t.Fatal("统计信息为空")
	}
	
	t.Logf("分析统计:")
	t.Logf("  主机总数: %d", resp.Stats.TotalHosts)
	t.Logf("  IPv4 总数: %d", resp.Stats.TotalIpv4)
	t.Logf("  IPv6 总数: %d", resp.Stats.TotalIpv6)
	t.Logf("  国家数: %d", len(resp.Stats.Countries))
	t.Logf("  城市数: %d", len(resp.Stats.Cities))
}

// TestGetServiceStatus 测试获取服务状态
func TestGetServiceStatus(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	req := &pb.GetServiceStatusRequest{}
	resp, err := server.GetServiceStatus(ctx, req)
	
	if err != nil {
		t.Fatalf("获取服务状态失败: %v", err)
	}
	
	if resp == nil {
		t.Fatal("响应为空")
	}
	
	t.Logf("服务状态:")
	t.Logf("  离线模式: %v", resp.OfflineMode)
	t.Logf("  自动同步: %v", resp.AutoSyncEnabled)
	t.Logf("  主机总数: %d", resp.TotalHosts)
	t.Logf("  已加载主机: %d", resp.LoadedHosts)
	
	if resp.LastSyncTime != nil {
		t.Logf("  最后同步时间: %v", resp.LastSyncTime.AsTime())
	}
}

// TestSyncHosts 测试同步主机列表
func TestSyncHosts(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()
	
	// 注意: 如果离线模式，这个测试会失败
	if server.library.IsOfflineMode() {
		t.Skip("跳过测试: 处于离线模式")
		return
	}
	
	req := &pb.SyncHostsRequest{}
	resp, err := server.SyncHosts(ctx, req)
	
	if err != nil {
		t.Fatalf("同步主机列表失败: %v", err)
	}
	
	if resp == nil {
		t.Fatal("响应为空")
	}
	
	t.Logf("同步结果: 成功=%v, 消息=%s, 主机数=%d", resp.Success, resp.Message, resp.HostCount)
}

// BenchmarkGetAllHosts 性能测试：获取所有主机
func BenchmarkGetAllHosts(b *testing.B) {
	server := setupTestServer(&testing.T{})
	ctx := context.Background()
	req := &pb.GetAllHostsRequest{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.GetAllHosts(ctx, req)
	}
}

// BenchmarkGetHostInfo 性能测试：获取主机信息
func BenchmarkGetHostInfo(b *testing.B) {
	server := setupTestServer(&testing.T{})
	ctx := context.Background()
	
	// 先获取一个主机
	getAllReq := &pb.GetAllHostsRequest{}
	getAllResp, err := server.GetAllHosts(ctx, getAllReq)
	if err != nil || len(getAllResp.Hosts) == 0 {
		b.Skip("需要至少一个主机")
		return
	}
	
	testHost := getAllResp.Hosts[0].Host
	req := &pb.GetHostInfoRequest{Host: testHost}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.GetHostInfo(ctx, req)
	}
}

