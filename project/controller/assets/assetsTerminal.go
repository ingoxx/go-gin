package assets

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/logger"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/Lxb921006/Gin-bms/project/utils/encryption"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
	"sync"
	"time"
)

type CommandCapture struct {
	Buffer    []byte
	Lock      sync.Mutex
	StartTime time.Time
}

type outputSync struct {
	dataChan chan []byte
	quitChan chan struct{}
}

type WebTerminal struct {
	wsConn   *websocket.Conn
	ip       string
	remoteIp string
	user     string
	am       model.AssetsModel
	cmdCache *CommandCapture
	wsMutex  sync.Mutex
}

func NewWebTerminal(wc *websocket.Conn, user, remoteIp, ip string) *WebTerminal {
	return &WebTerminal{
		wsConn:   wc,
		ip:       ip,
		remoteIp: remoteIp,
		user:     user,
		cmdCache: &CommandCapture{
			StartTime: time.Now(),
		},
	}
}

func (wt *WebTerminal) Ssh() (err error) {
	wt.am, err = wt.am.GetServer(wt.ip)
	if err != nil {
		return err
	}

	config, err := wt.sshConfig()
	if err != nil {
		return err
	}

	session, err := wt.sshSession(config)
	if err != nil {
		return err
	}

	defer session.Close()

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()
	reader := io.MultiReader(stdout, stderr)

	if err := session.Shell(); err != nil {
		logger.Error(fmt.Sprintf("启动 shell 失败, errMsg: %s", err.Error()))
		if err := wt.wsConn.WriteMessage(websocket.TextMessage, []byte("启动 shell 失败: "+err.Error())); err != nil {
			logger.Error(fmt.Sprintf("websocket响应失败, errMsg: %s", err.Error()))
		}

		return err
	}

	go wt.handleOutput(reader)

	wt.handleInput(stdin, session)

	return nil
}

func (wt *WebTerminal) sshSession(config *ssh.ClientConfig) (*ssh.Session, error) {
	addr := fmt.Sprintf("%s:%d", wt.am.Ip, wt.am.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		logger.Error(fmt.Sprintf("SSH 连接失败: %s", err.Error()))
		if err := wt.wsConn.WriteMessage(websocket.TextMessage, []byte("SSH 连接失败: "+err.Error())); err != nil {
			logger.Error(fmt.Sprintf("websocket响应失败, errMsg: %s", err.Error()))
		}
		return nil, err
	}
	//defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		logger.Error(fmt.Sprintf("创建 SSH Session 失败: %s", err.Error()))
		if err := wt.wsConn.WriteMessage(websocket.TextMessage, []byte("创建 SSH Session 失败: "+err.Error())); err != nil {
			logger.Error(fmt.Sprintf("websocket响应失败, errMsg: %s", err.Error()))
		}

		return nil, err
	}

	cols, rows := 80, 24

	modes := ssh.TerminalModes{
		ssh.ECHO:    1,
		ssh.ECHOCTL: 0,
		ssh.IGNCR:   0,
		ssh.ICRNL:   1,
		ssh.OCRNL:   1,
		ssh.ONLCR:   1,
	}

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		logger.Error(fmt.Sprintf("请求伪终端失败: %s", err.Error()))
		if err := wt.wsConn.WriteMessage(websocket.TextMessage, []byte("请求伪终端失败: "+err.Error())); err != nil {
			logger.Error(fmt.Sprintf("websocket响应失败, errMsg: %s", err.Error()))
		}

		return nil, err
	}

	return session, nil
}

func (wt *WebTerminal) handleInput(stdin io.Writer, session *ssh.Session) {
	for {
		_, msg, err := wt.wsConn.ReadMessage()
		if err != nil {
			logger.Error(fmt.Sprintf("WebSocket 读取失败: %s", err.Error()))
			break
		}

		if len(msg) > 0 && msg[0] == '{' {
			var resize struct {
				Type string `json:"type"`
				Cols int    `json:"cols"`
				Rows int    `json:"rows"`
			}
			if json.Unmarshal(msg, &resize) == nil && resize.Type == "resize" {
				if err := session.WindowChange(resize.Rows, resize.Cols); err != nil {
					if err := wt.wsConn.WriteMessage(websocket.BinaryMessage, []byte(err.Error())); err != nil {
						logger.Error(fmt.Sprintf("WebSocket 响应失败: %s", err.Error()))
					}
				}
				continue
			}
		}

		wt.processInput(msg)

		if _, err := stdin.Write(msg); err != nil {
			logger.Error(fmt.Sprintf("WebSocket 输入写入 SSH 失败: %s", err.Error()))
			break
		}
	}
}

