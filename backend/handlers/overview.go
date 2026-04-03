package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Overview(c *gin.Context) {
	var totalServers, onlineServers, offlineServers int64
	var totalDC, totalCabinets int64

	database.DB.Model(&models.Server{}).Count(&totalServers)
	database.DB.Model(&models.Server{}).Where("status = ?", "online").Count(&onlineServers)
	database.DB.Model(&models.Server{}).Where("status = ?", "offline").Count(&offlineServers)
	database.DB.Model(&models.DataCenter{}).Count(&totalDC)
	database.DB.Model(&models.Cabinet{}).Count(&totalCabinets)

	// CPU/内存平均使用率
	var avgCPU, avgMem float64
	row := database.DB.Model(&models.Server{}).Select("AVG(cpu_usage)").Row()
	row.Scan(&avgCPU)
	row2 := database.DB.Model(&models.Server{}).Select("AVG(CAST(mem_used AS FLOAT)/NULLIF(mem_total,0)*100)").Row()
	row2.Scan(&avgMem)

	// 最近上报的服务器，返回50条供前端翻页
	var recentServers []models.Server
	database.DB.Order("last_report desc").Limit(50).Find(&recentServers)

	c.JSON(http.StatusOK, gin.H{
		"total_servers":   totalServers,
		"online_servers":  onlineServers,
		"offline_servers": offlineServers,
		"total_dc":        totalDC,
		"total_cabinets":  totalCabinets,
		"avg_cpu":         avgCPU,
		"avg_mem":         avgMem,
		"recent_servers":  recentServers,
	})
}
