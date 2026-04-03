package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStats 资源统计 Top20
// @Summary      资源统计
// @Description  返回 CPU/内存/磁盘/网络流量 Top 20 在线服务器
// @Tags         统计
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  object
// @Router       /stats [get]
func GetStats(c *gin.Context) {
	var servers []models.Server
	database.DB.Where("status = ?", "online").Find(&servers)

	type ServerStat struct {
		ID        uint    `json:"id"`
		Hostname  string  `json:"hostname"`
		IP        string  `json:"ip"`
		Value     float64 `json:"value"`
		ValueStr  string  `json:"value_str"`
	}

	fmtBytes := func(b int64) string {
		if b >= 1073741824 {
			return fmt.Sprintf("%.1f GB", float64(b)/1073741824)
		}
		if b >= 1048576 {
			return fmt.Sprintf("%.1f MB", float64(b)/1048576)
		}
		if b >= 1024 {
			return fmt.Sprintf("%.1f KB", float64(b)/1024)
		}
		return fmt.Sprintf("%d B", b)
	}

	top := func(n int, less func(a, b models.Server) bool, val func(s models.Server) (float64, string)) []ServerStat {
		sorted := make([]models.Server, len(servers))
		copy(sorted, servers)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if less(sorted[j], sorted[i]) {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		if n > len(sorted) { n = len(sorted) }
		result := make([]ServerStat, n)
		for i := 0; i < n; i++ {
			v, vs := val(sorted[i])
			result[i] = ServerStat{ID: sorted[i].ID, Hostname: sorted[i].Hostname, IP: sorted[i].IP, Value: v, ValueStr: vs}
		}
		return result
	}

	cpuTop := top(20,
		func(a, b models.Server) bool { return a.CPUUsage > b.CPUUsage },
		func(s models.Server) (float64, string) { return s.CPUUsage, fmt.Sprintf("%.1f%%", s.CPUUsage) },
	)

	memTop := top(20,
		func(a, b models.Server) bool {
			ap := float64(0); if a.MemTotal > 0 { ap = float64(a.MemUsed) / float64(a.MemTotal) * 100 }
			bp := float64(0); if b.MemTotal > 0 { bp = float64(b.MemUsed) / float64(b.MemTotal) * 100 }
			return ap > bp
		},
		func(s models.Server) (float64, string) {
			if s.MemTotal == 0 { return 0, "-" }
			pct := float64(s.MemUsed) / float64(s.MemTotal) * 100
			return pct, fmt.Sprintf("%.1f%% (%s)", pct, fmtBytes(s.MemUsed))
		},
	)

	diskTop := top(20,
		func(a, b models.Server) bool {
			ap := float64(0); if a.DiskTotal > 0 { ap = float64(a.DiskUsed) / float64(a.DiskTotal) * 100 }
			bp := float64(0); if b.DiskTotal > 0 { bp = float64(b.DiskUsed) / float64(b.DiskTotal) * 100 }
			return ap > bp
		},
		func(s models.Server) (float64, string) {
			if s.DiskTotal == 0 { return 0, "-" }
			pct := float64(s.DiskUsed) / float64(s.DiskTotal) * 100
			return pct, fmt.Sprintf("%.1f%% (%s)", pct, fmtBytes(s.DiskUsed))
		},
	)

	netTop := top(20,
		func(a, b models.Server) bool { return (a.NetIn + a.NetOut) > (b.NetIn + b.NetOut) },
		func(s models.Server) (float64, string) {
			total := s.NetIn + s.NetOut
			return float64(total), fmt.Sprintf("↓%s/s ↑%s/s", fmtBytes(s.NetIn), fmtBytes(s.NetOut))
		},
	)

	c.JSON(http.StatusOK, gin.H{
		"cpu":  cpuTop,
		"mem":  memTop,
		"disk": diskTop,
		"net":  netTop,
	})
}
