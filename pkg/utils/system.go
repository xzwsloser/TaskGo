package utils

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
)

/*
	@Description: System Info Operation
*/

const (
	B  = 1
	KB = 1024 * B 
	MB = 1024 * KB
	GB = 1024 * MB
)

// Handware
type Os struct {
	GOOS			string	`json:"goos"`
	NumCPU			int		`json:"numCpu"`
	Compiler		string  `json:"compiler"`
	GoVersion		string	`json:"goVersion"`
	NumGoroutine	int		`json:"numGoroutine"`
}

type Cpu struct {
	Cpus	[]float64	`json:"cpus"`
	Cores	int			`json:"cores"`
}

type Ram struct {
	UsedMB		int		 `json:"usedMb"`
	TotalMB		int		 `json:"totalMb"`
	UsedPercent int 	 `json:"usedPercent"`
}

type Disk struct {
	UsedMB		int		`json:"usedMb"`
	UsedGB		int		`json:"usedGb"`
	TotalMB		int		`json:"totalMb"`
	TotalGB		int		`json:"totalGb"`
	UsedPercent int 	`json:"usedPercent"`
}

// System Info
type SystemInfo struct {
	Os		Os		`json:"os"`
	Cpu		Cpu		`json:"cpu"`
	Ram		Ram		`json:"ram"`
	Disk	Disk	`json:"disk"`
}

func InitOS() (os Os) {
	os.GOOS   = runtime.GOOS
	os.NumCPU =	runtime.NumCPU() 
	os.Compiler = runtime.Compiler
	os.GoVersion = runtime.Version()
	os.NumGoroutine = runtime.NumGoroutine()
	return 
}

func InitCPU() (c Cpu, err error) {
	if c.Cores, err = cpu.Counts(false) ; err != nil {
		return 
	}

	if c.Cpus, err = cpu.Percent(time.Duration(200)*time.Millisecond, true); err != nil {
		return
	}

	return
}

func InitRAM() (ram Ram, err error) {
	if u, err := mem.VirtualMemory(); err != nil {
		return ram, err
	} else {
		ram.UsedMB = int(u.Used) / MB
		ram.TotalMB = int(u.Total) / MB
		ram.UsedPercent = int(u.UsedPercent)
	} 

	return 
}

func InitDisk() (d Disk, err error) {
	if u, err := disk.Usage("/"); err != nil {
		return d, err
	} else {
		d.UsedMB = int(u.Used) / MB
		d.UsedGB = int(u.Used) / GB
		d.TotalMB = int(u.Total) / MB
		d.TotalGB = int(u.Total) / GB
		d.UsedPercent = int(u.UsedPercent) 
	}

	return 
}

func GetSystemInfo() (sf *SystemInfo, err error) {
	var info SystemInfo
	info.Os = InitOS()
	if info.Cpu, err = InitCPU(); err != nil {
		return &info, err
	}

	if info.Ram, err = InitRAM(); err != nil {
		return &info, err
	}

	if info.Disk, err = InitDisk(); err != nil {
		return &info, err
	}

	return &info, nil
}



