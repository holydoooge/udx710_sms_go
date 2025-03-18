package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type SysInfo struct {
	CPUAndMemory *CPUAndMemory `json:"cpu_mem"`
	Temperature  *Temperature  `json:"temp"`
	Battery      *BatteryInfo  `json:"battery"`
}

type CPUAndMemory struct {
	CPUUsage    float64 `json:"cpu_usage_per"`
	TotalMem    uint64  `json:"total_mem"`
	UsedMem     uint64  `json:"used_mem"`
	FreeMem     uint64  `json:"free_mem"`
	UsedPercent float64 `json:"used_mem_per"`
}

type Temperature map[string]string
type BatteryInfo map[string]string

func getCPUAndMemory() (*CPUAndMemory, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	CPUAndMemory := &CPUAndMemory{
		CPUUsage:    cpuPercent[0],
		TotalMem:    vMem.Total,
		UsedMem:     vMem.Used,
		FreeMem:     vMem.Free,
		UsedPercent: vMem.UsedPercent,
	}
	return CPUAndMemory, nil
}

func getTemperature() (Temperature, error) {
	temps := make(Temperature)
	typeFiles, err := filepath.Glob("/sys/class/thermal/thermal_zone*/type")
	if err != nil {
		return nil, err
	}

	for _, typeFile := range typeFiles {
		typeData, err := os.ReadFile(typeFile)
		if err != nil {
			return nil, err
		}
		typeName := strings.TrimSpace(string(typeData))
		tempFile := filepath.Join(filepath.Dir(typeFile), "temp")
		tempData, err := os.ReadFile(tempFile)
		if err != nil {
			return nil, err
		}
		temps[typeName] = strings.TrimSpace(string(tempData))
	}
	return temps, nil
}

func getBatteryInfo() (BatteryInfo, error) {
	batteryInfo := make(BatteryInfo)
	ueventPath := "/sys/devices/platform/charger-manager/power_supply/battery/uevent"
	if _, err := os.Stat(ueventPath); os.IsNotExist(err) {
		return batteryInfo, nil
	}

	content, err := os.ReadFile(ueventPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			batteryInfo[parts[0]] = parts[1]
		}
	}
	return batteryInfo, nil
}

func getSysInfo() (*SysInfo, error) {
	cpuAndMemory, err := getCPUAndMemory()
	if err != nil {
		return nil, err
	}

	temperature, err := getTemperature()
	if err != nil {
		return nil, err
	}

	batteryInfo, err := getBatteryInfo()
	if err != nil {
		return nil, err
	}

	sysInfo := &SysInfo{
		CPUAndMemory: cpuAndMemory,
		Temperature:  &temperature,
		Battery:      &batteryInfo,
	}
	return sysInfo, nil
}
