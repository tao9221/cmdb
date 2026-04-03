package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	gossh "golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// SSH连接参数（通过WebSocket第一条消息传入）
type SSHConnParams struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	// 公钥认证时传私钥内容（PEM格式）
	PrivateKey string `json:"private_key"`
	Rows       uint32 `json:"rows"`
	Cols       uint32 `json:"cols"`
}

// WebSocket消息类型
type WSMsg struct {
	Type string `json:"type"` // input / resize
	Data string `json:"data"`
	Rows uint32 `json:"rows"`
	Cols uint32 `json:"cols"`
}

func SSHTerminal(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer ws.Close()

	// 第一条消息：连接参数
	_, msg, err := ws.ReadMessage()
	if err != nil {
		writeWSError(ws, "读取连接参数失败")
		return
	}
	var params SSHConnParams
	if err := json.Unmarshal(msg, &params); err != nil {
		writeWSError(ws, "参数格式错误")
		return
	}
	if params.Port == "" {
		params.Port = "22"
	}
	if params.Rows == 0 {
		params.Rows = 24
	}
	if params.Cols == 0 {
		params.Cols = 80
	}

	// 构建SSH认证
	var authMethods []gossh.AuthMethod
	if params.PrivateKey != "" {
		signer, err := gossh.ParsePrivateKey([]byte(params.PrivateKey))
		if err != nil {
			writeWSError(ws, "私钥解析失败: "+err.Error())
			return
		}
		authMethods = append(authMethods, gossh.PublicKeys(signer))
	}
	if params.Password != "" {
		authMethods = append(authMethods, gossh.Password(params.Password))
	}
	if len(authMethods) == 0 {
		writeWSError(ws, "请提供密码或私钥")
		return
	}

	// 建立SSH连接
	sshCfg := &gossh.ClientConfig{
		User:            params.Username,
		Auth:            authMethods,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	client, err := gossh.Dial("tcp", params.Host+":"+params.Port, sshCfg)
	if err != nil {
		writeWSError(ws, "SSH连接失败: "+err.Error())
		return
	}
	defer client.Close()

	// 创建Session
	session, err := client.NewSession()
	if err != nil {
		writeWSError(ws, "创建Session失败: "+err.Error())
		return
	}
	defer session.Close()

	// 请求伪终端
	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", int(params.Rows), int(params.Cols), modes); err != nil {
		writeWSError(ws, "请求PTY失败: "+err.Error())
		return
	}

	// 管道
	stdin, err := session.StdinPipe()
	if err != nil {
		writeWSError(ws, "stdin pipe失败")
		return
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		writeWSError(ws, "stdout pipe失败")
		return
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		writeWSError(ws, "stderr pipe失败")
		return
	}

	if err := session.Shell(); err != nil {
		writeWSError(ws, "启动Shell失败: "+err.Error())
		return
	}

	// SSH输出 → WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				ws.WriteMessage(websocket.TextMessage, buf[:n])
			}
			if err != nil {
				if err != io.EOF {
					log.Println("stdout read error:", err)
				}
				ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
		}
	}()
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				ws.WriteMessage(websocket.TextMessage, buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()

	// WebSocket输入 → SSH stdin / resize
	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			break
		}
		// 尝试解析为控制消息
		var m WSMsg
		if json.Unmarshal(raw, &m) == nil && m.Type == "resize" {
			session.WindowChange(int(m.Rows), int(m.Cols))
			continue
		}
		// 普通输入
		stdin.Write(raw)
	}
}

func writeWSError(ws *websocket.Conn, msg string) {
	ws.WriteMessage(websocket.TextMessage, []byte("\r\n\033[31m[错误] "+msg+"\033[0m\r\n"))
}
