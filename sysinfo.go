package main

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// 系统信息结构体
type SysInfo struct {
	CPUUsage    float64 `json:"cpu_usage_per"`
	TotalMem    uint64  `json:"total_mem"`
	UsedMem     uint64  `json:"used_mem"`
	FreeMem     uint64  `json:"free_mem"`
	UsedPercent float64 `json:"used_mem_per"`
}

// 获取系统信息
func getSysInfo() (*SysInfo, error) {
	// cpu百分比
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	// 内存信息
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	// 构造返回数据
	sysInfo := &SysInfo{
		CPUUsage:    cpuPercent[0],
		TotalMem:    vMem.Total,
		UsedMem:     vMem.Used,
		FreeMem:     vMem.Free,
		UsedPercent: vMem.UsedPercent,
	}
	return sysInfo, nil
}

// 获取系统信息的 API 处理函数
