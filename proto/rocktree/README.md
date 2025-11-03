# RockTree Proto 定义说明

## 概述

RockTree.proto 文件定义了 Google Earth 中用于表示 3D 地球数据的 Protocol Buffer 消息结构。它主要用于在客户端和服务器之间传输地理空间数据，包括地形、纹理和元数据等。

## 生成 Go 代码

### 1. 安装依赖

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### 2. 生成代码

```bash
# 从项目根目录执行
protoc --go_out=. --go_opt=paths=source_relative \
       proto/rocktree/RockTree.proto
```

**注意**：此 proto 文件使用 proto2 语法，且不包含 gRPC 服务定义，因此只需要生成 protobuf 消息代码，无需生成 gRPC 代码。

### 3. 安装 Protobuf 依赖

```bash
go get google.golang.org/protobuf
```

### 4. 验证生成结果

```bash
# 验证生成的代码可以编译
go build ./proto/rocktree
```

## 消息类型详解

### TileKeyBounds
定义瓦片的边界范围，包含层级、行号和列号的最小最大值。
- `level`: 瓦片层级
- `min_row`, `max_row`: 行号范围
- `min_column`, `max_column`: 列号范围

### KmlCoordinate
表示 KML 格式的坐标点。
- `latitude`: 纬度
- `longitude`: 经度
- `altitude`: 海拔高度

### ViewportMetadataRequest/ViewportMetadata
用于请求和响应视口元数据，支持视口内的数据加载优化。
- `epoch`: 时间戳
- `tile_key_bounds`: 瓦片键边界范围列表
- `omit_ancestors`: 是否忽略祖先节点
- `bulk_metadata_response_mode`: 批量元数据响应模式

### BulkMetadataRequest
批量元数据请求消息。
- `node_key`: 节点键

### NodeDataRequest
节点数据请求消息。
- `node_key`: 节点键
- `texture_format`: 纹理格式
- `imagery_epoch`: 影像纪元
- `omit_texture`: 是否忽略纹理
- `date`, `milliseconds`: 日期和毫秒
- `texture_is_shared`: 纹理是否共享

### NodeKey
节点键，用于唯一标识一个节点。
- `path`: 路径
- `epoch`: 纪元

### RockTreeTasks
RockTree 任务结构。
- `head_node_key`: 头节点键
- `node_metadata_tasks`: 节点元数据任务列表
- `default_imagery_epoch`: 默认影像纪元
- `default_available_texture_formats`: 默认可用纹理格式

### NodeMetadataTasks
节点元数据任务。
- `path_and_flags`: 路径和标志
- `epoch`: 纪元
- `bulk_metadata_epoch`: 批量元数据纪元
- `imagery_epoch`: 影像纪元
- `available_texture_formats`: 可用纹理格式

### CopyrightRequest
版权请求消息。
- `epoch`: 纪元

### TextureDataRequest
纹理数据请求消息。
- `node_key`: 节点键
- `texture_format`: 纹理格式
- `view_direction`: 视图方向
- `imagery_epoch`: 影像纪元

### BulkMetadata
批量元数据，包含多个节点的元数据信息。
- `node_metadata`: 节点元数据列表
- `head_node_key`: 头节点键值
- `head_node_center`: 头节点中心坐标
- `meters_per_texel`: 每个纹理像素对应的米数
- `default_imagery_epoch`: 默认图像纪元
- `default_available_texture_formats`: 默认可用的纹理格式
- `default_available_view_dependent_textures`: 默认可用的视图相关纹理
- `default_available_view_dependent_texture_formats`: 默认可用的视图相关纹理格式
- `common_dated_nodes`: 共享的日期节点列表
- `default_acquisition_date_range`: 默认采集日期范围