// 启动输出处理协程
func (wt *WebTerminal) handleOutput(stdout io.Reader) {
	s := &outputSync{
		dataChan: make(chan []byte, 100),
		quitChan: make(chan struct{}),
	}

	// 读取协程
	go func() {
		reader := bufio.NewReader(stdout)
		for {
			buf := make([]byte, 4096)
			n, err := reader.Read(buf)
			if err != nil {
				close(s.dataChan)
				return
			}

			// 深拷贝数据并发送到管道
			data := make([]byte, n)
			copy(data, buf[:n])
			s.dataChan <- data
		}
	}()

	// 发送协程
	go func() {
		defer close(s.quitChan)
		for data := range s.dataChan {
			// 同步发送保证顺序
			wt.wsMutex.Lock()
			if err := wt.wsConn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				wt.wsMutex.Unlock()
				return
			}
			wt.wsMutex.Unlock()

			// 调试输出（显示实际字节）
			fmt.Printf("发送数据块 [%d 字节]: %s\n", len(data), string(data))
		}
	}()

	// 等待退出信号
	<-s.quitChan
}

func (wt *WebTerminal) saveCmd(cmd string) error {
	var addLog model.OperateLogModel
	var record = make(map[string]string)

	record["user"] = wt.user
	record["url"] = fmt.Sprintf("终端命令操作审计, server ip: %s, run cmd: %s", wt.ip, cmd)
	record["ip"] = wt.remoteIp

	if err := addLog.AloneAddOperateLog(record); err != nil {
		return err
	}

	return nil
}

func (wt *WebTerminal) processInput(data []byte) {
	wt.cmdCache.Lock.Lock()
	defer wt.cmdCache.Lock.Unlock()

	for _, b := range data {
		switch b {
		case '\r': // 捕获回车键
			if len(wt.cmdCache.Buffer) > 0 {
				fullCmd := string(wt.cmdCache.Buffer)
				if fullCmd == "" {
					return
				}
				go func() {
					if err := wt.saveCmd(fullCmd); err != nil {
						logger.Error(fmt.Sprintf("保存cmd失败, cmd: %s, ip: %s, errMsg: %s", fullCmd, wt.ip, err.Error()))
					}
				}()
				wt.cmdCache.Buffer = []byte{}
			}
		case '\x08': // 处理退格键
			if len(wt.cmdCache.Buffer) > 0 {
				wt.cmdCache.Buffer = wt.cmdCache.Buffer[:len(wt.cmdCache.Buffer)-1]
			}
		case '\t': // 转换TAB为可读格式
			wt.cmdCache.Buffer = append(wt.cmdCache.Buffer, ' ', ' ')
		default:
			if b >= 32 && b <= 126 { // 只记录可打印ASCII字符
				wt.cmdCache.Buffer = append(wt.cmdCache.Buffer, b)
			}
		}
	}
}

func (wt *WebTerminal) sshConfig() (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod
	if wt.am.ConnectType == 1 {
		password, err := encryption.NewKeyPwdEncryption(wt.am.Password, 1).Decryption()
		if err != nil {
			return nil, err
		}
		auth = []ssh.AuthMethod{
			ssh.Password(password),
		}
	} else if wt.am.ConnectType == 2 {
		signer, err := wt.parsePrivateKey()
		if err != nil {
			return nil, err
		}
		auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		return nil, errors.New("未知登陆类型")
	}

	config := &ssh.ClientConfig{
		User:            wt.am.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 跳过主机密钥验证（不安全，生产环境需改进）
		Timeout:         5 * time.Second,
	}
	return config, nil
}

func (wt *WebTerminal) parsePrivateKey() (ssh.Signer, error) {
	key, err := encryption.NewKeyPwdEncryption(wt.am.Key, 1).Decryption()
	privateKeyBytes := []byte(key)
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	return signer, nil
}
