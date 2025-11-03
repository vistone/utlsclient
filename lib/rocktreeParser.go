package utls_client

// PathAndFlagsResult 表示解包后的路径与标志位结果
// 用于 NodeMetadataTasks.path_and_flags 字段的解析结果
type PathAndFlagsResult struct {
	Path  string // 路径字符串（由数字字符 '0'-'7' 组成）
	Level int    // 路径层级（1-4，从低2位提取：1 + (pathID & 3)）
	Flags uint64 // 标志位（剩余的高位）
}

// NodeMetadataPathAndFlagsResult 表示 NodeMetadata.path_and_flags 字段的解析结果
// 用于 NodeMetadata.path_and_flags 字段（与 NodeMetadataTasks 编码格式不同）
type NodeMetadataPathAndFlagsResult struct {
	PathLength uint32 // 路径长度（单位：字节，高24位）
	Flags      uint32 // 标志位（低8位，参考 NodeMetadata.Flags 枚举）
}

// UnpackPathAndFlags 解包 NodeMetadataTasks.path_and_flags 字段
// 编码格式：
//   - 低2位：层级（level = 1 + (pathID & 3)），表示路径字符数量
//   - 然后每3位：路径字符（从 '0' 开始：'0' + (pathID & 7)），共 level 个字符
//   - 剩余位：标志位
//
// 示例：
//   - pathID = 0x12345678
//   - level = 1 + (0x78 & 3) = 1 + 0 = 1（最低2位是 00）
//   - pathID >>= 2 = 0x48D159E
//   - 提取1个路径字符：(0x48D159E & 7) = 6，字符为 '6'
//   - flags = 0x48D159E >> 3 = 0x91A2B3C
//
// 对应 proto 定义：proto/rocktree/RockTree.proto 中的 NodeMetadataTasks.path_and_flags
func UnpackPathAndFlags(pathID uint64) PathAndFlagsResult {
	// 提取层级（最低 2 位）
	level := 1 + int(pathID&3)
	pathID >>= 2

	// 提取路径（每层 3 位，以字符 '0' 开始累加）
	var pathRunes []rune
	for range level {
		digit := rune('0') + rune(pathID&7)
		pathRunes = append(pathRunes, digit)
		pathID >>= 3
	}

	// 剩余位为标志位
	flags := pathID

	return PathAndFlagsResult{
		Path:  string(pathRunes),
		Level: level,
		Flags: flags,
	}
}

// UnpackNodeMetadataPathAndFlags 解包 NodeMetadata.path_and_flags 字段
// 编码格式：
//   - 高24位（bit 8-31）：路径长度（单位：字节）
//   - 低8位（bit 0-7）：标志位（参考 NodeMetadata.Flags 枚举）
//
// 标志位枚举值（参考 proto/rocktree/RockTree.proto NodeMetadata.Flags）：
//   - RICH3D_LEAF = 1
//   - RICH3D_NODATA = 2
//   - LEAF = 4
//   - NODATA = 8
//   - USE_IMAGERY_EPOCH = 16
//
// 对应 proto 定义：proto/rocktree/RockTree.proto 中的 NodeMetadata.path_and_flags
func UnpackNodeMetadataPathAndFlags(pathAndFlags uint32) NodeMetadataPathAndFlagsResult {
	// 低8位为标志位
	flags := pathAndFlags & 0xFF

	// 高24位为路径长度
	pathLength := (pathAndFlags >> 8) & 0xFFFFFF

	return NodeMetadataPathAndFlagsResult{
		PathLength: pathLength,
		Flags:      flags,
	}
}

// NodeMetadataFlags 定义 NodeMetadata 的标志位常量
// 对应 proto/rocktree/RockTree.proto 中的 NodeMetadata.Flags 枚举
const (
	// FlagRich3DLeaf 富几何细节的叶子节点
	FlagRich3DLeaf uint32 = 1
	// FlagRich3DNoData 无效节点（保留结构但无数据）
	FlagRich3DNoData uint32 = 2
	// FlagLeaf 基础叶子节点标识
	FlagLeaf uint32 = 4
	// FlagNoData 无有效数据节点
	FlagNoData uint32 = 8
	// FlagUseImageryEpoch 使用独立影像时间戳
	FlagUseImageryEpoch uint32 = 16
)

// HasFlag 检查 NodeMetadata 标志位是否包含指定标志
func (r NodeMetadataPathAndFlagsResult) HasFlag(flag uint32) bool {
	return (r.Flags & flag) != 0
}
