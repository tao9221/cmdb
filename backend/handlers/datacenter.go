package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func adminCheck(c *gin.Context) bool {
	role, _ := c.Get("role")
	// token 里有 role 直接判断
	if role == "admin" {
		return true
	}
	// fallback：从数据库查用户角色（兼容旧 token）
	uid, exists := c.Get("user_id")
	if exists {
		var user models.User
		if err := database.DB.First(&user, uid).Error; err == nil && user.Role == "admin" {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
	return false
}

func ListDataCenters(c *gin.Context) {
	var dcs []models.DataCenter
	database.DB.Preload("Cabinets").Find(&dcs)
	c.JSON(http.StatusOK, dcs)
}

func GetDataCenter(c *gin.Context) {
	id := c.Param("id")
	var dc models.DataCenter
	if err := database.DB.Preload("Cabinets").First(&dc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "机房不存在"})
		return
	}
	c.JSON(http.StatusOK, dc)
}

func CreateDataCenter(c *gin.Context) {
	if !adminCheck(c) { return }
	var dc models.DataCenter
	if err := c.ShouldBindJSON(&dc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&dc)
	c.JSON(http.StatusOK, dc)
}

func UpdateDataCenter(c *gin.Context) {
	if !adminCheck(c) { return }
	id := c.Param("id")
	var dc models.DataCenter
	if err := database.DB.First(&dc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "机房不存在"})
		return
	}
	c.ShouldBindJSON(&dc)
	database.DB.Save(&dc)
	c.JSON(http.StatusOK, dc)
}

func DeleteDataCenter(c *gin.Context) {
	if !adminCheck(c) { return }
	id := c.Param("id")
	database.DB.Delete(&models.DataCenter{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func ListCabinets(c *gin.Context) {
	dcID := c.Param("id")
	var cabinets []models.Cabinet
	database.DB.Where("data_center_id = ?", dcID).Find(&cabinets)
	type CabinetWithCount struct {
		models.Cabinet
		ServerCount int64 `json:"server_count"`
	}
	var result []CabinetWithCount
	for _, cab := range cabinets {
		var count int64
		database.DB.Model(&models.Server{}).Where("cabinet_id = ?", cab.ID).Count(&count)
		result = append(result, CabinetWithCount{Cabinet: cab, ServerCount: count})
	}
	c.JSON(http.StatusOK, result)
}

func GetCabinetServers(c *gin.Context) {
	cabID := c.Param("id")
	var servers []models.Server
	database.DB.Where("cabinet_id = ?", cabID).Find(&servers)
	c.JSON(http.StatusOK, servers)
}

func CreateCabinet(c *gin.Context) {
	if !adminCheck(c) { return }
	var cab models.Cabinet
	if err := c.ShouldBindJSON(&cab); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&cab)
	c.JSON(http.StatusOK, cab)
}

func DeleteCabinet(c *gin.Context) {
	if !adminCheck(c) { return }
	id := c.Param("id")
	database.DB.Delete(&models.Cabinet{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 批量更新机柜位置
func UpdateCabinetPositions(c *gin.Context) {
	if !adminCheck(c) { return }
	var positions []struct {
		ID   uint `json:"id"`
		PosX int  `json:"pos_x"`
		PosY int  `json:"pos_y"`
	}
	if err := c.ShouldBindJSON(&positions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, p := range positions {
		database.DB.Model(&models.Cabinet{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
			"pos_x": p.PosX, "pos_y": p.PosY,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "位置已保存"})
}
