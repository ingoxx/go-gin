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
	MemUsedPercent float64 `json:"内存使用率(百分比)"`
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
	DiskUsedPercent float64 `json:"根目录使用率(百分比)"`
}

type ServerResourcesMonitor struct {
	Ip   string
	Days uint
}

func NewCpuLoadMonitor(clf GetServerStatusQuery) *ServerResourcesMonitor {
	return &ServerResourcesMonitor{
		Ip:   clf.Ip,
		Days: clf.Days,
	}
}

func (c *ServerResourcesMonitor) GetCpuLoadData() (CpuLoadData, error) {
	var rows []CpuChartRow
	var data CpuLoadData
	var startTs int64
	var endTs int64

	if c.Days > 0 {
		startTs = time.Now().Add(-time.Duration(c.Days) * 24 * time.Hour).Unix()
		endTs = time.Now().Unix()
	}

	key := fmt.Sprintf("%s%s", cpuKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	for i := len(values) - 1; i >= 0; i-- { // 倒序取出数据
		var entry CpuLoadEntry
		if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
			t := time.Unix(entry.Timestamp, 0).Format("01/02 15:04")
			if entry.Timestamp >= startTs && entry.Timestamp <= endTs {
				rows = append(rows, CpuChartRow{
					Time: t,
					Load: entry.Load1,
				})
			}
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
	var startTs int64
	var endTs int64

	if c.Days > 0 {
		startTs = time.Now().Add(-time.Duration(c.Days) * 24 * time.Hour).Unix()
		endTs = time.Now().Unix()
	}

	key := fmt.Sprintf("%s%s", ramKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	for i := len(values) - 1; i >= 0; i-- { // 倒序取出数据
		var entry MemUsageEntry
		if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
			t := time.Unix(entry.Timestamp, 0).Format("01/02 15:04")
			if entry.Timestamp >= startTs && entry.Timestamp <= endTs {
				rows = append(rows, MemUsageChartRow{
					Time:           t,
					MemUsedPercent: entry.MemUsedPercent,
				})
			}
		} else {
			return data, err
		}
	}

	data.Rows = rows
	data.Columns = []string{"时间", "内存使用率(百分比)"}

	return data, nil
}

func (c *ServerResourcesMonitor) GetDiskUsageData() (DiskUsageData, error) {
	var rows []DiskUsageChartRow
	var data DiskUsageData

	var startTs int64
	var endTs int64

	if c.Days > 0 {
		startTs = time.Now().Add(-time.Duration(c.Days) * 24 * time.Hour).Unix()
		endTs = time.Now().Unix()
	}

	key := fmt.Sprintf("%s%s", diskKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	for i := len(values) - 1; i >= 0; i-- { // 倒序取出数据
		var entry DiskUsageEntry
		if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
			t := time.Unix(entry.Timestamp, 0).Format("01/02 15:04")
			if entry.Timestamp >= startTs && entry.Timestamp <= endTs {
				rows = append(rows, DiskUsageChartRow{
					Time:            t,
					DiskUsedPercent: entry.DiskUsedPercent,
				})
			}
		} else {
			return data, err
		}
	}

	data.Rows = rows
	data.Columns = []string{"时间", "根目录使用率(百分比)"}

	return data, nil
}

func (c *ServerResourcesMonitor) getDayStartTimestamps(days int) []int64 {
	var points []int64
	now := time.Now()
	location := now.Location()
	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, location)
		points = append(points, dayStart.Unix())
	}
	return points
}
