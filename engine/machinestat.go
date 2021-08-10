package engine

import (
	"errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"strings"
	"time"
)

type PartitionUsage struct {
	Mount string `json:"mount"`
	FsType string `json:"fs_type"`
	Device string `json:"device"`
	Total uint64 `json:"total"`
	TotalGB uint64 `json:"total_gb"`
	Free uint64 `json:"free"`
	FreeGB uint64 `json:"free_gb"`
}

type DiskUsage struct {
	TotalGB uint64 `json:"total_gb"`
	Usage [] PartitionUsage `json:"usage"`
}

type MachineStats struct {
	TotalMemory  uint64  `json:"totalMemory"`
	UsedMemory   uint64  `json:"usedMemory"`
	CpuLoad      float32 `json:"cpuLoad"`
	CoresNumber  int     `json:"cores_number"`
	BytesSent    uint64  `json:"bytesSent"`
	ByteReceived uint64  `json:"byteReceived"`
	Uptime       uint64  `json:"uptime"`
	Disk 		 DiskUsage `json:"disk"`
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
	info, err2 := cpu.Info()
	if err2 != nil {
		return err2
	}
	result.CoresNumber = len(info)
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

func getDiskUsage(result *MachineStats) error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}
	var du DiskUsage
	du.TotalGB = 0
	for _, v := range partitions{
		if strings.HasPrefix(v.Device, "/dev/loop") {
			continue
		}
		usage, err := disk.Usage(v.Mountpoint)
		if err != nil {
			return err
		}

		var pu PartitionUsage
		pu.FsType = v.Fstype
		pu.Mount = v.Mountpoint
		pu.Device = v.Device
		pu.Free = usage.Free
		pu.FreeGB = pu.Free / (1024*1024*1024)
		pu.Total = usage.Total
		pu.TotalGB = pu.Total / (1024*1024*1024)
		du.TotalGB += pu.TotalGB
		du.Usage = append(du.Usage, pu)
	}
	result.Disk = du
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

func getUptime(result *MachineStats) error {
	uptime, err := host.Uptime()
	if err != nil {
		return err
	}
	result.Uptime = uptime
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
	if err := getUptime(&result); err != nil {
		return nil, err
	}
	if err := getDiskUsage(&result); err != nil {
		return nil, err
	}
	return &result, nil

}