package engine

import (
	"errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"time"
)

type MachineStats struct {
	TotalMemory  uint64  `json:"totalMemory"`
	UsedMemory   uint64  `json:"usedMemory"`
	CpuLoad      float32 `json:"cpuLoad"`
	BytesSent    uint64  `json:"bytesSent"`
	ByteReceived uint64  `json:"byteReceived"`
}

func getMemory(result *MachineStats) error {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	result.UsedMemory = vmem.Used
	result.TotalMemory = vmem.Total
	return nil
}

func getCpu(result *MachineStats) error {
	m, _ := time.ParseDuration("1s")
	loads, err := cpu.Percent(m, false)
	if err != nil {
		return err
	}
	if len(loads) == 0 {
		return errors.New("Failed to obtain CPU loads")
	}
	result.CpuLoad = float32(loads[0])
	return nil
}

func getNetworkStats(result *MachineStats) error {
	interfaces, err := net.IOCounters(false)
	if err != nil {
		return err
	}
	if len(interfaces) == 0 {
		return errors.New("Failed to obtain NIC loads")
	}
	result.ByteReceived = interfaces[0].BytesRecv
	result.BytesSent = interfaces[0].BytesSent

	return nil
}



func GetMachineStats() (*MachineStats, error) {
	var result MachineStats
	if err := getMemory(&result); err != nil {
		return nil, err
	}
	if err := getCpu(&result); err != nil {
		return nil, err
	}
	if err := getNetworkStats(&result); err != nil {
		return nil, err
	}
	return &result, nil
}