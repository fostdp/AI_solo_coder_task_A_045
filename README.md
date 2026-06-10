# 古代战场遗址空间分布与军事地理分析系统

全栈GIS分析平台，基于Go + PostgreSQL/PostGIS + Leaflet/Canvas，对春秋战国至明清800个战场遗址进行空间建模与军事地理分区。

## 系统架构

```
                    ┌─────────────────────────────────────────┐
                    │            Nginx (:80)                  │
                    │  Gzip压缩 / 静态缓存 / API反向代理      │
                    └────────────┬────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌────────▼────────┐  ┌───────────▼──────────┐  ┌────────▼────────┐
│  Go API (:8080) │  │  Prometheus (:9090)  │  │  Grafana (:3000)│
│  pprof (:6060)  │  │  指标采集&存储       │  │  可视化仪表盘   │
│  Gin + gonum    │  └───────────┬──────────┘  └─────────────────┘
│  模型参数外置   │              │
└────────┬────────┘              │
         │                       │
┌────────▼──────────────────────▼─────────────────────────────────┐
│              PostgreSQL + PostGIS (:5432)                        │
│  GiST空间索引 / RASTER栅格 / 7张核心表 / 性能优化配置           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────┐
│  Simulator (一次性运行)      │
│  CLI参数化数据生成器         │
│  → /data/data.json          │
└─────────────────────────────┘
```

## 快速部署

### Docker Compose 一键启动

```bash
# 构建并启动全部服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f api

# 停止
docker-compose down
```

启动后访问：
| 服务 | 地址 | 用途 |
|------|------|------|
| 前端 | http://localhost | 古战场GIS界面 |
| API | http://localhost:8080/api/statistics | REST API |
| pprof | http://localhost:6060/debug/pprof | 性能分析 |
| Prometheus | http://localhost:9091 | 指标查询 |
| Grafana | http://localhost:3000 | 监控仪表盘 (admin/密码见.env) |
| PostgreSQL | localhost:5432 | 数据库直连 |

### 单独运行（无Docker）

```bash
# 1. 生成模拟数据
go run ./cmd/data_generator -count 800 -output ./web/data/data.json -v

# 2. 启动API服务
go run ./cmd/server -config config.yaml -port 8080

# 3. 前端（开发模式）
cd web && python -m http.server 8081
```

## 模拟器用法

模拟器 `cmd/data_generator` 支持CLI参数化生成不同配置的战场数据：

```bash
# 默认：800个战场，全年代全地形
go run ./cmd/data_generator -count 800 -output ./web/data/data.json

# 仅生成秦汉时期战场
go run ./cmd/data_generator -count 200 -era "秦汉" -output ./web/data/qinhan.json

# 仅生成山地关隘战场
go run ./cmd/data_generator -count 500 -terrain "山地,关隘" -output ./web/data/mountain.json

# 高分辨率DEM（0.5度格网）
go run ./cmd/data_generator -count 800 -dem-resolution 0.5 -output ./web/data/hires.json

# 可复现的随机种子
go run ./cmd/data_generator -count 800 -seed 42 -output ./web/data/reproducible.json

# 详细输出（年代/地形分布统计）
go run ./cmd/data_generator -count 800 -v
```

### CLI参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-count` | 800 | 战场数量 |
| `-era` | all | 年代筛选: all, 春秋战国, 秦汉, 三国两晋南北朝, 隋唐五代, 宋辽金元, 明清 |
| `-terrain` | all | 地形筛选: all, 山地, 平原, 河谷, 关隘 |
| `-dem-resolution` | 1.0 | DEM分辨率(度)，越小精度越高 |
| `-roads` | 60 | 道路数量 |
| `-rivers` | 25 | 河流数量 |
| `-seed` | 0 | 随机种子(0=当前时间) |
| `-output` | ./web/data/data.json | 输出路径 |
| `-v` | false | 详细输出（分布统计） |

### DEM高程模型

5区海拔基线 + 随机噪声，模拟中国地形：

| 区域 | 经纬度条件 | 基线海拔 |
|------|-----------|---------|
| 青藏高原 | lng < 95° | 3500m |
| 横断山脉 | lat > 30°, lng < 105° | 2000m |
| 黄土高原 | lat > 40°, lng < 110° | 1200m |
| 东南丘陵 | lat < 25° | 200m |
| 中部过渡带 | 其余 | 600m |

## 模型参数配置

[config.yaml](config.yaml) 外置47个参数，无需重新编译：

