package models

import (
	"time"
)

// 覆盖 gorm.Model 让 ID 序列化为小写 id
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at,omitempty" gorm:"index"`
}

type User struct {
	BaseModel
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"-" gorm:"not null"`
	Role     string `json:"role" gorm:"default:'user'"`
	Email    string `json:"email"`
	Remark   string `json:"remark"`
}

// 系统设置（单行配置表）
type SystemConfig struct {
	ID    uint   `json:"id" gorm:"primarykey"`
	Key   string `json:"key" gorm:"unique;not null"`
	Value string `json:"value"`
}

// 用户可访问的主机授权（空表示全部可访问，仅admin有效）
// SSH 密钥配置（全局，仅管理员可读写）
type SSHKeyConfig struct {
	ID         uint   `json:"id" gorm:"primarykey"`
	Username   string `json:"username"`
	Port       string `json:"port"`
	PrivateKey string `json:"private_key"` // 存储时加密
}

type UserServerAccess struct {
	ID       uint `json:"id" gorm:"primarykey"`
	UserID   uint `json:"user_id" gorm:"index"`
	ServerID uint `json:"server_id"`
}

type DataCenter struct {
	BaseModel
	Name     string    `json:"name" gorm:"not null"`
	Location string    `json:"location"`
	Desc     string    `json:"desc"`
	Cabinets []Cabinet `json:"cabinets,omitempty" gorm:"foreignKey:DataCenterID"`
}

type Cabinet struct {
	BaseModel
	Name         string     `json:"name" gorm:"not null"`
	DataCenterID uint       `json:"data_center_id"`
	Row          string     `json:"row"`
	Col          string     `json:"col"`
	PosX         int        `json:"pos_x" gorm:"default:0"`
	PosY         int        `json:"pos_y" gorm:"default:0"`
	Servers      []Server   `json:"servers,omitempty" gorm:"foreignKey:CabinetID"`
}

type Server struct {
	BaseModel
	Hostname     string    `json:"hostname"`
	IP           string    `json:"ip" gorm:"unique"`
	CabinetID    *uint     `json:"cabinet_id"`
	Slot         int       `json:"slot" gorm:"default:0"` // 机柜位号，0表示未指定
	Vendor       string    `json:"vendor"`
	ServerModel  string    `json:"model" gorm:"column:model"`
	OS           string    `json:"os"`
	CPUModel     string    `json:"cpu_model"`
	CPUCores     int       `json:"cpu_cores"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemTotal     int64     `json:"mem_total"`
	MemUsed      int64     `json:"mem_used"`
	DiskTotal    int64     `json:"disk_total"`
	DiskUsed     int64     `json:"disk_used"`
	Status       string    `json:"status" gorm:"default:'online'"`
	AgentVersion string    `json:"agent_version"`
	LastReport   time.Time `json:"last_report"`
	Manual       bool      `json:"manual" gorm:"default:false"`
	Remark       string    `json:"remark"`
	WarrantyEnd  *time.Time `json:"warranty_end" gorm:"default:null"` // 维保到期日，手动填写
	NetIn        int64     `json:"net_in"`  // 入流量 bytes/s
	NetOut       int64     `json:"net_out"` // 出流量 bytes/s
	Cabinet      *Cabinet  `json:"cabinet,omitempty" gorm:"foreignKey:CabinetID"`
}

type AgentReport struct {
	Hostname     string  `json:"hostname"`
	IP           string  `json:"ip"`
	Vendor       string  `json:"vendor"`
	ServerModel  string  `json:"model"`
	OS           string  `json:"os"`
	CPUModel     string  `json:"cpu_model"`
	CPUCores     int     `json:"cpu_cores"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemTotal     int64   `json:"mem_total"`
	MemUsed      int64   `json:"mem_used"`
	DiskTotal    int64   `json:"disk_total"`
	DiskUsed     int64   `json:"disk_used"`
	NetIn        int64   `json:"net_in"`
	NetOut       int64   `json:"net_out"`
	AgentVersion string  `json:"agent_version"`
}
