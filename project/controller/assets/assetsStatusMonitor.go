package assets

import (
	"encoding/json"
	"fmt"
	"github.com/ingoxx/go-gin/project/dao"
	"sort"
	"strconv"
	"time"
)

const (
	cpuKey  = "cpu_loads_"
	ramKey  = "mem_usage_"
	diskKey = "disk_usage_" // 这里只监控根目录的使用率
	maxData = 8
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
	Data float64 `json:"CPU负载"`
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
	Time string  `json:"时间"`
	Data float64 `json:"内存使用率(百分比)"`
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
	Time string  `json:"时间"`
	Data float64 `json:"根目录使用率(百分比)"`
}

type ServerResourcesMonitor struct {
	Ip   string
	Days uint
}

type getServerData struct {
	Rows  interface{}
	Entry interface{}
	Key   string
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

	key := fmt.Sprintf("%s%s", cpuKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	var dateRangeCount = make(map[int64]float64)
	dateRange := c.getDayStartTimestamps()

	for i1 := 0; i1 < len(dateRange); i1++ {
		var dataCount int                  // 统计某个日期的累加次数获取平均值
		var dataCurrentDayCount int        // 获取当天的最新50条
		for i := 0; i < len(values); i++ { // 倒序取出数据
			var entry CpuLoadEntry

			if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
				t := time.Unix(entry.Timestamp, 0).Format("01/02 15:04:05")
				if len(dateRange) > 1 { // 获取2天以上的数据
					if dateRange[(len(dateRange)-1)] == dateRange[i1] { // 获取当天的最新数据
						if dataCurrentDayCount == maxData {
							break
						}
						rows = append(rows, CpuChartRow{
							Time: t,
							Data: entry.Load1,
						})
						dataCurrentDayCount++
					} else { // 获取近几天除了当天的数据
						end := time.Unix(dateRange[i1], 0).Add(-time.Duration(1) * 24 * time.Hour).Unix()
						if _, ok := dateRangeCount[dateRange[i1]]; !ok {
							dateRangeCount[dateRange[i1]] = 0
						}
						if entry.Timestamp >= end && entry.Timestamp <= dateRange[i1] {
							dateRangeCount[dateRange[i1]] += entry.Load1
						}
						dataCount++
					}
				} else { // 这里是默认获取1天最新的数据
					if len(rows) == maxData {
						break
					}
					rows = append(rows, CpuChartRow{
						Time: t,
						Data: entry.Load1,
					})
				}
			} else {
				return data, err
			}
		}
		// 获取2天以上的数据并计算平均值
		if len(dateRange) > 1 && dateRange[(len(dateRange)-1)] != dateRange[i1] {
			var load1 float64
			t := time.Unix(dateRange[i1], 0).Format("01/02 15:04:05")
			if dateRangeCount[dateRange[i1]] == 0 {
				load1 = 0
			} else {
				load1, err = strconv.ParseFloat(fmt.Sprintf("%.2f", dateRangeCount[dateRange[i1]]/float64(dataCount)), 64)
				if err != nil {
					return data, err
				}
			}
			rows = append(rows, CpuChartRow{
				Time: t,
				Data: load1,
			})
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

	var dateRangeCount = make(map[int64]float64)
	dateRange := c.getDayStartTimestamps()

	for i1 := 0; i1 < len(dateRange); i1++ {
		var dataCount int                  // 统计某个日期的累加次数获取平均值
		var dataCurrentDayCount int        // 获取当天的最新50条
		for i := 0; i < len(values); i++ { // 倒序取出数据
			var entry MemUsageEntry

			if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
				t := time.Unix(entry.Timestamp, 0).Format("01/02 15:04:05")
				if len(dateRange) > 1 { // 获取2天以上的数据
					if dateRange[(len(dateRange)-1)] == dateRange[i1] { // 获取当天的最新数据
						if dataCurrentDayCount == maxData {
							break
						}
						rows = append(rows, MemUsageChartRow{
							Time: t,
							Data: entry.MemUsedPercent,
						})
						dataCurrentDayCount++
					} else { // 获取近几天除了当天的数据
						end := time.Unix(dateRange[i1], 0).Add(-time.Duration(1) * 24 * time.Hour).Unix()
						if _, ok := dateRangeCount[dateRange[i1]]; !ok {
							dateRangeCount[dateRange[i1]] = 0
						}
						if entry.Timestamp >= end && entry.Timestamp <= dateRange[i1] {
							dateRangeCount[dateRange[i1]] += entry.MemUsedPercent
						}
						dataCount++
					}
				} else { // 这里是默认获取1天最新的数据
					if len(rows) == maxData {
						break
					}
					rows = append(rows, MemUsageChartRow{
						Time: t,
						Data: entry.MemUsedPercent,
					})
				}
			} else {
				return data, err
			}
		}
		// 获取2天以上的数据并计算平均值
		if len(dateRange) > 1 && dateRange[(len(dateRange)-1)] != dateRange[i1] {
			var load1 float64
			t := time.Unix(dateRange[i1], 0).Format("01/02 15:04:05")
			if dateRangeCount[dateRange[i1]] == 0 {
				load1 = 0
			} else {
				load1, err = strconv.ParseFloat(fmt.Sprintf("%.2f", dateRangeCount[dateRange[i1]]/float64(dataCount)), 64)
				if err != nil {
					return data, err
				}
			}
			rows = append(rows, MemUsageChartRow{
				Time: t,
				Data: load1,
			})
		}
	}

	data.Rows = rows
	data.Columns = []string{"时间", "内存使用率(百分比)"}

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

	var dateRangeCount = make(map[int64]float64)
	dateRange := c.getDayStartTimestamps()

	for i1 := 0; i1 < len(dateRange); i1++ {
		var dataCount int                  // 统计某个日期的累加次数获取平均值
		var dataCurrentDayCount int        // 获取当天的最新50条
		for i := 0; i < len(values); i++ { // 倒序取出数据
			var entry DiskUsageEntry

			if err := json.Unmarshal([]byte(values[i]), &entry); err == nil {
				t := time.Unix(entry.Timestamp, 0).Format("01/02 15:04:05")
				if len(dateRange) > 1 { // 获取2天以上的数据
					if dateRange[(len(dateRange)-1)] == dateRange[i1] { // 获取当天的最新数据
						if dataCurrentDayCount == maxData {
							break
						}
						rows = append(rows, DiskUsageChartRow{
							Time: t,
							Data: entry.DiskUsedPercent,
						})
						dataCurrentDayCount++
					} else { // 获取近几天除了当天的数据
						end := time.Unix(dateRange[i1], 0).Add(-time.Duration(1) * 24 * time.Hour).Unix()
						if _, ok := dateRangeCount[dateRange[i1]]; !ok {
							dateRangeCount[dateRange[i1]] = 0
						}
						if entry.Timestamp >= end && entry.Timestamp <= dateRange[i1] {
							dateRangeCount[dateRange[i1]] += entry.DiskUsedPercent
						}
						dataCount++
					}
				} else { // 这里是默认获取1天最新的数据
					if len(rows) == maxData {
						break
					}
					rows = append(rows, DiskUsageChartRow{
						Time: t,
						Data: entry.DiskUsedPercent,
					})
				}
			} else {
				return data, err
			}
		}
		// 获取2天以上的数据并计算平均值
		if len(dateRange) > 1 && dateRange[(len(dateRange)-1)] != dateRange[i1] {
			var load1 float64
			t := time.Unix(dateRange[i1], 0).Format("01/02 15:04:05")
			if dateRangeCount[dateRange[i1]] == 0 {
				load1 = 0
			} else {
				load1, err = strconv.ParseFloat(fmt.Sprintf("%.2f", dateRangeCount[dateRange[i1]]/float64(dataCount)), 64)
				if err != nil {
					return data, err
				}
			}
			rows = append(rows, DiskUsageChartRow{
				Time: t,
				Data: load1,
			})
		}
	}

	data.Rows = rows
	data.Columns = []string{"时间", "根目录使用率(百分比)"}

	return data, nil
}

func (c *ServerResourcesMonitor) getDayStartTimestamps() []int64 {
	var points []int64
	now := time.Now()
	location := now.Location()
	for i := int(c.Days) - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 59, location)
		points = append(points, dayStart.Unix())
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i] < points[j]
	})

	return points
}
