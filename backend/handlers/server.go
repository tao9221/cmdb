package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ListServers(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")

	query := database.DB.Preload("Cabinet")
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("hostname LIKE ? OR ip LIKE ? OR vendor LIKE ? OR model LIKE ?", like, like, like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	// cabinet_id=unassigned 查未分配服务器
	cabinetID := c.Query("cabinet_id")
	if cabinetID == "unassigned" {
		query = query.Where("cabinet_id IS NULL")
	} else if cabinetID != "" {
		query = query.Where("cabinet_id = ?", cabinetID)
	}

	// 普通用户只能看到授权的主机
	role, _ := c.Get("role")
	if role != "admin" {
		uid, _ := c.Get("user_id")
		var accesses []models.UserServerAccess
		database.DB.Where("user_id = ?", uid).Find(&accesses)
		if len(accesses) > 0 {
			ids := make([]uint, 0, len(accesses))
			for _, a := range accesses {
				ids = append(ids, a.ServerID)
			}
			query = query.Where("id IN ?", ids)
		} else {
			c.JSON(http.StatusOK, []models.Server{})
			return
		}
	}

	var servers []models.Server
	query.Find(&servers)
	c.JSON(http.StatusOK, servers)
}

func GetServer(c *gin.Context) {
	id := c.Param("id")
	var server models.Server
	if err := database.DB.Preload("Cabinet").First(&server, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "服务器不存在"})
		return
	}
	c.JSON(http.StatusOK, server)
}

// 手动创建服务器（仅管理员）
func CreateServer(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		uid, _ := c.Get("user_id")
		var user models.User
		if err := database.DB.First(&user, uid).Error; err != nil || user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			return
		}
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip, _ := body["ip"].(string)
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP 不能为空"})
		return
	}

	server := models.Server{Manual: true, Status: "online"}
	server.IP = ip
	if v, ok := body["hostname"].(string); ok { server.Hostname = v }
	if v, ok := body["vendor"].(string); ok { server.Vendor = v }
	if v, ok := body["model"].(string); ok { server.ServerModel = v }
	if v, ok := body["os"].(string); ok { server.OS = v }
	if v, ok := body["cpu_model"].(string); ok { server.CPUModel = v }
	if v, ok := body["remark"].(string); ok { server.Remark = v }
	if v, ok := body["cpu_cores"].(float64); ok { server.CPUCores = int(v) }
	if v, ok := body["mem_total"].(float64); ok { server.MemTotal = int64(v) }
	if v, ok := body["disk_total"].(float64); ok { server.DiskTotal = int64(v) }
	if v, ok := body["slot"].(float64); ok { server.Slot = int(v) }
	if v, ok := body["cabinet_id"]; ok && v != nil {
		if f, ok := v.(float64); ok { uid := uint(f); server.CabinetID = &uid }
	}
	if v, ok := body["warranty_end"]; ok && v != nil {
		if s, ok := v.(string); ok && s != "" {
			if t, err := time.Parse("2006-01-02", s); err == nil {
				server.WarrantyEnd = &t
			}
		}
	}

	// 校验机位冲突
	if server.CabinetID != nil && server.Slot > 0 {
		var conflict models.Server
		if err := database.DB.Where("cabinet_id = ? AND slot = ?", *server.CabinetID, server.Slot).First(&conflict).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该机位已被占用，请选择其他机位"})
			return
		}
	}

	if err := database.DB.Create(&server).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP 已存在"})
		return
	}
	c.JSON(http.StatusOK, server)
}

