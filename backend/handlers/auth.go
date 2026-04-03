package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/middleware"
	"cmdb-backend/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/go-ldap/ldap/v3"
)

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	user, err := tryLocalLogin(req.Username, req.Password)
	if err != nil {
		// 本地登录失败，尝试 LDAP
		user, err = tryLDAPLogin(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString(middleware.JWTSecret)
	c.JSON(http.StatusOK, gin.H{"token": tokenStr, "user": user})
}

func tryLocalLogin(username, password string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("wrong password")
	}
	return &user, nil
}

func tryLDAPLogin(username, password string) (*models.User, error) {
	// 检查 LDAP 是否启用
	if getConfig("ldap_enabled") != "true" {
		return nil, fmt.Errorf("ldap disabled")
	}
	host := getConfig("ldap_host")
	baseDN := getConfig("ldap_base_dn")
	bindDN := getConfig("ldap_bind_dn")
	bindPass := getConfig("ldap_bind_pass")
	filter := getConfig("ldap_filter")
	if host == "" || baseDN == "" {
		return nil, fmt.Errorf("ldap not configured")
	}
	if filter == "" {
		filter = "(objectClass=person)"
	}

	// 连接 LDAP
	l, err := ldap.DialURL(host)
	if err != nil {
		return nil, fmt.Errorf("ldap connect: %v", err)
	}
	defer l.Close()

	// 管理员 bind 搜索用户 DN
	if err = l.Bind(bindDN, bindPass); err != nil {
		return nil, fmt.Errorf("ldap bind: %v", err)
	}
	emailAttrName := getConfig("ldap_email_attr")
	if emailAttrName == "" { emailAttrName = "mail" }
	searchAttrs := []string{"dn", "cn", "mail", "email", "userPrincipalName", "uid", "sAMAccountName", emailAttrName}

	searchReq := ldap.NewSearchRequest(
		baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&%s(|(uid=%s)(sAMAccountName=%s)(cn=%s)))", filter, username, username, username),
		searchAttrs,
		nil,
	)
	sr, err := l.Search(searchReq)
	if err != nil || len(sr.Entries) == 0 {
		return nil, fmt.Errorf("ldap user not found")
	}
	userDN := sr.Entries[0].DN
	// 依次尝试常见邮箱属性
	email := sr.Entries[0].GetAttributeValue(emailAttrName)
	if email == "" { email = sr.Entries[0].GetAttributeValue("mail") }
	if email == "" { email = sr.Entries[0].GetAttributeValue("email") }
	if email == "" { email = sr.Entries[0].GetAttributeValue("userPrincipalName") }
	log.Printf("LDAP user %s DN=%s email=%s", username, userDN, email)

	// 用用户自己的凭据 bind 验证密码
	if err = l.Bind(userDN, password); err != nil {
		return nil, fmt.Errorf("ldap auth failed")
	}

	// LDAP 认证成功，查找或创建本地用户记录
	var localUser models.User
	if err = database.DB.Where("username = ?", username).First(&localUser).Error; err != nil {
		// 首次登录，创建本地记录，角色默认 user，密码设为不可用（空 hash）
		localUser = models.User{
			Username: username,
			Password: "ldap_user_no_local_password",
			Role:     "user",
			Email:    email,
			Remark:   "LDAP用户",
		}
		database.DB.Create(&localUser)
	} else {
		// 已存在，更新邮箱（角色由管理员手动管理，不覆盖）
		if email != "" && localUser.Email != email {
			localUser.Email = email
			database.DB.Save(&localUser)
		}
	}
	return &localUser, nil
}
