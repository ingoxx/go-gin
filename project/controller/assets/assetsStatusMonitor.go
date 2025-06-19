package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ingoxx/go-gin/project/dao"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	cpuKey            = "cpu_loads_"
	ramKey            = "mem_usage_"
	diskKey           = "disk_usage_" // 这里只监控根目录的使用率
	maxData           = 8
	maxCurrentDayData = 3000
)

type ChartsData struct {
	Columns []string     `json:"columns"`
	Rows    []ChartsRows `json:"rows"`
}

type ChartsRows struct {
	Time string  `json:"time"`
	Data float64 `json:"data"`
}

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
	Time string  `json:"time"`
	Data float64 `json:"data"`
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
	lock  *sync.Mutex
	wg    *sync.WaitGroup
	limit chan struct{}
	Ip    string
	Days  uint
}

type GetServerData struct {
	CpuData  ChartsData
	RamData  ChartsData
	DiskData ChartsData
	Error    error
}

func NewCpuLoadMonitor(clf GetServerStatusQuery) *ServerResourcesMonitor {
	return &ServerResourcesMonitor{
		Ip:    clf.Ip,
		Days:  clf.Days,
		lock:  new(sync.Mutex),
		wg:    new(sync.WaitGroup),
		limit: make(chan struct{}, 10),
	}
}

