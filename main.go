package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type Security struct {
	Authorization string `header:"Authorization"`
}

func (sec *Security) IsValid(value string) bool {
	return sec.Authorization == value
}

type Proc struct {
	Pid        int32    `json:"pid"`
	Ppid       int32    `json:"ppid"`
	Uids       []int32  `json:"uids"`
	Gids       []int32  `json:"gids"`
	Groups     []int32  `json:"groups"`
	Nthreads   int32    `json:"nthreads"`
	User       string   `json:"user"`
	Nice       int32    `json:"nice"`
	Name       string   `json:"name"`
	Executable string   `json:"executable"`
	Cmdline    []string `json:"cmdline"`
	Cwd        string   `json:"cwd"`
	CpuPerc    float64  `json:"cpuperc"`
	MemPerc    float32  `json:"memperc"`
	CreateTime int64    `json:"createtime"`
	IsRunning  bool     `json:"isrunning"`
}

type Storage struct {
	Partition disk.PartitionStat `json:"partition"`
	Usage     *disk.UsageStat    `json:"usage"`
}

type System struct {
	Procs   []Proc                 `json:"procs"`
	Storage []Storage              `json:"storage"`
	SysLoad *load.AvgStat          `json:"sysload"`
	MemLoad *mem.VirtualMemoryStat `json:"memload"`
	Host    *host.InfoStat         `json:"host"`
	Net     []net.InterfaceStat    `json:"net"`
}

var (
	apikey string
)

func GetSystemProcs() []Proc {
	processes, _ := process.Processes()
	array := make([]Proc, len(processes))

	for i, process := range processes {
		proc := Proc{}
		proc.Pid = process.Pid
		proc.Ppid, _ = process.Ppid()
		proc.Uids, _ = process.Uids()
		proc.Gids, _ = process.Gids()
		proc.Groups, _ = process.Groups()
		proc.Nthreads, _ = process.NumThreads()
		proc.User, _ = process.Username()
		proc.Nice, _ = process.Nice()
		proc.Name, _ = process.Name()
		proc.Executable, _ = process.Exe()
		proc.Cmdline, _ = process.CmdlineSlice()
		proc.Cwd, _ = process.Cwd()
		proc.CpuPerc, _ = process.CPUPercent()
		proc.MemPerc, _ = process.MemoryPercent()
		proc.CreateTime, _ = process.CreateTime()
		proc.IsRunning, _ = process.IsRunning()
		array[i] = proc
	}

	return array
}

func GetStorage() []Storage {
	parts, _ := disk.Partitions(false)
	array := make([]Storage, len(parts))

	for i, part := range parts {
		storage := Storage{}
		storage.Partition = part
		storage.Usage, _ = disk.Usage(part.Mountpoint)
		array[i] = storage
	}

	return array
}

func GetSystemInfo(c *gin.Context) {
	var sec Security

	if err := c.BindHeader(&sec); err != nil || !sec.IsValid(apikey) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	sys := System{}
	sys.Procs = GetSystemProcs()
	sys.SysLoad, _ = load.Avg()
	sys.MemLoad, _ = mem.VirtualMemory()
	sys.Storage = GetStorage()
	sys.Net, _ = net.Interfaces()
	sys.Host, _ = host.Info()
	c.JSON(http.StatusOK, &sys)
}

func main() {
	var bind, path string
	var debug, uuidgen bool
	flag.StringVar(&path, "path", "/", "web path")
	flag.StringVar(&bind, "bind", ":8080", "bind address")
	flag.StringVar(&apikey, "apikey", "", "sets the api key")
	flag.BoolVar(&uuidgen, "uuidgen", false, "generate a new uuid and exit")
	flag.BoolVar(&debug, "debug", false, "enable debug messages")
	flag.Parse()

	if uuidgen {
		fmt.Println(uuid.NewString())
		return
	}

	gin.DisableConsoleColor()

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if len(apikey) < 1 {
		panic("-apikey was not set")
		return
	}

	engine := gin.Default()
	engine.GET(path, GetSystemInfo)
	engine.Run(bind)
}
