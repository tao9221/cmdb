package handlers

import (
	"cmdb-backend/database"
	"cmdb-backend/models"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

func getConfig(key string) string {
	var cfg models.SystemConfig
	if err := database.DB.Where("key = ?", key).First(&cfg).Error; err != nil {
		return ""
	}
	return cfg.Value
}

func sendMail(to, subject, body string) error {
	host := getConfig("smtp_host")
	port := getConfig("smtp_port")
	user := getConfig("smtp_user")
	pass := getConfig("smtp_pass")
	from := getConfig("smtp_from")
	
	if host == "" {
		return fmt.Errorf("SMTP 服务器未配置")
	}
	if user == "" {
		return fmt.Errorf("发件人账号未配置")
	}
	if pass == "" {
		return fmt.Errorf("发件人密码未配置")
	}
	if port == "" { port = "465" }
	if from == "" { from = user }

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", from, to, subject, body)
	addr := host + ":" + port

	tlsCfg := &tls.Config{ServerName: host, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		// 降级到普通 SMTP
		return smtp.SendMail(addr, smtp.PlainAuth("", user, pass, host), from, []string{to}, []byte(msg))
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, host)
	if err != nil { return err }
	if err = client.Auth(smtp.PlainAuth("", user, pass, host)); err != nil { return err }
	if err = client.Mail(from); err != nil { return err }
	if err = client.Rcpt(to); err != nil { return err }
	w, err := client.Data()
	if err != nil { return err }
	_, err = w.Write([]byte(msg))
	if err != nil { return err }
	return w.Close()
}

// 获取所有管理员邮箱（去重）
func getAdminEmails() []string {
	var admins []models.User
	database.DB.Where("role = ? AND email != ''", "admin").Find(&admins)
	seen := map[string]bool{}
	emails := make([]string, 0, len(admins))
	for _, a := range admins {
		if a.Email != "" && !seen[a.Email] {
			seen[a.Email] = true
			emails = append(emails, a.Email)
		}
	}
	return emails
}

// StartAlertScheduler 启动报警定时任务
func StartAlertScheduler() {
	go func() {
		for {
			runAlertCheck()
			time.Sleep(1 * time.Minute)
		}
	}()
	log.Println("报警定时任务已启动")
}

func getAlertSentTime(key string) time.Time {
	val := getConfig(key)
	if val == "" { return time.Time{} }
	t, err := time.Parse(time.RFC3339, val)
	if err != nil { return time.Time{} }
	return t
}

func setAlertSentTime(key string, t time.Time) {
	val := t.Format(time.RFC3339)
	var cfg models.SystemConfig
	if err := database.DB.Where("key = ?", key).First(&cfg).Error; err != nil {
		database.DB.Create(&models.SystemConfig{Key: key, Value: val})
	} else {
		cfg.Value = val
		database.DB.Save(&cfg)
	}
}

func runAlertCheck() {
	cycleStr := getConfig("alert_cycle_minutes")
	if cycleStr == "" { return }
	cycle, err := strconv.Atoi(cycleStr)
	if err != nil || cycle <= 0 { return }

	// 读取发送间隔，默认 60 分钟
	intervalStr := getConfig("alert_interval_minutes")
	interval := 60
	if n, err := strconv.Atoi(intervalStr); err == nil && n > 0 {
		interval = n
	}

	emails := getAdminEmails()
	if len(emails) == 0 { return }

	now := time.Now()
	threshold := now.Add(-time.Duration(cycle) * time.Minute)
	intervalDur := time.Duration(interval) * time.Minute

	// 检测离线机器
	var offlineServers []models.Server
	database.DB.Where("manual = false AND last_report < ? AND last_report > '0001-01-01'", threshold).Find(&offlineServers)
	if len(offlineServers) > 0 {
		// 超时未上报的机器状态更新为 offline
		for _, s := range offlineServers {
			if s.Status != "offline" {
				database.DB.Model(&s).Update("status", "offline")
				log.Printf("服务器 %s (%s) 超时未上报，状态更新为 offline", s.Hostname, s.IP)
			}
		}
		var toAlert []models.Server
		for _, s := range offlineServers {
			// 用数据库持久化上次发送时间，key: alert_sent_offline_{id}
			key := fmt.Sprintf("alert_sent_offline_%d", s.ID)
			lastSent := getAlertSentTime(key)
			if now.Sub(lastSent) >= intervalDur {
				toAlert = append(toAlert, s)
			}
		}
		if len(toAlert) > 0 {
			lines := make([]string, 0, len(toAlert))
			for _, s := range toAlert {
				lines = append(lines, fmt.Sprintf("  - %s (%s) 最后上报: %s", s.Hostname, s.IP, s.LastReport.Format("2006-01-02 15:04:05")))
				setAlertSentTime(fmt.Sprintf("alert_sent_offline_%d", s.ID), now)
			}
			body := fmt.Sprintf("以下服务器超过 %d 分钟未上报，请检查：\n\n%s", cycle, strings.Join(lines, "\n"))
			for _, email := range emails {
				if err := sendMail(email, "[CMDB告警] 服务器离线告警", body); err != nil {
					log.Printf("发送离线告警邮件失败 %s: %v", email, err)
				} else {
					log.Printf("离线告警邮件已发送至 %s，共 %d 台", email, len(toAlert))
				}
			}
		}
	}

	// 检测维保到期（30天内）
	warrantyThreshold := now.Add(30 * 24 * time.Hour)
	var warrantyServers []models.Server
	database.DB.Where("warranty_end IS NOT NULL AND warranty_end > ? AND warranty_end < ?", now, warrantyThreshold).Find(&warrantyServers)
	if len(warrantyServers) > 0 {
		var toAlert []models.Server
		for _, s := range warrantyServers {
			key := fmt.Sprintf("alert_sent_warranty_%d", s.ID)
			lastSent := getAlertSentTime(key)
			if now.Sub(lastSent) >= 24*time.Hour {
				toAlert = append(toAlert, s)
				setAlertSentTime(key, now)
			}
		}
		if len(toAlert) > 0 {
			lines := make([]string, 0, len(toAlert))
			for _, s := range toAlert {
				days := int(s.WarrantyEnd.Sub(now).Hours() / 24)
				lines = append(lines, fmt.Sprintf("  - %s (%s) 维保到期: %s（剩余 %d 天）", s.Hostname, s.IP, s.WarrantyEnd.Format("2006-01-02"), days))
			}
			body := fmt.Sprintf("以下服务器维保即将到期（30天内），请及时续保：\n\n%s", strings.Join(lines, "\n"))
			for _, email := range emails {
				if err := sendMail(email, "[CMDB告警] 维保到期告警", body); err != nil {
					log.Printf("发送维保告警邮件失败 %s: %v", email, err)
				} else {
					log.Printf("维保告警邮件已发送至 %s，共 %d 台", email, len(toAlert))
				}
			}
		}
	}
}