func (c *ServerResourcesMonitor) getCpuLoadData() (ChartsData, error) {
	var data ChartsData
	var entry CpuLoadEntry
	key := fmt.Sprintf("%s%s", cpuKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	data, err = c.generic(entry, values)
	if err != nil {
		return data, err
	}

	return data, nil

}

func (c *ServerResourcesMonitor) getMemUsageData() (ChartsData, error) {
	var data ChartsData
	var entry MemUsageEntry
	key := fmt.Sprintf("%s%s", ramKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	data, err = c.generic(entry, values)
	if err != nil {
		return data, err
	}

	return data, nil
}

func (c *ServerResourcesMonitor) getDiskUsageData() (ChartsData, error) {
	var data ChartsData
	var entry DiskUsageEntry

	key := fmt.Sprintf("%s%s", diskKey, c.Ip)
	values, err := dao.Rds.GetServerCpuLoadData(key)
	if err != nil {
		return data, err
	}

	data, err = c.generic(entry, values)
	if err != nil {
		return data, err
	}

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

func (c *ServerResourcesMonitor) GetServerStatus() (map[string]*GetServerData, error) {
	var fd = make(map[string]*GetServerData)

	var f = []func(){
		func() {
			defer c.wg.Done()
			var gsd = new(GetServerData)
			data, err := c.getDiskUsageData()
			if err != nil {
				gsd.Error = errors.Join(err)
				return
			}
			gsd.DiskData = data
			c.lock.Lock()
			fd["disk"] = gsd
			c.lock.Unlock()
		},
		func() {
			defer c.wg.Done()
			var gsd = new(GetServerData)
			data, err := c.getMemUsageData()
			if err != nil {
				gsd.Error = errors.Join(err)
				return
			}
			gsd.RamData = data
			c.lock.Lock()
			fd["ram"] = gsd
			c.lock.Unlock()
		},
		func() {
			defer c.wg.Done()
			var gsd = new(GetServerData)
			data, err := c.getCpuLoadData()
			if err != nil {
				gsd.Error = errors.Join(err)
				return
			}
			gsd.CpuData = data
			c.lock.Lock()
			fd["cpu"] = gsd
			c.lock.Unlock()
		},
	}

	for _, v1 := range f {
		c.wg.Add(1)
		go func(fc func()) {
			fc()
		}(v1)
	}

	c.wg.Wait()

	for v := range fd {
		if fd[v].Error != nil {
			return fd, fd[v].Error
		}
	}

	return fd, nil
}

func (c *ServerResourcesMonitor) generic(tp interface{}, values []string) (ChartsData, error) {
	var data ChartsData
	var rows []ChartsRows
	var err error

	var dateRangeCount = make(map[int64]float64)
	dateRange := c.getDayStartTimestamps()
	for i1 := len(dateRange) - 1; i1 >= 0; i1-- {
		var dataCount int // 统计某天的数据累加获取平均值
		//var dataCurrentDayCount int        // 获取当天的最新数据
		for i := 0; i < len(values); i++ { // 倒序取出数据
			//var entry DiskUsageEntry
			parserEntry, f, err := c.parserEntry(tp, values[i])
			if err == nil {
				//t := time.Unix(parserEntry, 0).Format("01/02 15:04:05")
				if len(dateRange) > 1 { // 获取2天以上的数据
					if dateRange[(len(dateRange)-1)] != dateRange[i1] { // 获取当天的最新数据
						// 获取近几天除了当天的数据
						end := time.Unix(dateRange[i1], 0).Add(-time.Duration(1) * 24 * time.Hour).Unix()
						if _, ok := dateRangeCount[dateRange[i1]]; !ok {
							dateRangeCount[dateRange[i1]] = 0
						}
						if parserEntry >= end && parserEntry <= dateRange[i1] {
							dateRangeCount[dateRange[i1]] += f
							dataCount++
						}
					}
				}
			} else {
				return data, err
			}
		}
		// 获取2天以上的数据并计算平均值
		if len(dateRange) > 1 && dateRange[(len(dateRange)-1)] != dateRange[i1] {
			var load1 float64
			t := time.Unix(dateRange[i1], 0).Format("01/02 15:04")
			if dateRangeCount[dateRange[i1]] == 0 {
				load1 = 0
			} else {
				load1, err = strconv.ParseFloat(fmt.Sprintf("%.2f", dateRangeCount[dateRange[i1]]/float64(dataCount)), 64)
				if err != nil {
					return data, err
				}
			}
			rows = append(rows, ChartsRows{
				Time: t,
				Data: load1,
			})
		}
	}

	var qd []ChartsRows
	if len(values) > 3000 {
		qd, err = c.querySegmentedData(values[:3000], tp)
		if err != nil {
			return data, err
		}
	} else {
		qd, err = c.querySegmentedData(values[:], tp)
		if err != nil {
			return data, err
		}
	}

	rows = append(rows, qd...)

	sort.Slice(rows, func(i, j int) bool {
		it := rows[i].Time
		jt := rows[j].Time
		ift := fmt.Sprintf("%d/%s", time.Now().Year(), it)
		jft := fmt.Sprintf("%d/%s", time.Now().Year(), jt)
		t1, err := time.Parse("2006/01/02 15:04", ift)
		if err != nil {
			panic(err)
		}

		t2, err := time.Parse("2006/01/02 15:04", jft)
		if err != nil {
			panic(err)
		}

		return t1.Unix() < t2.Unix()
	})

	data.Rows = rows
	data.Columns = []string{"time", "data"}

	return data, nil
}

func (c *ServerResourcesMonitor) parserEntry(target interface{}, val string) (int64, float64, error) {
	var ts int64
	var vd float64
	switch target.(type) {
	case DiskUsageEntry:
		var entry DiskUsageEntry
		if err := json.Unmarshal([]byte(val), &entry); err != nil {
			return ts, vd, err
		}
		return entry.Timestamp, entry.DiskUsedPercent, nil
	case CpuLoadEntry:
		var entry CpuLoadEntry
		if err := json.Unmarshal([]byte(val), &entry); err != nil {
			return ts, vd, err
		}
		return entry.Timestamp, entry.Load1, nil
	case MemUsageEntry:
		var entry MemUsageEntry
		if err := json.Unmarshal([]byte(val), &entry); err != nil {
			return ts, vd, err
		}
		return entry.Timestamp, entry.MemUsedPercent, nil
	}

	return ts, vd, nil
}

func (c *ServerResourcesMonitor) querySegmentedData(values []string, tp interface{}) ([]ChartsRows, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	end := time.Now().Unix()
	start := end - 24*3600

	segmentCount := 7
	interval := int64(4 * 3600) // 4 小时
	segments := make([][]float64, segmentCount)
	labels := make([]string, segmentCount)

	for i := 0; i < segmentCount; i++ {
		segStart := start + int64(i)*interval
		labels[i] = time.Unix(segStart, 0).Format("01/02 15:04") // eg: 06-16 00:00
	}

	for i := 0; i < len(values); i++ {
		timestamp, f, err := c.parserEntry(tp, values[i])
		if err != nil {
			return []ChartsRows{}, err
		}

		if timestamp < start || timestamp > end {
			continue
		}

		index := int((timestamp - start) / interval)
		if index >= 0 && index < segmentCount {
			segments[index] = append(segments[index], f)
		}
	}

	var rows []ChartsRows
	for i, group := range segments {
		if len(group) == 0 {
			timestamp, f, err := c.parserEntry(tp, values[0])
			t := time.Unix(timestamp, 0).Format("01/02 15:04")
			if err != nil {
				return rows, err
			}
			rows = append(rows, ChartsRows{
				Time: t,
				Data: f,
			})
			continue
		}
		sum := 0.0
		for _, item := range group {
			sum += item
		}
		avg := sum / float64(len(group))
		rows = append(rows, ChartsRows{
			Time: labels[i],
			Data: avg,
		})
	}

	return rows, nil
}
