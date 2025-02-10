package assets

import (
	"errors"
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/logger"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/Lxb921006/Gin-bms/project/utils/encryption"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
)

type CommandCapture struct {
	Buffer    []byte
	Lock      sync.Mutex
	StartTime time.Time
}

type WebTerminal struct {
	wsConn   *websocket.Conn
	ip       string
	remoteIp string
	user     string
	am       model.AssetsModel
	cmdCache *CommandCapture
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

	// 连接 SSH
	addr := fmt.Sprintf("%s:%d", wt.am.Ip, wt.am.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		if err := wt.wsConn.WriteMessage(websocket.TextMessage, []byte("SSH 连接失败: "+err.Error())); err != nil {
			logger.Error(fmt.Sprintf("websocket响应失败 1, errMsg: %s", err.Error()))
			return err
		}
		return err
	}

	defer client.Close()

	// 创建 SSH Session
	session, err := client.NewSession()
	if err != nil {
		if err := wt.wsConn.WriteMessage(websocket.TextMessage, []byte("创建 SSH Session 失败: "+err.Error())); err != nil {
			logger.Error(fmt.Sprintf("websocket响应失败 2, errMsg: %s", err.Error()))
			return err
		}
		return err
	}
	defer session.Close()

	// 分配伪终端
	//modes := ssh.TerminalModes{
	//	ssh.ECHO:          1, // 启用回显
	//	ssh.TTY_OP_ISPEED: 14400,
	//	ssh.TTY_OP_OSPEED: 14400,
	//	ssh.ECHOCTL:       0, // 禁用控制字符回显
	//	ssh.OCRNL:         1, // 将回车转换为换行
	//	ssh.ONLCR:         1, // 将换行转换为回车换行（重要）
	//	ssh.ICRNL:         1, // 将回车转换为换行（输入方向）
	//	ssh.IXON:          1, // 启用软件流控制
	//}

	modes := ssh.TerminalModes{
		ssh.ECHO:    1, // 启用回显
		ssh.ECHOCTL: 0, // 禁用控制字符回显
		ssh.OCRNL:   1, // 转换回车为换行
		ssh.ONLCR:   1, // 转换换行为回车换行
		ssh.ICRNL:   1, // 转换输入回车为换行
		ssh.IGNCR:   0, // 不忽略回车
		ssh.INLCR:   0, // 不转换换行为回车
		ssh.ISIG:    1, // 启用信号
		ssh.IEXTEN:  1, // 启用扩展功能
		ssh.IXANY:   1, // 允许任何字符重启输出
		ssh.IXON:    1, // 启用软件流控
		ssh.VSTATUS: 0, // 禁用状态返回
	}
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		if err := wt.wsConn.WriteMessage(websocket.TextMessage, []byte("请求伪终端失败: "+err.Error())); err != nil {
			logger.Error(fmt.Sprintf("websocket响应失败 3, errMsg: %s", err.Error()))
			return err
		}
		return err
	}

	// 获取 SSH 输入输出
	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()

	// 开启 shell
	if err := session.Shell(); err != nil {
		if err := wt.wsConn.WriteMessage(websocket.TextMessage, []byte("启动 shell 失败: "+err.Error())); err != nil {
			logger.Error(fmt.Sprintf("websocket响应失败 4, errMsg: %s", err.Error()))
			return err
		}
		return err
	}

	// 读取 SSH 输出并写入 WebSocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			if err := wt.wsConn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				logger.Error(fmt.Sprintf("websocket响应失败 5, errMsg: %s", err.Error()))
				return
			}
		}
	}()

	// 读取 WebSocket 输入并写入 SSH
	for {
		_, msg, err := wt.wsConn.ReadMessage()
		if err != nil {
			break
		}
		// 新增命令处理
		wt.processInput(msg)
		if _, err := stdin.Write(msg); err != nil {
			logger.Error(fmt.Sprintf("WebSocket 输入并写入SSH失败, errMsg: %s", err.Error()))
			return err
		}
	}

	return
}

func (wt *WebTerminal) saveCmd(cmd string) error {
	var addLog model.OperateLogModel
	var record = make(map[string]string)

	record["user"] = wt.user
	record["url"] = fmt.Sprintf("server ip: %s, run cmd: %s", wt.ip, cmd)
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
				//duration := time.Since(wt.cmdCache.StartTime)
				//log.Printf("[CMD] %s | Duration: %v", fullCmd, duration)
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

func (wt *WebTerminal) isDangerCmd() {

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
