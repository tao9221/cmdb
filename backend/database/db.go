package database

import (
	"cmdb-backend/models"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	// 确保数据目录存在
	os.MkdirAll("data", 0755)
	var err error
	DB, err = gorm.Open(sqlite.Open("data/cmdb.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	DB.AutoMigrate(&models.User{}, &models.DataCenter{}, &models.Cabinet{}, &models.Server{}, &models.UserServerAccess{}, &models.SSHKeyConfig{}, &models.SystemConfig{})

	// 创建默认管理员
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		DB.Create(&models.User{Username: "admin", Password: string(hash), Role: "admin"})

		// 创建示例数据
		dc := models.DataCenter{Name: "北京数据中心", Location: "北京市朝阳区", Desc: "主数据中心"}
		DB.Create(&dc)
		dc2 := models.DataCenter{Name: "上海数据中心", Location: "上海市浦东新区", Desc: "灾备数据中心"}
		DB.Create(&dc2)

		for i := 1; i <= 4; i++ {
			cab := models.Cabinet{Name: "A" + string(rune('0'+i)), DataCenterID: dc.ID, Row: "A", Col: string(rune('0' + i))}
			DB.Create(&cab)
		}
		for i := 1; i <= 2; i++ {
			cab := models.Cabinet{Name: "B" + string(rune('0'+i)), DataCenterID: dc2.ID, Row: "B", Col: string(rune('0' + i))}
			DB.Create(&cab)
		}
	}
}
