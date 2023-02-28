package main

import (
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/cpu"
	"time"
	net2 "github.com/shirou/gopsutil/net"
	"strings"
	"regexp"
	"strconv"
	"github.com/shirou/gopsutil/host"
	"fmt"
	"io"
	"net"
	"encoding/json"
	"github.com/Unknwon/goconfig"
	"log"
	"path/filepath"
	"os"
	"io/ioutil"
	"sort"
)

// 发送数据Json结构体
type SentData struct {
	Cpu         float64 `json:"cpu"`
	NetworkRx   uint64  `json:"network_rx"`
	NetworkTx   uint64  `json:"network_tx"`
	NetworkIn   uint64  `json:"network_in"`
	NetworkOut  uint64  `json:"network_out"`
	Uptime      uint64  `json:"uptime"`
	Load        float64 `json:"load"`
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	HDDTotal    uint64  `json:"hdd_total"`
	HDDUsed     uint64  `json:"hdd_used"`
}

var vnstat string

func main() {
	// 输出连接
	fmt.Println("Connecting...")
	// 获取程序执行路径
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if nil != err {
		log.Fatalf("Can not find config file: %s\n", err.Error())
	}
	// 读取配置文件
	cfg, err := goconfig.LoadConfigFile(dir + "/status.ini")
	if nil != err {
		log.Fatalf("Can not load config file: %s\n", err.Error())
	}
	// 获取服务器地址
	server, err := cfg.GetValue("Status", "SERVER")
	if nil != err {
		log.Fatalf("Can not get server value: %s\n", err.Error())
	}
	// 获取端口
	port, err := cfg.GetValue("Status", "PORT")
	if nil != err {
		log.Fatalf("Can not get port value: %s\n", err.Error())
	}
	// 获取用户名
	user, err := cfg.GetValue("Status", "USER")
	if nil != err {
		log.Fatalf("Can not get user value: %s\n", err.Error())
	}
	// 获取密码
	pass, err := cfg.GetValue("Status", "PASSWORD")
	if nil != err {
		log.Fatalf("Can not get password value: %s\n", err.Error())
	}
	// 获取间隔
	interval, err := cfg.GetValue("Status", "INTERVAL")
	if nil != err {
		interval = "1"
	}
	// 获取vnstat执行路径
	// vnstat, err = cfg.GetValue("Status", "VNSTAT")
	// if nil != err {
	// 	log.Fatalf("Can not get vnstat exec path: %s\n", err.Error())
	// }
	// 组合连接地址
	addr := server + ":" + port
	for {
		// 连接服务器
		conn, err := net.Dial("tcp", addr)
		if nil != err {
			log.Printf("Can not connection server: %s\n", err.Error())
			// 关闭连接
			conn.Close()
			time.Sleep(time.Second * 3)
			continue
		}
		// 定义数据读取变量
		buf := make([]byte, 1024)
		// 读取1024字节数据
		_, err = conn.Read(buf)
		if nil != err && err != io.EOF {
			log.Printf("Read server data faild: %s\n", err.Error())
			// 关闭连接
			conn.Close()
			time.Sleep(time.Second * 3)
			continue
		}
		// 是否为验证提示
		if strings.Contains(string(buf), "Authentication required") {
			// 写入验证
			_, err := conn.Write([]byte(user + ":" + pass + "\n"))
			if nil != err {
				log.Printf("Send authentication required faild: %s\n", err.Error())
				// 关闭连接
				conn.Close()
				time.Sleep(time.Second * 3)
				continue
			}
			// 读取数据变量
			buf := make([]byte, 1024)
			// 读取1024字节数据
			_, err = conn.Read(buf)
			if nil != err && err != io.EOF {
				log.Printf("Read authentication required faild: %s\n", err.Error())
				// 关闭连接
				conn.Close()
				time.Sleep(time.Second * 3)
				continue
			}
			// 是否成功
			if !strings.Contains(string(buf), "Authentication successful") {
				log.Printf("Authentication required faild: %s\n", string(buf))
				// 关闭连接
				conn.Close()
				time.Sleep(time.Second * 3)
				continue
			}
		}
		// 直接死循环
		for {
			// 获取CPU信息
			cpuPercent := getCpuPercent()
			// 获取实时网速
			rx, tx := getSpeed()
			// 获取流量
			nin, nout := getTraffic()
			// 获取启动时间
			uptime := getUptime()
			// 获取CPU负载
			cpuLoad := getLoad()
			// 获取内存信息
			mTotal, mUsed := getMemory()
			// 获取交换空间信息
			sTotal, sUsed := getSwap()
			// 获取磁盘信息
			dTotal, dUsed := getDisk()
			// 组合为结构体
			data := SentData{
				Uptime:      uptime,
				Load:        cpuLoad,
				MemoryTotal: mTotal,
				MemoryUsed:  mUsed,
				SwapTotal:   sTotal,
				SwapUsed:    sUsed,
				HDDTotal:    dTotal,
				HDDUsed:     dUsed,
				Cpu:         cpuPercent,
				NetworkRx:   rx,
				NetworkTx:   tx,
				NetworkIn:   nin,
				NetworkOut:  nout,
			}
			// 转换为Json数据
			js, err := json.Marshal(data)
			if nil != err {
				log.Printf("Can not convert to json: %s", err.Error())
				// 关闭连接
				conn.Close()
				time.Sleep(time.Second * 3)
				break
			}
			// 发送实时数据
			_, err = conn.Write([]byte("update " + string(js) + "\n"))
			if nil != err {
				log.Printf("Can not send data: %s", err.Error())
				// 关闭连接
				conn.Close()
				time.Sleep(time.Second * 3)
				break
			}
			// 转换间隔为时间间隔
			inter, err := time.ParseDuration(interval)
			if nil != err {
				inter = 1
			}
			// 间隔
			time.Sleep(time.Second * inter)
		}
	}
}

