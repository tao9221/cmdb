package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 仅管理员可调用的中间件
func AdminOnly(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		c.Abort()
		return
	}
	c.Next()
}

func ListUsers(c *gin.Context) {
	var users []models.User
	database.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Email    string `json:"email"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}
	if req.Role == "" { req.Role = "user" }
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := models.User{Username: req.Username, Password: string(hash), Role: req.Role, Email: req.Email, Remark: req.Remark}
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	var req struct {
		Password string `json:"password"`
		Role     string `json:"role"`
		Email    string `json:"email"`
		Remark   string `json:"remark"`
	}
	c.ShouldBindJSON(&req)
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user.Password = string(hash)
	}
	if req.Role != "" { user.Role = req.Role }
	user.Email = req.Email
	user.Remark = req.Remark
	database.DB.Save(&user)
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	// 不能删除自己
	selfID, _ := c.Get("user_id")
	if selfID.(float64) == float64(parseUint(id)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己"})
		return
	}
	database.DB.Delete(&models.User{}, id)
	database.DB.Where("user_id = ?", id).Delete(&models.UserServerAccess{})
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 获取用户的主机授权列表
func GetUserAccess(c *gin.Context) {
	id := c.Param("id")
	var accesses []models.UserServerAccess
	database.DB.Where("user_id = ?", id).Find(&accesses)
	serverIDs := make([]uint, 0, len(accesses))
	for _, a := range accesses {
		serverIDs = append(serverIDs, a.ServerID)
	}
	c.JSON(http.StatusOK, gin.H{"server_ids": serverIDs})
}

// 设置用户的主机授权（全量替换）
func SetUserAccess(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ServerIDs []uint `json:"server_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 全量替换
	database.DB.Where("user_id = ?", id).Delete(&models.UserServerAccess{})
	for _, sid := range req.ServerIDs {
		database.DB.Create(&models.UserServerAccess{UserID: parseUint(id), ServerID: sid})
	}
	c.JSON(http.StatusOK, gin.H{"message": "授权更新成功"})
}

// 获取当前登录用户信息
func GetMe(c *gin.Context) {
	uid, _ := c.Get("user_id")
	var user models.User
	database.DB.First(&user, uid)
	c.JSON(http.StatusOK, user)
}

func parseUint(s string) uint {
	var n uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + uint(c-'0')
		}
	}
	return n
}
