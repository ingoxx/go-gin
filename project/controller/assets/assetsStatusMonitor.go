package assets

import (
	"encoding/json"
	"fmt"
	"github.com/ingoxx/go-gin/project/dao"
	"time"
)

const (
	cpuKey  = "cpu_loads_"
	ramKey  = "mem_usage_"
	diskKey = "disk_usage_" // 这里只监控根目录的使用率
)

// CpuLoadEntry cpu负载监控可视化
type CpuLoadEntry struct {
	Timestamp int64   `json:"timestamp"`
	Load1     float64 `json:"load1"`
	Load5     float64 `json:"load5"`
	Load15    float64 `json:"load15"`
}

type CpuLoadData struct {
	Columns []string      `json:"columns"`
	Rows    []CpuChartRow `json:"rows"`
}

type CpuChartRow struct {
	Time string  `json:"时间"`
	Load float64 `json:"CPU负载"`
}

// MemUsageEntry 内存监控可视化
type MemUsageEntry struct {
	Timestamp      int64   `json:"timestamp"`
	MemUsedPercent float64 `json:"mem_used_percent"`
}

type MemUsageData struct {
	Columns []string           `json:"columns"`
	Rows    []MemUsageChartRow `json:"rows"`
}

type MemUsageChartRow struct {
	Time           string  `json:"时间"`
	MemUsedPercent float64 `json:"内存使用率"`
}

// DiskUsageEntry 根目录监控可视化
type DiskUsageEntry struct {
	Timestamp       int64   `json:"timestamp"`
	DiskUsedPercent float64 `json:"disk_used_percent"`
}

type DiskUsageData struct {
	Columns []string            `json:"columns"`
	Rows    []DiskUsageChartRow `json:"rows"`
}

type DiskUsageChartRow struct {
	Time            string  `json:"时间"`
	DiskUsedPercent float64 `json:"根目录使用率"`
}

type ServerResourcesMonitor struct {
	Ip string
}

func NewCpuLoadMonitor(ip string) *ServerResourcesMonitor {
	return &ServerResourcesMonitor{
		Ip: ip,
	}
}

func (c *ServerResourcesMonitor) GetCpuLoadData() (CpuLoadData, error) {
	var rows []CpuChartRow
	var data CpuLoadData
	key := fmt.Sprintf("%s%s", cpuKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	for i := len(values) - 1; i >= 0; i-- { // 倒序取出数据
		var entry CpuLoadEntry
		if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
			t := time.Unix(entry.Timestamp, 0).Format("15:04:05")
			rows = append(rows, CpuChartRow{
				Time: t,
				Load: entry.Load1, // 你也可以换成 Load5 / Load15
			})
		} else {
			return data, err
		}
	}

	data.Rows = rows
	data.Columns = []string{"时间", "CPU负载"}

	return data, nil
}

func (c *ServerResourcesMonitor) GetMemUsageData() (MemUsageData, error) {
	var rows []MemUsageChartRow
	var data MemUsageData
	key := fmt.Sprintf("%s%s", ramKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	for i := len(values) - 1; i >= 0; i-- { // 倒序取出数据
		var entry MemUsageEntry
		if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
			t := time.Unix(entry.Timestamp, 0).Format("15:04:05")
			rows = append(rows, MemUsageChartRow{
				Time:           t,
				MemUsedPercent: entry.MemUsedPercent,
			})
		} else {
			return data, err
		}
	}

	data.Rows = rows
	data.Columns = []string{"时间", "内存使用率"}

	return data, nil
}

func (c *ServerResourcesMonitor) GetDiskUsageData() (DiskUsageData, error) {
	var rows []DiskUsageChartRow
	var data DiskUsageData
	key := fmt.Sprintf("%s%s", diskKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	for i := len(values) - 1; i >= 0; i-- { // 倒序取出数据
		var entry DiskUsageEntry
		if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
			t := time.Unix(entry.Timestamp, 0).Format("15:04:05")
			rows = append(rows, DiskUsageChartRow{
				Time:            t,
				DiskUsedPercent: entry.DiskUsedPercent,
			})
		} else {
			return data, err
		}
	}

	data.Rows = rows
	data.Columns = []string{"时间", "根目录使用率"}

	return data, nil
}
