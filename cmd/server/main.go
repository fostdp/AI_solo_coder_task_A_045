package main

import (
	"flag"
	"log"
	"net/http"

	"ancient-battlefield/pkg/config"
	"ancient-battlefield/pkg/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	port := flag.String("port", "8080", "服务端口")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("加载配置失败: %v, 使用默认配置", err)
		cfg = &config.DefaultConfig
	}
	log.Printf("配置已加载: Bootstrap=%d, LR lr=%.6f, 聚类数=%d",
		cfg.Bootstrap.Runs, cfg.LogisticRegression.LearningRate, cfg.Clustering.DefaultK)

	h := handlers.New(cfg)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	r.Static("/", "./web")

	api := r.Group("/api")
	{
		api.GET("/battlefields", h.GetBattlefields)
		api.GET("/battlefields/:id", h.GetBattlefieldByID)
		api.GET("/roads", h.GetRoads)
		api.GET("/rivers", h.GetRivers)
		api.GET("/dem", h.GetDEM)
		api.GET("/terrain_profile", h.GetTerrainProfile)
		api.GET("/accessibility/:id", h.GetAccessibility)
		api.GET("/site_selection_factors", h.GetSiteSelectionFactors)
		api.GET("/enhanced_lr", h.GetEnhancedLR)
		api.GET("/high_prob_areas", h.GetHighProbAreas)
		api.GET("/military_regions", h.GetMilitaryRegions)
		api.GET("/fuzzy_cluster", h.GetFuzzyCluster)
		api.GET("/statistics", h.GetStatistics)
	}

	log.Printf("服务启动于 :%s", *port)
	if err := r.Run(":" + *port); err != nil {
		log.Fatal(err)
	}
}
