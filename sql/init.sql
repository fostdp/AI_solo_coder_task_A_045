-- 古代战场遗址空间分布与军事地理分析系统 - 数据库初始化脚本
-- PostgreSQL + PostGIS

-- 启用PostGIS扩展
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_raster;

-- ==================== 战场遗址表 ====================
DROP TABLE IF EXISTS battlefield CASCADE;
CREATE TABLE battlefield (
    id SERIAL PRIMARY KEY,
    battle_name VARCHAR(200) NOT NULL,
    dynasty VARCHAR(100) NOT NULL,
    era VARCHAR(50) NOT NULL,
    battle_year INT NOT NULL,
    belligerent_a VARCHAR(200) NOT NULL,
    belligerent_b VARCHAR(200) NOT NULL,
    troop_a INT NOT NULL DEFAULT 0,
    troop_b INT NOT NULL DEFAULT 0,
    total_troops INT NOT NULL DEFAULT 0,
    terrain_type VARCHAR(20) NOT NULL CHECK (terrain_type IN ('山地', '平原', '河谷', '关隘')),
    result VARCHAR(200) NOT NULL,
    geom GEOMETRY(Point, 4326) NOT NULL,
    elevation NUMERIC(10, 2) DEFAULT 0,
    distance_to_river NUMERIC(10, 2) DEFAULT 0,
    distance_to_road NUMERIC(10, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_battlefield_geom ON battlefield USING GIST(geom);
CREATE INDEX idx_battlefield_era ON battlefield(era);
CREATE INDEX idx_battlefield_terrain ON battlefield(terrain_type);
CREATE INDEX idx_battlefield_total_troops ON battlefield(total_troops);

-- ==================== 古代交通道路表 ====================
DROP TABLE IF EXISTS ancient_road CASCADE;
CREATE TABLE ancient_road (
    id SERIAL PRIMARY KEY,
    road_name VARCHAR(200) NOT NULL,
    road_type VARCHAR(50) NOT NULL DEFAULT '驿道' CHECK (road_type IN ('驿道', '栈道', '漕运', '官道', '古道')),
    dynasty VARCHAR(100) NOT NULL,
    importance INT NOT NULL DEFAULT 1 CHECK (importance BETWEEN 1 AND 5),
    geom GEOMETRY(LineString, 4326) NOT NULL
);

CREATE INDEX idx_ancient_road_geom ON ancient_road USING GIST(geom);
CREATE INDEX idx_ancient_road_dynasty ON ancient_road(dynasty);

-- ==================== 河流水系表 ====================
DROP TABLE IF EXISTS ancient_river CASCADE;
CREATE TABLE ancient_river (
    id SERIAL PRIMARY KEY,
    river_name VARCHAR(200) NOT NULL,
    river_type VARCHAR(50) NOT NULL DEFAULT '河流' CHECK (river_type IN ('河流', '湖泊', '运河')),
    geom GEOMETRY(LineString, 4326) NOT NULL
);

CREATE INDEX idx_ancient_river_geom ON ancient_river USING GIST(geom);

-- ==================== DEM地形栅格数据表 ====================
DROP TABLE IF EXISTS dem_tile CASCADE;
CREATE TABLE dem_tile (
    id SERIAL PRIMARY KEY,
    tile_x INT NOT NULL,
    tile_y INT NOT NULL,
    zoom INT NOT NULL,
    rast RASTER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dem_tile_rast ON dem_tile USING GIST(ST_ConvexHull(rast));
CREATE UNIQUE INDEX idx_dem_tile_xyz ON dem_tile(tile_x, tile_y, zoom);

-- ==================== 军事地理分区表 ====================
DROP TABLE IF EXISTS military_region CASCADE;
CREATE TABLE military_region (
    id SERIAL PRIMARY KEY,
    region_name VARCHAR(200) NOT NULL,
    region_code VARCHAR(50) NOT NULL,
    battle_count INT NOT NULL DEFAULT 0,
    avg_density NUMERIC(10, 4) NOT NULL DEFAULT 0,
    dominant_terrain VARCHAR(50),
    geom GEOMETRY(Polygon, 4326) NOT NULL
);

CREATE INDEX idx_military_region_geom ON military_region USING GIST(geom);

-- ==================== 高概率战场区域表 ====================
DROP TABLE IF EXISTS high_prob_area CASCADE;
CREATE TABLE high_prob_area (
    id SERIAL PRIMARY KEY,
    probability NUMERIC(5, 4) NOT NULL CHECK (probability BETWEEN 0 AND 1),
    terrain_factor NUMERIC(5, 4) NOT NULL DEFAULT 0,
    road_factor NUMERIC(5, 4) NOT NULL DEFAULT 0,
    river_factor NUMERIC(5, 4) NOT NULL DEFAULT 0,
    geom GEOMETRY(Polygon, 4326) NOT NULL
);

CREATE INDEX idx_high_prob_area_geom ON high_prob_area USING GIST(geom);
CREATE INDEX idx_high_prob_area_prob ON high_prob_area(probability);

-- ==================== 选址影响因素分析结果表 ====================
DROP TABLE IF EXISTS site_selection_factor CASCADE;
CREATE TABLE site_selection_factor (
    id SERIAL PRIMARY KEY,
    factor_name VARCHAR(100) NOT NULL,
    contribution NUMERIC(5, 4) NOT NULL DEFAULT 0,
    p_value NUMERIC(10, 6) NOT NULL DEFAULT 1,
    odds_ratio NUMERIC(10, 4) NOT NULL DEFAULT 1,
    method VARCHAR(50) NOT NULL DEFAULT '逻辑回归',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
