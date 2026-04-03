package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GenerateRDP 生成 .rdp 文件供浏览器下载，用户双击即可用 mstsc 连接
func GenerateRDP(c *gin.Context) {
	ip := c.Query("ip")
	username := c.Query("username")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip 不能为空"})
		return
	}

	rdpContent := fmt.Sprintf(`full address:s:%s
username:s:%s
prompt for credentials:i:1
authentication level:i:2
redirectclipboard:i:1
redirectprinters:i:0
redirectsmartcards:i:0
redirectcomports:i:0
redirectdrives:i:0
screen mode id:i:2
use multimon:i:0
desktopwidth:i:1920
desktopheight:i:1080
session bpp:i:32
compression:i:1
keyboardhook:i:2
audiocapturemode:i:0
videoplaybackmode:i:1
connection type:i:7
networkautodetect:i:1
bandwidthautodetect:i:1
displayconnectionbar:i:1
enableworkspacereconnect:i:0
disable wallpaper:i:1
allow font smoothing:i:1
allow desktop composition:i:1
disable full window drag:i:0
disable menu anims:i:0
disable themes:i:0
disable cursor setting:i:0
bitmapcachepersistenable:i:1
audiomode:i:0
redirectposdevices:i:0
redirectdirectx:i:1
drivestoredirect:s:
`, ip, username)

	filename := fmt.Sprintf("connect-%s.rdp", ip)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/x-rdp")
	c.String(http.StatusOK, rdpContent)
}
