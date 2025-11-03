package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"utls_client/ippool"
	pb "utls_client/proto/ippool"
)

// IPPoolServer gRPC 服务器实现
type IPPoolServer struct {
	pb.UnimplementedIPPoolServiceServer
	library  *ippool.IPPoolLibrary
	analyzer *ippool.Analyzer
}

// NewIPPoolServer 创建新的 gRPC 服务器实例
func NewIPPoolServer(library *ippool.IPPoolLibrary) *IPPoolServer {
	return &IPPoolServer{
		library:  library,
		analyzer: ippool.NewAnalyzer(library),
	}
}

// GetAllHosts 获取所有主机列表
func (s *IPPoolServer) GetAllHosts(ctx context.Context, req *pb.GetAllHostsRequest) (*pb.GetAllHostsResponse, error) {
	hosts := s.library.GetAllHosts()
	pbHosts := make([]*pb.HostInfo, len(hosts))
	for i, host := range hosts {
		pbHosts[i] = &pb.HostInfo{
			Host:         host.Host,
			FileName:     host.FileName,
			DetailFile:   host.DetailFile,
			Url:          host.URL,
			DetailUrl:    host.DetailURL,
			Exists:       host.Exists,
			DetailExists: host.DetailExists,
		}
	}
	return &pb.GetAllHostsResponse{Hosts: pbHosts}, nil
}

// GetHostInfo 获取指定主机信息
func (s *IPPoolServer) GetHostInfo(ctx context.Context, req *pb.GetHostInfoRequest) (*pb.GetHostInfoResponse, error) {
	hostInfo, err := s.library.GetHostInfo(req.Host)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "未找到主机: %v", err)
	}
	return &pb.GetHostInfoResponse{
		HostInfo: &pb.HostInfo{
			Host:         hostInfo.Host,
			FileName:     hostInfo.FileName,
			DetailFile:   hostInfo.DetailFile,
			Url:          hostInfo.URL,
			DetailUrl:    hostInfo.DetailURL,
			Exists:       hostInfo.Exists,
			DetailExists: hostInfo.DetailExists,
		},
	}, nil
}

// GetIPPool 获取简化格式 IP 池数据
func (s *IPPoolServer) GetIPPool(ctx context.Context, req *pb.GetIPPoolRequest) (*pb.GetIPPoolResponse, error) {
	pool, err := s.library.GetIPPool(req.Host)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "获取 IP 池失败: %v", err)
	}
	return &pb.GetIPPoolResponse{
		Ipv4: pool.IPv4,
		Ipv6: pool.IPv6,
	}, nil
}

// GetDetailIPPool 获取详细格式 IP 池数据
func (s *IPPoolServer) GetDetailIPPool(ctx context.Context, req *pb.GetDetailIPPoolRequest) (*pb.GetDetailIPPoolResponse, error) {
	detailPool, err := s.library.GetDetailIPPool(req.Host)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "获取详细 IP 池失败: %v", err)
	}

	// 转换 IP 详细信息
	pbIPs := make(map[string]*pb.IPDetailInfo)
	for ip, ipInfo := range detailPool.IPs {
		pbIPs[ip] = &pb.IPDetailInfo{
			Ip: ip,
			Location: &pb.IPLocationInfo{
				Country:    ipInfo.Location.Country,
				Region:     ipInfo.Location.Region,
				City:       ipInfo.Location.City,
				Isp:        ipInfo.Location.ISP,
				Org:        ipInfo.Location.Org,
				DataCenter: ipInfo.Location.DataCenter,
				IpType:     ipInfo.Location.IPType,
			},
		}
	}

	// 转换统计信息
	var lastUpdated *timestamppb.Timestamp
	if !detailPool.Stats.LastUpdated.IsZero() {
		lastUpdated = timestamppb.New(detailPool.Stats.LastUpdated)
	}

	return &pb.GetDetailIPPoolResponse{
		Ips: pbIPs,
		Stats: &pb.PoolStats{
			Ipv4Count:   int32(detailPool.Stats.IPv4Count),
			Ipv6Count:   int32(detailPool.Stats.IPv6Count),
			LastUpdated: lastUpdated,
		},
	}, nil
}

