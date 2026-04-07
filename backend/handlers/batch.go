package handlers

import (
	"bytes"
	"cmdb-backend/database"
	"cmdb-backend/models"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

type BatchExecRequest struct {
	ServerIDs []uint `json:"server_ids"`
	Command   string `json:"command"`
	Timeout   int    `json:"timeout"`
}

type BatchExecResult struct {
	ServerID uint   `json:"server_id"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// BatchExec 批量执行命令
func BatchExec(c *gin.Context) {
	var req BatchExecRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.ServerIDs) == 0 || req.Command == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：需要 server_ids 和 command"})
		return
	}
	if req.Timeout <= 0 || req.Timeout > 300 {
		req.Timeout = 30
	}
	var servers []models.Server
	database.DB.Where("id IN ?", req.ServerIDs).Find(&servers)
	if len(servers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到指定服务器"})
		return
	}
	var keyCfg models.SSHKeyConfig
	database.DB.First(&keyCfg)

	results := make([]BatchExecResult, len(servers))
	var wg sync.WaitGroup
	for i, srv := range servers {
		wg.Add(1)
		go func(idx int, s models.Server) {
			defer wg.Done()
			results[idx] = execOnServer(s, req.Command, keyCfg, req.Timeout)
		}(i, srv)
	}
	wg.Wait()
	c.JSON(http.StatusOK, results)
}

// BatchScript 批量执行脚本
func BatchScript(c *gin.Context) {
	var req struct {
		ServerIDs []uint `json:"server_ids"`
		Script    string `json:"script"`
		Shell     string `json:"shell"`
		Timeout   int    `json:"timeout"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.ServerIDs) == 0 || req.Script == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：需要 server_ids 和 script"})
		return
	}
	if req.Timeout <= 0 || req.Timeout > 600 {
		req.Timeout = 60
	}
	if req.Shell == "" {
		req.Shell = "bash"
	}
	var servers []models.Server
	database.DB.Where("id IN ?", req.ServerIDs).Find(&servers)
	var keyCfg models.SSHKeyConfig
	database.DB.First(&keyCfg)

	command := fmt.Sprintf("%s -s << 'CMDB_EOF'\n%s\nCMDB_EOF", req.Shell, req.Script)

	results := make([]BatchExecResult, len(servers))
	var wg sync.WaitGroup
	for i, srv := range servers {
		wg.Add(1)
		go func(idx int, s models.Server) {
			defer wg.Done()
			results[idx] = execOnServer(s, command, keyCfg, req.Timeout)
		}(i, srv)
	}
	wg.Wait()
	c.JSON(http.StatusOK, results)
}

// BatchUpload 批量文件分发
func BatchUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}
	defer file.Close()

	remotePath := c.PostForm("remote_path")
	if remotePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写远程目标路径"})
		return
	}
	serverIDsStr := c.PostForm("server_ids")
	if serverIDsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择目标服务器"})
		return
	}

	var serverIDs []uint
	for _, s := range strings.Split(serverIDsStr, ",") {
		s = strings.TrimSpace(s)
		var id uint
		if _, e := fmt.Sscanf(s, "%d", &id); e == nil && id > 0 {
			serverIDs = append(serverIDs, id)
		}
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	var servers []models.Server
	database.DB.Where("id IN ?", serverIDs).Find(&servers)
	var keyCfg models.SSHKeyConfig
	database.DB.First(&keyCfg)

	results := make([]BatchExecResult, len(servers))
	var wg sync.WaitGroup
	for i, srv := range servers {
		wg.Add(1)
		go func(idx int, s models.Server) {
			defer wg.Done()
			results[idx] = uploadToServer(s, fileContent, header.Filename, remotePath, keyCfg)
		}(i, srv)
	}
	wg.Wait()
	c.JSON(http.StatusOK, results)
}

func uploadToServer(s models.Server, content []byte, filename, remotePath string, keyCfg models.SSHKeyConfig) BatchExecResult {
	result := BatchExecResult{ServerID: s.ID, IP: s.IP, Hostname: s.Hostname}
	start := time.Now()

	client, err := dialSSH(s.IP, keyCfg)
	if err != nil {
		result.Error = err.Error()
		result.ExitCode = -1
		result.Duration = time.Since(start).Milliseconds()
		return result
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		result.Error = "SFTP 初始化失败: " + err.Error()
		result.ExitCode = -1
		result.Duration = time.Since(start).Milliseconds()
		return result
	}
	defer sftpClient.Close()

	// 确定目标路径
	destPath := remotePath
	if strings.HasSuffix(remotePath, "/") {
		sftpClient.MkdirAll(remotePath)
		destPath = remotePath + filename
	} else {
		dir := remotePath[:strings.LastIndex(remotePath, "/")+1]
		if dir != "" {
			sftpClient.MkdirAll(dir)
		}
	}

	dstFile, err := sftpClient.Create(destPath)
	if err != nil {
		result.Error = "创建远程文件失败: " + err.Error()
		result.ExitCode = -1
		result.Duration = time.Since(start).Milliseconds()
		return result
	}
	defer dstFile.Close()

	if _, err := dstFile.Write(content); err != nil {
		result.Error = "写入文件失败: " + err.Error()
		result.ExitCode = -1
		result.Duration = time.Since(start).Milliseconds()
		return result
	}

	result.Stdout = fmt.Sprintf("✓ 已上传至 %s (%d bytes)", destPath, len(content))
	result.Duration = time.Since(start).Milliseconds()
	return result
}

func dialSSH(ip string, keyCfg models.SSHKeyConfig) (*gossh.Client, error) {
	if keyCfg.PrivateKey == "" {
		return nil, fmt.Errorf("未配置 SSH 密钥，请在系统设置中配置")
	}
	signer, err := gossh.ParsePrivateKey([]byte(keyCfg.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("SSH 密钥解析失败: %v", err)
	}
	username := keyCfg.Username
	if username == "" {
		username = "root"
	}
	port := keyCfg.Port
	if port == "" {
		port = "22"
	}
	cfg := &gossh.ClientConfig{
		User:            username,
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	return gossh.Dial("tcp", fmt.Sprintf("%s:%s", ip, port), cfg)
}

func execOnServer(s models.Server, command string, keyCfg models.SSHKeyConfig, timeoutSec int) BatchExecResult {
	result := BatchExecResult{ServerID: s.ID, IP: s.IP, Hostname: s.Hostname}
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	client, err := dialSSH(s.IP, keyCfg)
	if err != nil {
		result.Error = err.Error()
		result.ExitCode = -1
		result.Duration = time.Since(start).Milliseconds()
		return result
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		result.Error = "创建 Session 失败: " + err.Error()
		result.ExitCode = -1
		result.Duration = time.Since(start).Milliseconds()
		return result
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	done := make(chan error, 1)
	go func() { done <- session.Run(command) }()

	select {
	case <-ctx.Done():
		result.Error = "执行超时"
		result.ExitCode = -1
	case err := <-done:
		if err != nil {
			if exitErr, ok := err.(*gossh.ExitError); ok {
				result.ExitCode = exitErr.ExitStatus()
			} else {
				result.ExitCode = -1
				result.Error = err.Error()
			}
		}
	}

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.Duration = time.Since(start).Milliseconds()
	return result
}