// 获取内存信息
func getMemory() (uint64, uint64) {
	Mem, err := mem.VirtualMemory()
	if nil != err {
		return 0, 0
	}
	return Mem.Total / 1024, Mem.Used / 1024
}

// 获取启动时间
func getUptime() uint64 {
	up, err := host.BootTime()
	if nil != err {
		return 0
	}
	return uint64(time.Now().Unix()) - up
}

// 获取交换空间信息
func getSwap() (uint64, uint64) {
	Swap, err := mem.SwapMemory()
	if nil != err {
		return 0, 0
	}
	return Swap.Total / 1024, Swap.Used / 1024
}

// 获取硬盘信息
func getDisk() (uint64, uint64) {
	valid_fs := []string{"ext4", "ext3", "ext2", "reiserfs", "jfs", "btrfs", "fuseblk", "zfs", "simfs", "ntfs", "fat32", "exfat","xfs"}
	sort.Strings(valid_fs)
	// 获取所有分区
	ds, err := disk.Partitions(true)
	if nil != err {
		return 0, 0
	}
	// 总空间及使用空间变量
	var Total, Used uint64
	// 循环所有分区
	for _, d := range ds {
		//data, _ := json.MarshalIndent(d, "", "  ")
		index := sort.SearchStrings(valid_fs, d.Fstype)
		if index >= len(valid_fs) && valid_fs[index] != d.Fstype { //需要注意此处的判断，先判断 &&左侧的条件，如果不满足则结束此处判断，不会再进行右侧的判断
			continue
		}
		//log.Printf("type", string(data), valid_fs[0])
		// 读取分区使用情况
		Disk, err := disk.Usage(d.Mountpoint)
		if nil != err {
			continue
		}
		// 加上空间总量
		Total += Disk.Total
		// 加上使用总量
		Used += Disk.Used
	}
	return Total / 1024 / 1024, Used / 1024 / 1024
}

// 获取负载信息
func getLoad() float64 {
	Load, err := load.Avg()
	if nil != err {
		return -1.0
	}
	return Load.Load1
}

// 获取CPU使用率
func getCpuPercent() float64 {
	Cpu, err := cpu.Percent(time.Second, false)
	if nil != err {
		return 0.0
	}
	return Cpu[0]
}

// 获取实时网速
func getSpeed() (uint64, uint64) {
	// 读取所有网卡网速
	Net, err := net2.IOCounters(true)
	if nil != err {
		return 0, 0
	}
	// 定义网速保存变量
	var rx, tx uint64
	// 循环网络信息
	for _, nv := range Net {
		// 去除多余信息
		if "lo" == nv.Name || strings.Contains(nv.Name, "tun") {
			continue
		}
		// 加上网速信息
		rx += nv.BytesRecv
		tx += nv.BytesSent
	}
	// 暂停一秒
	time.Sleep(time.Second)
	// 重新读取网络信息
	Net, err = net2.IOCounters(true)
	if nil != err {
		return 0, 0
	}
	// 网络信息保存变量
	var rx2, tx2 uint64
	// 循环网络信息
	for _, nv := range Net {
		// 去除多余信息
		if "lo" == nv.Name || strings.Contains(nv.Name, "tun") {
			continue
		}
		// 加上网速信息
		rx2 += nv.BytesRecv
		tx2 += nv.BytesSent
	}
	// 两次相见
	return rx2 - rx, tx2 - tx
}

func getTraffic() (uint64, uint64) {
	file, err := os.Open("/proc/net/dev")
   	if err != nil {
    	panic(err)
   	}
	defer file.Close()
   	content, err := ioutil.ReadAll(file)
    buf := string(content)
    //解析正则表达式，如果成功返回解释器
    reg1 := regexp.MustCompile(`([^\s]+):[\s]{0,}(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)`)
    if reg1 == nil { //解释失败，返回nil
        fmt.Println("regexp err")
        return 0, 0
    }
    //根据规则提取关键信息
	result1 := reg1.FindAllStringSubmatch(buf, -1)
    // fmt.Println("result1 = ", result1[0][1])
    // fmt.Println("result1 = ", result1[0][8])
    for _,r := range result1{
        if (r[1] == "lo" || strings.Contains(r[1], `tun`) || strings.Contains(r[1], `docker`) ||
                 strings.Contains(r[1], `veth`) || strings.Contains(r[1], `br-`) ||
                 strings.Contains(r[1], `vmbr`) || strings.Contains(r[1], `vnet`) ||
                 strings.Contains(r[1], `kube`)){
            continue;
        } else {
			in, _ := strconv.ParseUint(r[2], 10, 64)
			out, _ := strconv.ParseUint(r[10], 10, 64)
			return in, out
        }
    }
	return 0, 0
}