### NodeMetadata
存储空间节点的元数据信息，包含几何属性、时间戳、纹理格式等关键参数。
- `path_and_flags`: 路径标识组合字段
- `epoch`: 元数据版本时间戳
- `bulk_metadata_epoch`: 批量处理时的元数据时间戳
- `oriented_bounding_box`: 定向包围盒数据
- `meters_per_texel`: 空间分辨率
- `processing_oriented_bounding_box`: 处理专用包围盒
- `imagery_epoch`: 影像数据时间戳
- `available_texture_formats`: 可用纹理格式掩码
- `available_view_dependent_textures`: 视角相关纹理数量
- `available_view_dependent_texture_formats`: 视角相关纹理格式掩码
- `dated_nodes`: 时间关联子节点列表
- `acquisition_date_range`: 数据采集时间范围
- `Flags`: 元数据标志位定义

### DatedNode
带日期的节点。
- `date`: 日期
- `milliseconds`: 毫秒
- `epoch`: 纪元
- `shared_epoch`: 共享纪元
- `coarse_level`: 粗略层级

### AcquisitionDate
采集日期。
- `date`: 日期
- `milliseconds`: 毫秒

### AcquisitionDateRange
采集日期范围。
- `begin`: 开始日期
- `end`: 结束日期

### NodeData
节点的实际几何数据。
- `matrix_globe_from_mesh`: 从网格到地球的变换矩阵
- `meshes`: 网格列表
- `copyright_ids`: 版权ID列表
- `node_key`: 节点键
- `kml_bounding_box`: KML边界框
- `water_mesh`: 水面网格
- `overlay_surface_meshes`: 覆盖面网格列表
- `normal_table`: 法线表

### Mesh
网格数据结构。
- `vertices`: 顶点数据
- `vertex_alphas`: 顶点透明度
- `texture_coords`: 纹理坐标
- `indices`: 索引数据
- `octant_ranges`: 八分体范围
- `layer_counts`: 图层计数
- `texture`: 纹理列表
- `texture_coordinates`: 纹理坐标
- `uv_offset_and_scale`: UV偏移和缩放
- `layer_and_octant_counts`: 图层和八分体计数
- `normals`: 法线数据
- `normals_dev`: 法线偏差
- `mesh_id`: 网格ID
- `skirt_flags`: 裙边标志
- `Layer`: 图层类型枚举
- `LayerMask`: 图层掩码枚举

### Texture
纹理数据。
- `data`: 纹理数据
- `format`: 纹理格式
- `width`, `height`: 宽度和高度
- `view_direction`: 视图方向
- `mesh_id`: 网格ID
- `measurement_data`: 质量测量数据
- `Format`: 纹理格式枚举
- `ViewDirection`: 视图方向枚举

### QualityMeasurements
质量测量。
- `psnr`: 峰值信噪比

### TextureData
纹理数据。
- `node_key`: 节点键
- `textures`: 纹理列表
- `transform_info`: 变换信息列表
- `projection_origin`: 投影原点

### Copyrights
版权集合。
- `copyrights`: 版权列表

### Copyright
版权信息。
- `id`: ID
- `text`: 文本
- `text_clean`: 清理后的文本

### PlanetoidMetadata
行星体元数据。
- `root_node_metadata`: 根节点元数据
- `radius`: 半径
- `min_terrain_altitude`: 最小地形高度
- `max_terrain_altitude`: 最大地形高度
- `max_imagery_epoch`: 最大影像纪元

## 工作原理

RockTree 是 Google Earth 中用于组织和传输 3D 地球数据的数据结构。它采用树状结构组织地理空间数据，每个节点代表地球表面的一个区域。

### 层次细节 (Level of Detail)
通过 `meters_per_texel` 字段控制细节层次，根据观察距离决定加载的数据精度。

### 空间划分
使用 `processing_oriented_bounding_box` 进行视锥剔除，只加载可见区域的数据。

### 时间管理
通过 epoch 字段管理数据版本，支持数据更新和时间轴功能。

### 纹理优化
支持多种压缩格式和视角相关的纹理，提高渲染效率。

### 批量传输
使用 `BulkMetadata` 机制减少网络请求次数，提高传输效率。

## 使用场景

1. Google Earth 客户端与服务器之间的数据传输
2. 3D 地理空间数据的存储和管理
3. LOD (Level of Detail) 系统实现
4. 大规模地理数据的流式加载