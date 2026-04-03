package main

// @title           CMDB 智能运维平台 API
// @version         1.0
// @description     CMDB 后端接口文档，包含服务器管理、机房管理、用户管理、告警配置等接口。
// @host            localhost:8088
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 格式: Bearer {token}

import (
	"cmdb-backend/database"
	"cmdb-backend/handlers"
	"cmdb-backend/middleware"
	_ "cmdb-backend/docs"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	database.Init()

	// 启动报警定时任务
	handlers.StartAlertScheduler()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 公开接口
	r.POST("/api/login", handlers.Login)

	// 服务前端静态文件（生产环境，dist目录放在backend同级）
	r.Static("/assets", "./dist/assets")
	r.StaticFile("/favicon.ico", "./dist/favicon.ico")
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.File("./dist/index.html")
	})
	r.POST("/api/agent/report", handlers.AgentReport)
	r.GET("/api/ssh/terminal", handlers.SSHTerminal)
	r.GET("/api/rdp/connect", handlers.GenerateRDP)

	// 需要认证的接口
	api := r.Group("/api", middleware.AuthMiddleware())
	{
		api.GET("/overview", handlers.Overview)

		// 机房
		api.GET("/datacenters", handlers.ListDataCenters)
		api.GET("/datacenters/:id", handlers.GetDataCenter)
		api.POST("/datacenters", handlers.CreateDataCenter)
		api.PUT("/datacenters/:id", handlers.UpdateDataCenter)
		api.DELETE("/datacenters/:id", handlers.DeleteDataCenter)

		// 机柜
		api.GET("/datacenters/:id/cabinets", handlers.ListCabinets)
		api.POST("/cabinets", handlers.CreateCabinet)
		api.DELETE("/cabinets/:id", handlers.DeleteCabinet)
		api.GET("/cabinets/:id/servers", handlers.GetCabinetServers)
		api.PUT("/cabinets/positions", handlers.UpdateCabinetPositions)

		// 服务器
		api.GET("/servers", handlers.ListServers)
		api.POST("/servers", handlers.CreateServer)
		api.GET("/servers/:id", handlers.GetServer)
		api.PUT("/servers/:id", handlers.UpdateServer)
		api.DELETE("/servers/:id", handlers.DeleteServer)
		api.DELETE("/servers", handlers.BatchDeleteServers)

		// 当前用户信息
		api.GET("/me", handlers.GetMe)

		// SSH 密钥配置
		api.GET("/sshkey", handlers.GetSSHKey)
		api.POST("/sshkey", handlers.SaveSSHKey)

		// 资源统计
		api.GET("/stats", handlers.GetStats)

		// 用户管理 + 系统设置（仅管理员）
		admin := api.Group("/admin", handlers.AdminOnly)
		{
			admin.GET("/users", handlers.ListUsers)
			admin.POST("/users", handlers.CreateUser)
			admin.PUT("/users/:id", handlers.UpdateUser)
			admin.DELETE("/users/:id", handlers.DeleteUser)
			admin.GET("/users/:id/access", handlers.GetUserAccess)
			admin.PUT("/users/:id/access", handlers.SetUserAccess)
			admin.GET("/settings", handlers.GetSettings)
			admin.POST("/settings", handlers.SaveSettings)
			admin.POST("/settings/test-mail", handlers.TestMail)
		}
	}

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("CMDB后端启动在 :8088")
	r.Run(":8088")
}