func UpdateServer(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		uid, _ := c.Get("user_id")
		var user models.User
		if err := database.DB.First(&user, uid).Error; err != nil || user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			return
		}
	}
	id := c.Param("id")
	var server models.Server
	if err := database.DB.First(&server, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "服务器不存在"})
		return
	}

	// 用 map 接收，单独处理 warranty_end 日期字符串
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 逐字段更新，避免 ShouldBindJSON 无法解析 *time.Time
	if v, ok := body["hostname"]; ok { server.Hostname = v.(string) }
	if v, ok := body["ip"]; ok { server.IP = v.(string) }
	if v, ok := body["vendor"]; ok { server.Vendor = v.(string) }
	if v, ok := body["model"]; ok { server.ServerModel = v.(string) }
	if v, ok := body["os"]; ok { server.OS = v.(string) }
	if v, ok := body["status"]; ok { server.Status = v.(string) }
	if v, ok := body["cpu_model"]; ok { server.CPUModel = v.(string) }
	if v, ok := body["remark"]; ok { server.Remark = v.(string) }
	if v, ok := body["cpu_cores"]; ok {
		if f, ok := v.(float64); ok { server.CPUCores = int(f) }
	}
	if v, ok := body["slot"]; ok {
		if f, ok := v.(float64); ok { server.Slot = int(f) }
	}
	if v, ok := body["cabinet_id"]; ok {
		if v == nil {
			server.CabinetID = nil
		} else if f, ok := v.(float64); ok {
			uid := uint(f)
			server.CabinetID = &uid
		}
	}

	// 校验机位冲突（排除自身）
	if server.CabinetID != nil && server.Slot > 0 {
		var conflict models.Server
		if err := database.DB.Where("cabinet_id = ? AND slot = ? AND id != ?", *server.CabinetID, server.Slot, server.ID).First(&conflict).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该机位已被占用，请选择其他机位"})
			return
		}
	}
	// warranty_end: "2026-04-30" 或 null
	if v, ok := body["warranty_end"]; ok {
		if v == nil {
			server.WarrantyEnd = nil
		} else if s, ok := v.(string); ok && s != "" {
			t, err := time.Parse("2006-01-02", s)
			if err == nil {
				server.WarrantyEnd = &t
			}
		}
	}

	database.DB.Save(&server)
	c.JSON(http.StatusOK, server)
}

func DeleteServer(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}
	id := c.Param("id")
	if err := database.DB.Delete(&models.Server{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func BatchDeleteServers(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}
	var body struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids 不能为空"})
		return
	}
	database.DB.Delete(&models.Server{}, body.IDs)
	c.JSON(http.StatusOK, gin.H{"message": "批量删除成功", "count": len(body.IDs)})
}

// Agent上报接口 - 手动添加的机器不被覆盖
func AgentReport(c *gin.Context) {
	var report models.AgentReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var server models.Server
	result := database.DB.Where("ip = ?", report.IP).First(&server)
	if result.Error != nil {
		// 新服务器，自动创建
		server = models.Server{
			Hostname:     report.Hostname,
			IP:           report.IP,
			Vendor:       report.Vendor,
			ServerModel:  report.ServerModel,
			OS:           report.OS,
			CPUModel:     report.CPUModel,
			CPUCores:     report.CPUCores,
			CPUUsage:     report.CPUUsage,
			MemTotal:     report.MemTotal,
			MemUsed:      report.MemUsed,
			DiskTotal:    report.DiskTotal,
			DiskUsed:     report.DiskUsed,
			AgentVersion: report.AgentVersion,
			Status:       "online",
			LastReport:   time.Now(),
		}
		database.DB.Create(&server)
	} else {
		// 手动添加的机器：跳过信息覆盖，只更新在线状态
		if server.Manual {
			database.DB.Model(&server).Updates(map[string]interface{}{
				"status":      "online",
				"last_report": time.Now(),
			})
			c.JSON(http.StatusOK, gin.H{"message": "手动机器，跳过覆盖", "server_id": server.ID})
			return
		}
		// 正常 agent 上报更新 - 只更新变化的字段，避免全量 UPDATE
		database.DB.Model(&server).Updates(map[string]interface{}{
			"hostname":      report.Hostname,
			"vendor":        report.Vendor,
			"model":         report.ServerModel,
			"os":            report.OS,
			"cpu_model":     report.CPUModel,
			"cpu_cores":     report.CPUCores,
			"cpu_usage":     report.CPUUsage,
			"mem_total":     report.MemTotal,
			"mem_used":      report.MemUsed,
			"disk_total":    report.DiskTotal,
			"disk_used":     report.DiskUsed,
			"net_in":        report.NetIn,
			"net_out":       report.NetOut,
			"agent_version": report.AgentVersion,
			"status":        "online",
			"last_report":   time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "上报成功", "server_id": server.ID})
}