// GetIPDetail 获取指定 IP 的详细信息
func (s *IPPoolServer) GetIPDetail(ctx context.Context, req *pb.GetIPDetailRequest) (*pb.GetIPDetailResponse, error) {
	ipInfo, err := s.library.GetIPDetail(req.Host, req.Ip)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "获取 IP 详细信息失败: %v", err)
	}
	return &pb.GetIPDetailResponse{
		IpDetail: &pb.IPDetailInfo{
			Ip: ipInfo.IP,
			Location: &pb.IPLocationInfo{
				Country:    ipInfo.Location.Country,
				Region:     ipInfo.Location.Region,
				City:       ipInfo.Location.City,
				Isp:        ipInfo.Location.ISP,
				Org:        ipInfo.Location.Org,
				DataCenter: ipInfo.Location.DataCenter,
				IpType:     ipInfo.Location.IPType,
			},
		},
	}, nil
}

// SearchIPs 搜索 IP（支持多条件）
func (s *IPPoolServer) SearchIPs(ctx context.Context, req *pb.SearchIPsRequest) (*pb.SearchIPsResponse, error) {
	ips, err := s.analyzer.SearchIPs(req.Host, req.Country, req.City, req.Isp, req.DataCenter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "搜索 IP 失败: %v", err)
	}

	pbIPs := make([]*pb.IPDetailInfo, len(ips))
	for i, ipInfo := range ips {
		pbIPs[i] = &pb.IPDetailInfo{
			Ip: ipInfo.IP,
			Location: &pb.IPLocationInfo{
				Country:    ipInfo.Location.Country,
				Region:     ipInfo.Location.Region,
				City:       ipInfo.Location.City,
				Isp:        ipInfo.Location.ISP,
				Org:        ipInfo.Location.Org,
				DataCenter: ipInfo.Location.DataCenter,
				IpType:     ipInfo.Location.IPType,
			},
		}
	}
	return &pb.SearchIPsResponse{Ips: pbIPs}, nil
}

// GetRandomIP 获取随机 IP
func (s *IPPoolServer) GetRandomIP(ctx context.Context, req *pb.GetRandomIPRequest) (*pb.GetRandomIPResponse, error) {
	ip, err := s.analyzer.GetRandomIP(req.Host)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "获取随机 IP 失败: %v", err)
	}
	return &pb.GetRandomIPResponse{Ip: ip}, nil
}

// GetAllIPsByHost 获取指定主机的所有 IP
func (s *IPPoolServer) GetAllIPsByHost(ctx context.Context, req *pb.GetAllIPsByHostRequest) (*pb.GetAllIPsByHostResponse, error) {
	ipv4, ipv6, err := s.analyzer.GetAllIPsByHost(req.Host)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "获取主机 IP 列表失败: %v", err)
	}
	return &pb.GetAllIPsByHostResponse{
		Ipv4: ipv4,
		Ipv6: ipv6,
	}, nil
}

// AnalyzeAll 分析所有数据
func (s *IPPoolServer) AnalyzeAll(ctx context.Context, req *pb.AnalyzeAllRequest) (*pb.AnalyzeAllResponse, error) {
	stats, err := s.analyzer.AnalyzeAll()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "分析数据失败: %v", err)
	}
	return &pb.AnalyzeAllResponse{
		Stats: s.convertAnalyzeStats(stats),
	}, nil
}

// AnalyzeByHost 分析指定主机
func (s *IPPoolServer) AnalyzeByHost(ctx context.Context, req *pb.AnalyzeByHostRequest) (*pb.AnalyzeByHostResponse, error) {
	stats, err := s.analyzer.AnalyzeByHost(req.Host)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "分析主机数据失败: %v", err)
	}
	return &pb.AnalyzeByHostResponse{
		Stats: s.convertAnalyzeStats(stats),
	}, nil
}

// AnalyzeByCountry 按国家分析
func (s *IPPoolServer) AnalyzeByCountry(ctx context.Context, req *pb.AnalyzeByCountryRequest) (*pb.AnalyzeByCountryResponse, error) {
	ips, err := s.analyzer.AnalyzeByCountry(req.Country)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "按国家分析失败: %v", err)
	}
	return &pb.AnalyzeByCountryResponse{
		Ips: s.convertIPDetailInfos(ips),
	}, nil
}

// AnalyzeByCity 按城市分析
func (s *IPPoolServer) AnalyzeByCity(ctx context.Context, req *pb.AnalyzeByCityRequest) (*pb.AnalyzeByCityResponse, error) {
	ips, err := s.analyzer.AnalyzeByCity(req.City)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "按城市分析失败: %v", err)
	}
	return &pb.AnalyzeByCityResponse{
		Ips: s.convertIPDetailInfos(ips),
	}, nil
}

