package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/models"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetSettings(c *gin.Context) {
	var configs []models.SystemConfig
	database.DB.Find(&configs)
	result := map[string]string{}
	for _, cfg := range configs {
		result[cfg.Key] = cfg.Value
	}
	c.JSON(http.StatusOK, result)
}

func SaveSettings(c *gin.Context) {
	if !adminCheck(c) { return }
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for k, v := range body {
		var cfg models.SystemConfig
		if err := database.DB.Where("key = ?", k).First(&cfg).Error; err != nil {
			database.DB.Create(&models.SystemConfig{Key: k, Value: v})
		} else {
			cfg.Value = v
			database.DB.Save(&cfg)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "保存成功"})
}

func TestMail(c *gin.Context) {
	if !adminCheck(c) { return }

	emails := getAdminEmails()
	if len(emails) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有配置管理员邮箱，请先在用户管理中为管理员账号设置邮箱"})
		return
	}

	var errs []string
	for _, email := range emails {
		if err := sendMail(email, "[CMDB] 测试邮件", "这是一封来自 CMDB 系统的测试邮件，收到说明邮件配置正确。"); err != nil {
			errs = append(errs, email+": "+err.Error())
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败: " + strings.Join(errs, "; ")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("测试邮件已发送至 %s", strings.Join(emails, ", "))})
}
