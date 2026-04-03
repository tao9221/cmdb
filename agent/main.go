// CMDB Agent - 自动采集并上报服务器信息
// 用法: ./agent -server http://localhost:8080 -interval 60
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Report struct {
	Hostname     string  `json:"hostname"`
	IP           string  `json:"ip"`
	Vendor       string  `json:"vendor"`
	Model        string  `json:"model"`
	OS           string  `json:"os"`
	CPUModel     string  `json:"cpu_model"`
	CPUCores     int     `json:"cpu_cores"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemTotal     int64   `json:"mem_total"`
	MemUsed      int64   `json:"mem_used"`
	DiskTotal    int64   `json:"disk_total"`
	DiskUsed     int64   `json:"disk_used"`
	NetIn        int64   `json:"net_in"`
	NetOut       int64   `json:"net_out"`
	AgentVersion string  `json:"agent_version"`
}

const version = "1.0.0"

func main() {
	server := flag.String("server", "http://localhost:8080", "CMDB服务器地址")
	interval := flag.Int("interval", 60, "上报间隔(秒)")
	flag.Parse()

	log.Printf("CMDB Agent v%s 启动, 上报到 %s, 间隔 %ds", version, *server, *interval)

	for {
		report := collect()
		if err := send(*server, report); err != nil {
			log.Printf("上报失败: %v", err)
		} else {
			log.Printf("上报成功: %s (%s)", report.Hostname, report.IP)
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

func collect() Report {
	r := Report{AgentVersion: version, OS: runtime.GOOS + "/" + runtime.GOARCH}

	r.Hostname, _ = os.Hostname()
	r.IP = getIP()
	r.CPUCores = runtime.NumCPU()
	r.CPUModel = getCPUModel()
	r.CPUUsage = getCPUUsage()
	r.MemTotal, r.MemUsed = getMemory()
	r.DiskTotal, r.DiskUsed = getDisk()
	r.Vendor, r.Model = getVendor()
	r.NetIn, r.NetOut = getNetworkSpeed()

	return r
}

func getIP() string {
	// 方法1：通过出口路由判断主 IP（最准确，不发真实数据包）
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			ip := addr.IP.String()
			if ip != "" && ip != "0.0.0.0" {
				return ip
			}
		}
	}

	// 方法2：遍历网卡，过滤虚拟/容器网卡，取第一个物理网卡 IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	// 虚拟网卡名称前缀黑名单
	virtualPrefixes := []string{"docker", "veth", "virbr", "br-", "lo", "tun", "tap", "vbox", "vmnet", "dummy"}
	for _, iface := range ifaces {
		name := strings.ToLower(iface.Name)
		skip := false
		for _, prefix := range virtualPrefixes {
			if strings.HasPrefix(name, prefix) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			return ip.String()
		}
	}
	return "127.0.0.1"
}

func getCPUModel() string {
	if runtime.GOOS == "linux" {
		out, _ := exec.Command("sh", "-c", "grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2").Output()
		return strings.TrimSpace(string(out))
	}
	if runtime.GOOS == "windows" {
		out, _ := exec.Command("wmic", "cpu", "get", "name", "/value").Output()
		for _, l := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(l, "Name=") {
				return strings.TrimSpace(strings.TrimPrefix(l, "Name="))
			}
		}
	}
	return runtime.GOARCH
}

func getCPUUsage() float64 {
	if runtime.GOOS == "linux" {
		// 简单采样
		read := func() (idle, total uint64) {
			out, _ := exec.Command("sh", "-c", "head -1 /proc/stat").Output()
			fields := strings.Fields(string(out))
			if len(fields) < 8 {
				return
			}
			vals := make([]uint64, len(fields)-1)
			for i, f := range fields[1:] {
				vals[i], _ = strconv.ParseUint(f, 10, 64)
				total += vals[i]
			}
			idle = vals[3]
			return
		}
		idle1, total1 := read()
		time.Sleep(200 * time.Millisecond)
		idle2, total2 := read()
		if total2-total1 == 0 {
			return 0
		}
		return float64(total2-total1-(idle2-idle1)) / float64(total2-total1) * 100
	}
	return 0
}

func getMemory() (total, used int64) {
	if runtime.GOOS == "linux" {
		out, _ := exec.Command("sh", "-c", "free -b | grep Mem").Output()
		fields := strings.Fields(string(out))
		if len(fields) >= 3 {
			total, _ = strconv.ParseInt(fields[1], 10, 64)
			used, _ = strconv.ParseInt(fields[2], 10, 64)
		}
	} else if runtime.GOOS == "windows" {
		out, _ := exec.Command("wmic", "OS", "get", "TotalVisibleMemorySize,FreePhysicalMemory", "/value").Output()
		var freeKB, totalKB int64
		for _, l := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(l, "TotalVisibleMemorySize=") {
				totalKB, _ = strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(l, "TotalVisibleMemorySize=")), 10, 64)
			}
			if strings.HasPrefix(l, "FreePhysicalMemory=") {
				freeKB, _ = strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(l, "FreePhysicalMemory=")), 10, 64)
			}
		}
		total = totalKB * 1024
		used = (totalKB - freeKB) * 1024
	}
	return
}

func getDisk() (total, used int64) {
	if runtime.GOOS == "linux" {
		out, _ := exec.Command("sh", "-c", "df -B1 / | tail -1").Output()
		fields := strings.Fields(string(out))
		if len(fields) >= 3 {
			total, _ = strconv.ParseInt(fields[1], 10, 64)
			used, _ = strconv.ParseInt(fields[2], 10, 64)
		}
	} else if runtime.GOOS == "windows" {
		out, _ := exec.Command("wmic", "logicaldisk", "where", "DeviceID='C:'", "get", "Size,FreeSpace", "/value").Output()
		var free, size int64
		for _, l := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(l, "Size=") {
				size, _ = strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(l, "Size=")), 10, 64)
			}
			if strings.HasPrefix(l, "FreeSpace=") {
				free, _ = strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(l, "FreeSpace=")), 10, 64)
			}
		}
		total = size
		used = size - free
	}
	return
}

func getVendor() (vendor, model string) {
	if runtime.GOOS == "linux" {
		v, _ := exec.Command("sh", "-c", "cat /sys/class/dmi/id/sys_vendor 2>/dev/null").Output()
		m, _ := exec.Command("sh", "-c", "cat /sys/class/dmi/id/product_name 2>/dev/null").Output()
		vendor = strings.TrimSpace(string(v))
		model = strings.TrimSpace(string(m))
	} else if runtime.GOOS == "windows" {
		out, _ := exec.Command("wmic", "computersystem", "get", "Manufacturer,Model", "/value").Output()
		for _, l := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(l, "Manufacturer=") {
				vendor = strings.TrimSpace(strings.TrimPrefix(l, "Manufacturer="))
			}
			if strings.HasPrefix(l, "Model=") {
				model = strings.TrimSpace(strings.TrimPrefix(l, "Model="))
			}
		}
	}
	if vendor == "" {
		vendor = "Unknown"
	}
	return
}

func send(server string, r Report) error {
	data, _ := json.Marshal(r)
	resp, err := http.Post(fmt.Sprintf("%s/api/agent/report", server), "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// 采集网络速率（bytes/s），采样间隔1秒
func getNetworkSpeed() (in, out int64) {
	if runtime.GOOS == "linux" {
		readNet := func() (rx, tx int64) {
			out, _ := exec.Command("sh", "-c",
				"cat /proc/net/dev | awk 'NR>2 && !/lo/{rx+=$2; tx+=$10} END{print rx, tx}'").Output()
			fields := strings.Fields(string(out))
			if len(fields) >= 2 {
				rx, _ = strconv.ParseInt(fields[0], 10, 64)
				tx, _ = strconv.ParseInt(fields[1], 10, 64)
			}
			return
		}
		rx1, tx1 := readNet()
		time.Sleep(1 * time.Second)
		rx2, tx2 := readNet()
		return rx2 - rx1, tx2 - tx1
	}
	if runtime.GOOS == "windows" {
		readNet := func() (rx, tx int64) {
			out, _ := exec.Command("powershell", "-Command",
				"Get-NetAdapterStatistics | Measure-Object -Property ReceivedBytes,SentBytes -Sum | Select-Object -ExpandProperty Sum").Output()
			lines := strings.Fields(string(out))
			if len(lines) >= 2 {
				rx, _ = strconv.ParseInt(lines[0], 10, 64)
				tx, _ = strconv.ParseInt(lines[1], 10, 64)
			}
			return
		}
		rx1, tx1 := readNet()
		time.Sleep(1 * time.Second)
		rx2, tx2 := readNet()
		return rx2 - rx1, tx2 - tx1
	}
	return 0, 0
}
