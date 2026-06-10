package main

import (
	"log"
	"net/http"

	"ancient-battlefield/pkg/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(CORS())

	api := r.Group("/api")
	{
		api.GET("/battlefields", handlers.GetBattlefields)
		api.GET("/battlefields/:id", handlers.GetBattlefieldByID)
		api.GET("/roads", handlers.GetRoads)
		api.GET("/rivers", handlers.GetRivers)
		api.GET("/dem", handlers.GetDEM)
		api.GET("/terrain_profile", handlers.GetTerrainProfile)
		api.GET("/accessibility/:id", handlers.GetAccessibility)
		api.GET("/site_selection_factors", handlers.GetSiteSelectionFactors)
		api.GET("/enhanced_lr", handlers.GetEnhancedLR)
		api.GET("/high_prob_areas", handlers.GetHighProbAreas)
		api.GET("/military_regions", handlers.GetMilitaryRegions)
		api.GET("/fuzzy_cluster", handlers.GetFuzzyCluster)
		api.GET("/stats", handlers.GetStats)
	}

	r.Static("/web", "./web")
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/index.html", "./web/index.html")

	port := "8080"
	log.Printf("古代战场遗址空间分布与军事地理分析系统启动中...")
	log.Printf("服务地址: http://localhost:%s", port)
	log.Printf("前端页面: http://localhost:%s/index.html", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
