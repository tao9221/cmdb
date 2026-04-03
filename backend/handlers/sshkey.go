package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetSSHKey(c *gin.Context) {
	var cfg models.SSHKeyConfig
	database.DB.First(&cfg)
	// 普通用户不返回私钥内容
	role, _ := c.Get("role")
	if role != "admin" {
		cfg.PrivateKey = ""
	}
	c.JSON(http.StatusOK, cfg)
}

func SaveSSHKey(c *gin.Context) {
	if !adminCheck(c) { return }
	var req models.SSHKeyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var cfg models.SSHKeyConfig
	database.DB.First(&cfg)
	if cfg.ID == 0 {
		database.DB.Create(&req)
		c.JSON(http.StatusOK, req)
	} else {
		cfg.Username = req.Username
		cfg.Port = req.Port
		if req.PrivateKey != "" {
			cfg.PrivateKey = req.PrivateKey
		}
		database.DB.Save(&cfg)
		c.JSON(http.StatusOK, cfg)
	}
}