// AnalyzeByISP 按 ISP 分析
func (s *IPPoolServer) AnalyzeByISP(ctx context.Context, req *pb.AnalyzeByISPRequest) (*pb.AnalyzeByISPResponse, error) {
	ips, err := s.analyzer.AnalyzeByISP(req.Isp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "按 ISP 分析失败: %v", err)
	}
	return &pb.AnalyzeByISPResponse{
		Ips: s.convertIPDetailInfos(ips),
	}, nil
}

// AnalyzeByDataCenter 按数据中心分析
func (s *IPPoolServer) AnalyzeByDataCenter(ctx context.Context, req *pb.AnalyzeByDataCenterRequest) (*pb.AnalyzeByDataCenterResponse, error) {
	ips, err := s.analyzer.AnalyzeByDataCenter(req.DataCenter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "按数据中心分析失败: %v", err)
	}
	return &pb.AnalyzeByDataCenterResponse{
		Ips: s.convertIPDetailInfos(ips),
	}, nil
}

// SyncAll 同步所有数据
func (s *IPPoolServer) SyncAll(ctx context.Context, req *pb.SyncAllRequest) (*pb.SyncAllResponse, error) {
	err := s.library.SyncAll()
	if err != nil {
		return &pb.SyncAllResponse{
			Success: false,
			Message: fmt.Sprintf("同步失败: %v", err),
		}, nil
	}
	return &pb.SyncAllResponse{
		Success: true,
		Message: "同步成功",
	}, nil
}

// SyncHosts 同步主机列表
func (s *IPPoolServer) SyncHosts(ctx context.Context, req *pb.SyncHostsRequest) (*pb.SyncHostsResponse, error) {
	err := s.library.SyncHosts()
	if err != nil {
		return &pb.SyncHostsResponse{
			Success:   false,
			Message:   fmt.Sprintf("同步主机列表失败: %v", err),
			HostCount: 0,
		}, nil
	}
	hosts := s.library.GetAllHosts()
	return &pb.SyncHostsResponse{
		Success:   true,
		Message:   "同步成功",
		HostCount: int32(len(hosts)),
	}, nil
}

// SyncIPPool 同步简化格式 IP 池
func (s *IPPoolServer) SyncIPPool(ctx context.Context, req *pb.SyncIPPoolRequest) (*pb.SyncIPPoolResponse, error) {
	err := s.library.SyncIPPool(req.Host)
	if err != nil {
		return &pb.SyncIPPoolResponse{
			Success:   false,
			Message:   fmt.Sprintf("同步 IP 池失败: %v", err),
			Ipv4Count: 0,
			Ipv6Count: 0,
		}, nil
	}
	pool, err := s.library.GetIPPool(req.Host)
	if err != nil {
		return &pb.SyncIPPoolResponse{
			Success:   true,
			Message:   "同步成功，但获取数据失败",
			Ipv4Count: 0,
			Ipv6Count: 0,
		}, nil
	}
	return &pb.SyncIPPoolResponse{
		Success:   true,
		Message:   "同步成功",
		Ipv4Count: int32(len(pool.IPv4)),
		Ipv6Count: int32(len(pool.IPv6)),
	}, nil
}

// SyncDetailIPPool 同步详细格式 IP 池
func (s *IPPoolServer) SyncDetailIPPool(ctx context.Context, req *pb.SyncDetailIPPoolRequest) (*pb.SyncDetailIPPoolResponse, error) {
	err := s.library.SyncDetailIPPool(req.Host, req.Force)
	if err != nil {
		return &pb.SyncDetailIPPoolResponse{
			Success: false,
			Message: fmt.Sprintf("同步详细 IP 池失败: %v", err),
			IpCount: 0,
		}, nil
	}
	detailPool, err := s.library.GetDetailIPPool(req.Host)
	if err != nil {
		return &pb.SyncDetailIPPoolResponse{
			Success: true,
			Message: "同步成功，但获取数据失败",
			IpCount: 0,
		}, nil
	}
	return &pb.SyncDetailIPPoolResponse{
		Success: true,
		Message: "同步成功",
		IpCount: int32(len(detailPool.IPs)),
	}, nil
}

