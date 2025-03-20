package assets

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ingoxx/go-gin/project/logger"
	"github.com/ingoxx/go-gin/project/model"
	"github.com/ingoxx/go-gin/project/tools/ddwarning"
	"github.com/ingoxx/go-gin/project/utils/encryption"
	"golang.org/x/crypto/ssh"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

// 如果没有任何操作, ssh session过期时间(分钟)
const sshSession = 30

type TerminalParser struct {
	outputBuffer   *bytes.Buffer
	ansiEscapeCode *regexp.Regexp
	promptPattern  *regexp.Regexp
	currentLine    []byte
}

func NewTerminalParser() *TerminalParser {
	return &TerminalParser{
		outputBuffer:   bytes.NewBuffer(nil),
		ansiEscapeCode: regexp.MustCompile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`),
		promptPattern:  regexp.MustCompile(`^.*?([#\$➤])\s`), // 匹配常见提示符结尾
	}
}

func (tp *TerminalParser) ParseOutput(data []byte) string {
	tp.outputBuffer.Write(data)
	rawOutput := tp.outputBuffer.Bytes()

	// 1. 去除所有ANSI转义码
	cleanOutput := tp.ansiEscapeCode.ReplaceAll(rawOutput, []byte{})

	// 2. 分割为多行处理
	lines := bytes.Split(cleanOutput, []byte{'\n'})
	if len(lines) == 0 {
		return ""
	}

	// 3. 取最后一行进行处理
	currentLine := bytes.TrimRight(lines[len(lines)-1], "\r")

	// 4. 动态识别并去除提示符
	if loc := tp.promptPattern.FindSubmatchIndex(currentLine); loc != nil {
		// 找到提示符结尾位置（符号位置 + 1 + 空格）
		promptEnd := loc[3] + 1
		if promptEnd <= len(currentLine) {
			tp.currentLine = currentLine[promptEnd:]
			return string(tp.currentLine)
		}
	}

	// 5. 保留未识别提示符的原始内容（调试用）
	tp.currentLine = currentLine
	return string(tp.currentLine)
}

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
	wsConn    *websocket.Conn
	am        model.AssetsModel
	ip        string
	cmdCache  *CommandCapture
	wsMutex   sync.Mutex
	ctx       *gin.Context
	signal    chan string
	parser    *TerminalParser
	sshClient *ssh.Client
	inputChan chan struct{}
}

func NewWebTerminal(wc *websocket.Conn, ctx *gin.Context) *WebTerminal {
	return &WebTerminal{
		wsConn: wc,
		ctx:    ctx,
		cmdCache: &CommandCapture{
			StartTime: time.Now(),
		},
		signal:    make(chan string),
		inputChan: make(chan struct{}),
		parser:    NewTerminalParser(),
	}
}

func (wt *WebTerminal) Ssh() (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wt.ip = wt.ctx.Query("ip")
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

	// 30分钟内没有任何输入就断开ssh, websocket连接
	go wt.monitorInputIsIdle(ctx)

	wt.handleInput(stdin, session)

	return nil
}

func (wt *WebTerminal) monitorInputIsIdle(ctx context.Context) {
	timer := time.NewTimer(time.Minute * sshSession)
	go func() {
		for {
			select {
			case <-timer.C:
				if wt.wsConn != nil && wt.sshClient != nil {
					wt.wsConn.Close()
					wt.sshClient.Close()
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-wt.inputChan:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(time.Minute * sshSession)
		case <-ctx.Done():
			return
		}
	}
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

	wt.sshClient = client

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
		ssh.ECHO:          1,     // 启用回显
		ssh.ECHOCTL:       0,     // 不将控制字符显示为^X
		ssh.TTY_OP_ISPEED: 14400, // 输入速度
		ssh.TTY_OP_OSPEED: 14400, // 输出速度
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

		// 检测窗口变化
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

		// 记录用户在终端的操作命令
		wt.processInput(msg)

		if _, err := stdin.Write(msg); err != nil {
			logger.Error(fmt.Sprintf("WebSocket 输入写入 SSH 失败: %s", err.Error()))
			break
		}

		// 监听输入
		wt.inputChan <- struct{}{}
	}
}

func (wt *WebTerminal) handleOutput(stdout io.Reader) {
	s := &outputSync{
		dataChan: make(chan []byte, 100),
		quitChan: make(chan struct{}),
	}

	go func() {
		reader := bufio.NewReader(stdout)
		for {
			buf := make([]byte, 4096)
			n, err := reader.Read(buf)
			if err != nil {
				close(wt.signal)
				close(s.dataChan)
				return
			}
			// 解析输出并更新当前输入行
			wt.parser.ParseOutput(buf[:n])

			// 深拷贝数据并发送到管道
			data := make([]byte, n)
			copy(data, buf[:n])
			s.dataChan <- data
		}
	}()

	go func() {
		defer close(s.quitChan)
		for {
			select {
			case data := <-s.dataChan:
				wt.wsMutex.Lock()
				if err := wt.wsConn.WriteMessage(websocket.BinaryMessage, data); err != nil {
					wt.wsMutex.Unlock()
					return
				}
				wt.wsMutex.Unlock()
			}
		}

	}()

	<-s.quitChan
}

func (wt *WebTerminal) processInput(data []byte) error {
	wt.cmdCache.Lock.Lock()
	defer wt.cmdCache.Lock.Unlock()

	// 从解析器获取当前实际输入行
	currentLine := wt.parser.currentLine
	wt.cmdCache.Buffer = make([]byte, len(currentLine))
	copy(wt.cmdCache.Buffer, currentLine)

	for _, b := range data {
		switch b {
		case '\r': // 捕获回车键
			if len(wt.cmdCache.Buffer) > 0 {
				fullCmd := string(wt.cmdCache.Buffer)

				if fullCmd == "" {
					return nil
				}
				go func() {
					if err := wt.saveCmd(fullCmd); err != nil {
						logger.Error(fmt.Sprintf("保存cmd失败, cmd: %s, ip: %s, errMsg: %s", fullCmd, wt.ip, err.Error()))
					}
				}()
				if err := wt.isDangerCmd(fullCmd); err != nil {
					return err
				}
				wt.cmdCache.Buffer = []byte{}
			}
		case '\x08': // 处理退格键
			if len(wt.cmdCache.Buffer) > 0 {
				wt.cmdCache.Buffer = wt.cmdCache.Buffer[:len(wt.cmdCache.Buffer)-1]
			}
		case '\t': // 转换TAB为可读格式
		default:
			if b >= 32 && b <= 126 { // 只记录可打印ASCII字符
				wt.cmdCache.Buffer = append(wt.cmdCache.Buffer, b)
			}
		}
	}

	return nil
}

func (wt *WebTerminal) saveCmd(cmd string) error {
	var addLog model.OperateLogModel
	var record = make(map[string]string)

	removeControlChars := func(r rune) rune {
		if r < 32 { // ASCII 0~31 是控制字符
			return -1 // 过滤掉该字符
		}
		return r
	}
	cleanStr := strings.Map(removeControlChars, cmd)
	record["user"] = wt.ctx.Query("user")
	record["url"] = fmt.Sprintf("%s, 终端命令操作日志, 操作服务器ip: %s, 执行命令: %v", wt.ctx.Request.URL.Path, wt.ip, cleanStr)
	record["ip"] = wt.ctx.RemoteIP()

	if err := addLog.AloneAddOperateLog(record); err != nil {
		return err
	}

	return nil
}

// isDangerCmd 高风险命令提醒
func (wt *WebTerminal) isDangerCmd(cmd string) error {
	if strings.HasPrefix(cmd, "rm") {
		msg := fmt.Sprintf("%s, 终端命令操作日志, 操作服务器ip: %s, 执行命令: %s", wt.ctx.Request.URL.Path, wt.ip, cmd)
		ddwarning.SendWarning(msg)
		return errors.New("danger cmd, not allow to execute")
	}

	return nil
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
