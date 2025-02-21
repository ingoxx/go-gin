package ddwarning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type WarnInfo struct {
	webHook string
	data    string
}

func NewWarnInfo(data string) *WarnInfo {
	return &WarnInfo{
		webHook: "https://oapi.dingtalk.com/robot/send?access_token=4797ec430d2b74acbfaa084960ca389d884068a0a1d6115ad10c4ac7ffabc395",
		data:    data,
	}
}

func SendWarning(data string) {
	if err := NewWarnInfo(data).sendWarningInfo(); err != nil {
		log.Printf("sending warning message failed, errMsg: %s\n", err.Error())
	}
}

func (w *WarnInfo) sendWarningInfo() error {
	data := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": fmt.Sprintf("告警信息\n告警时间：%s\n告警内容: %s", time.Now().Format("2006-01-02 15:04:05"), w.data),
		},
		"at": map[string]interface{}{
			"atMobiles": "15889709122",
			"isAtAll":   false,
		},
	}

	if err := w.send(data); err != nil {
		return err
	}
	return nil
}

func (w *WarnInfo) send(data map[string]interface{}) error {
	b, _ := json.Marshal(data)
	body := bytes.NewBuffer(b)
	resp, _ := http.Post(w.webHook, "application/json", body)
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}