// GetCountriesList 获取所有国家列表及其统计信息
func (s *IPPoolServer) GetCountriesList(ctx context.Context, req *pb.GetCountriesListRequest) (*pb.GetCountriesListResponse, error) {
	countries, err := s.analyzer.GetCountriesList()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取国家列表失败: %v", err)
	}

	// 转换 map[string]int 到 map[string]int32
	pbCountries := make(map[string]int32)
	for country, count := range countries {
		pbCountries[country] = int32(count)
	}

	return &pb.GetCountriesListResponse{Countries: pbCountries}, nil
}

// GetCitiesByCountry 获取指定国家的所有城市列表及其统计信息
func (s *IPPoolServer) GetCitiesByCountry(ctx context.Context, req *pb.GetCitiesByCountryRequest) (*pb.GetCitiesByCountryResponse, error) {
	cities, err := s.analyzer.GetCitiesByCountry(req.Country)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取城市列表失败: %v", err)
	}

	// 转换 map[string]int 到 map[string]int32
	pbCities := make(map[string]int32)
	for city, count := range cities {
		pbCities[city] = int32(count)
	}

	return &pb.GetCitiesByCountryResponse{Cities: pbCities}, nil
}

// GetIPsByCountryAndCity 获取指定国家和城市的所有 IP
func (s *IPPoolServer) GetIPsByCountryAndCity(ctx context.Context, req *pb.GetIPsByCountryAndCityRequest) (*pb.GetIPsByCountryAndCityResponse, error) {
	ips, err := s.analyzer.GetIPsByCountryAndCity(req.Country, req.City)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取 IP 列表失败: %v", err)
	}

	return &pb.GetIPsByCountryAndCityResponse{
		Ips: s.convertIPDetailInfos(ips),
	}, nil
}

// GetServiceStatus 获取服务状态
func (s *IPPoolServer) GetServiceStatus(ctx context.Context, req *pb.GetServiceStatusRequest) (*pb.GetServiceStatusResponse, error) {
	lastSyncTime := s.library.GetLastSyncTime()
	var pbLastSyncTime *timestamppb.Timestamp
	if !lastSyncTime.IsZero() {
		pbLastSyncTime = timestamppb.New(lastSyncTime)
	}
	hosts := s.library.GetAllHosts()
	return &pb.GetServiceStatusResponse{
		OfflineMode:     s.library.IsOfflineMode(),
		AutoSyncEnabled: s.library.IsAutoSyncEnabled(),
		LastSyncTime:    pbLastSyncTime,
		TotalHosts:      int32(len(hosts)),
		LoadedHosts:     int32(len(hosts)),
	}, nil
}

// convertAnalyzeStats 转换分析统计信息
func (s *IPPoolServer) convertAnalyzeStats(stats *ippool.AnalyzeStats) *pb.AnalyzeStats {
	if stats == nil {
		return &pb.AnalyzeStats{}
	}

	// 转换 map[string]int 到 map[string]int32
	convertIntMap := func(m map[string]int) map[string]int32 {
		result := make(map[string]int32)
		for k, v := range m {
			result[k] = int32(v)
		}
		return result
	}

	return &pb.AnalyzeStats{
		TotalHosts:  int32(stats.TotalHosts),
		TotalIpv4:   int32(stats.TotalIPv4),
		TotalIpv6:   int32(stats.TotalIPv6),
		Countries:   convertIntMap(stats.Countries),
		Cities:      convertIntMap(stats.Cities),
		Regions:     convertIntMap(stats.Regions),
		Isps:        convertIntMap(stats.ISPs),
		Orgs:        convertIntMap(stats.Orgs),
		DataCenters: convertIntMap(stats.DataCenters),
		IpTypes:     convertIntMap(stats.IPTypes),
	}
}

// convertIPDetailInfos 转换 IP 详细信息列表
func (s *IPPoolServer) convertIPDetailInfos(ips []*ippool.IPDetailInfo) []*pb.IPDetailInfo {
	result := make([]*pb.IPDetailInfo, len(ips))
	for i, ipInfo := range ips {
		result[i] = &pb.IPDetailInfo{
			Ip: ipInfo.IP,
			Location: &pb.IPLocationInfo{
				Country:    ipInfo.Location.Country,
				Region:     ipInfo.Location.Region,
				City:       ipInfo.Location.City,
				Isp:        ipInfo.Location.ISP,
				Org:        ipInfo.Location.Org,
				DataCenter: ipInfo.Location.DataCenter,
				IpType:     ipInfo.Location.IPType,
			},
		}
	}
	return result
}