```yaml
logistic_regression:
  learning_rate: 0.0001    # 梯度下降学习率
  epochs: 5000             # 最大迭代次数
  tolerance: 1e-7          # 收敛阈值

bootstrap:
  runs: 100                # 重抽样次数
  confidence_level: 0.95   # 置信水平

background_sampling:
  default_type: target_group  # target_group / random
  target_group_bandwidth_deg: 5.0  # 高斯核带宽
  num_background_points: 1000     # 背景点数量

clustering:
  default_num_regions: 8   # 默认分区数
  fcm_fuzzifier: 2.0       # 模糊化系数m
  fcm_max_iter: 100        # FCM最大迭代
  fcm_convergence_eps: 0.0001  # 收敛阈值
  troops_scale_factor: 10000.0  # 兵力归一化因子

high_prob_area:
  grid_lng_step_deg: 1.0   # 格网步长
  threshold: 0.6            # 概率阈值
```

## API端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/battlefields` | 战场列表 |
| GET | `/api/battlefields/:id` | 战场详情 |
| GET | `/api/roads` | 古代道路 |
| GET | `/api/rivers` | 河流水系 |
| GET | `/api/dem` | DEM高程数据 |
| GET | `/api/terrain_profile` | 地形剖面(参数: start_lng, start_lat, end_lng, end_lat, num_points) |
| GET | `/api/accessibility/:id` | 交通可达性 |
| GET | `/api/site_selection_factors` | 选址因素(参数: background, bootstrap) |
| GET | `/api/enhanced_lr` | 增强逻辑回归结果 |
| GET | `/api/high_prob_areas` | 高概率战场区域 |
| GET | `/api/military_regions` | 军事地理分区(参数: num_regions, fuzzy) |
| GET | `/api/fuzzy_cluster` | 模糊聚类结果 |
| GET | `/api/statistics` | 统计概览 |

## 监控

### pprof性能分析

```bash
# CPU profile (30秒采样)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# goroutine分析
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Prometheus指标

`http://localhost:9090/metrics` 暴露以下指标：

| 指标 | 类型 | 说明 |
|------|------|------|
| `battlefield_http_requests_total` | Counter | HTTP请求总数(method/path/status) |
| `battlefield_http_request_duration_seconds` | Histogram | 请求耗时分布 |
| `battlefield_data_battlefields_total` | Gauge | 战场数据总量 |
| `battlefield_data_regions_total` | Gauge | 分区数量 |
| `battlefield_analysis_runs_total` | Counter | 分析执行次数(type) |

### PostgreSQL性能配置

[sql/performance.sql](sql/performance.sql) 优化项：

- `shared_buffers`: 256MB
- `effective_cache_size`: 768MB
- `work_mem`: 16MB
- PostGIS栅格驱动: GTiff
- 并行查询: 4 workers

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.21 + Gin + gonum + Prometheus client |
| 数据库 | PostgreSQL 16 + PostGIS 3.4 |
| 前端 | Leaflet.js + HTML5 Canvas + OffscreenCanvas |
| 容器 | Docker + docker-compose |
| 反向代理 | Nginx (Gzip + 缓存) |
| 监控 | Prometheus + Grafana + pprof |

## 项目结构

```
.
├── cmd/
│   ├── server/           # API服务入口
│   └── data_generator/   # 模拟器
├── pkg/
│   ├── config/           # 配置加载(YAML)
│   ├── models/           # 数据模型
│   ├── handlers/         # HTTP处理器
│   ├── battlefield_loader/  # 战场数据导入
│   ├── terrain_analyzer/    # MaxEnt/逻辑回归
│   ├── geo_partitioner/     # FCM聚类/分区
│   └── metrics/          # Prometheus指标
├── sql/
│   ├── init.sql          # Schema (7表+GiST)
│   └── performance.sql   # PostgreSQL性能配置
├── web/
│   ├── index.html
│   ├── css/style.css
│   └── js/
│       ├── data.js            # 数据加载+fallback
│       ├── battlefield_map.js # 地图/Canvas/交互
│       ├── terrain_profile.js # 离屏渲染+节流
│       ├── charts.js          # 统计图表
│       └── app.js             # 业务协调
├── nginx/default.conf        # Gzip+缓存+反向代理
├── monitoring/
│   ├── prometheus.yml
│   ├── grafana-datasources/
│   └── grafana-dashboards/
├── config.yaml               # 模型参数
├── Dockerfile                # 多阶段构建
├── docker-compose.yml        # 服务编排
└── .env                      # 环境变量
```
