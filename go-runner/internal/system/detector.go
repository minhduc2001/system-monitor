package system

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Detector handles system information collection
type Detector struct {
	startTime time.Time
}

// NewDetector creates a new system detector
func NewDetector() *Detector {
	return &Detector{
		startTime: time.Now(),
	}
}

// GetSystemInfo collects comprehensive system information
func (d *Detector) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		Timestamp: time.Now(),
		Uptime:    time.Since(d.startTime),
	}

	// Get host information
	if err := d.getHostInfo(info); err != nil {
		return nil, fmt.Errorf("failed to get host info: %v", err)
	}

	// Get CPU information
	if err := d.getCPUInfo(info); err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %v", err)
	}

	// Get memory information
	if err := d.getMemoryInfo(info); err != nil {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}

	// Get disk information
	if err := d.getDiskInfo(info); err != nil {
		return nil, fmt.Errorf("failed to get disk info: %v", err)
	}

	// Get network information
	if err := d.getNetworkInfo(info); err != nil {
		return nil, fmt.Errorf("failed to get network info: %v", err)
	}

	// Get process information
	if err := d.getProcessInfo(info); err != nil {
		return nil, fmt.Errorf("failed to get process info: %v", err)
	}

	return info, nil
}

// getHostInfo collects host information
func (d *Detector) getHostInfo(info *SystemInfo) error {
	hostInfo, err := host.Info()
	if err != nil {
		return err
	}

	info.Hostname = hostInfo.Hostname
	info.Platform = hostInfo.Platform
	info.Architecture = hostInfo.KernelArch
	info.GoVersion = runtime.Version()

	return nil
}

// getCPUInfo collects CPU information
func (d *Detector) getCPUInfo(info *SystemInfo) error {
	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}

	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return err
	}

	// Get load average
	loadAvg, err := load.Avg()
	if err != nil {
		// Load average might not be available on all systems
		loadAvg = &load.AvgStat{}
	}

	// Calculate average CPU usage
	var avgUsage float64
	if len(cpuPercent) > 0 {
		for _, usage := range cpuPercent {
			avgUsage += usage
		}
		avgUsage = avgUsage / float64(len(cpuPercent))
	}

	info.CPU = CPUInfo{
		Usage:       avgUsage,
		Count:       len(cpuInfo),
		ModelName:   cpuInfo[0].ModelName,
		Mhz:         cpuInfo[0].Mhz,
		LoadAverage: []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15},
	}

	return nil
}

// getMemoryInfo collects memory information
func (d *Detector) getMemoryInfo(info *SystemInfo) error {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	swapInfo, err := mem.SwapMemory()
	if err != nil {
		// Swap might not be available on all systems
		swapInfo = &mem.SwapMemoryStat{}
	}

	info.Memory = MemoryInfo{
		Total:     memInfo.Total,
		Available: memInfo.Available,
		Used:      memInfo.Used,
		Free:      memInfo.Free,
		Usage:     memInfo.UsedPercent,
		SwapTotal: swapInfo.Total,
		SwapUsed:  swapInfo.Used,
		SwapFree:  swapInfo.Free,
		SwapUsage: swapInfo.UsedPercent,
	}

	return nil
}

// getDiskInfo collects disk information
func (d *Detector) getDiskInfo(info *SystemInfo) error {
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return err
	}

	// Get inode information
	inodeInfo, err := disk.Usage("/")
	if err != nil {
		// Inode info might not be available on all systems
		inodeInfo = &disk.UsageStat{}
	}

	info.Disk = DiskInfo{
		Total:       diskInfo.Total,
		Used:        diskInfo.Used,
		Free:        diskInfo.Free,
		Usage:       diskInfo.UsedPercent,
		InodesTotal: inodeInfo.InodesTotal,
		InodesUsed:  inodeInfo.InodesUsed,
		InodesFree:  inodeInfo.InodesFree,
		InodesUsage: inodeInfo.InodesUsedPercent,
	}

	return nil
}

// getNetworkInfo collects network information
func (d *Detector) getNetworkInfo(info *SystemInfo) error {
	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	// Get network I/O statistics
	netIO, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	// Convert interfaces to our format
	var networkInterfaces []NetworkInterface
	var totalBytesSent, totalBytesReceived uint64
	var totalPacketsSent, totalPacketsReceived uint64

	for _, iface := range interfaces {
		// Find corresponding I/O stats
		var ioStats *net.IOCountersStat
		for _, io := range netIO {
			if io.Name == iface.Name {
				ioStats = &io
				break
			}
		}

		if ioStats == nil {
			continue
		}

		// Convert addresses to strings
		var addrs []string
		for _, addr := range iface.Addrs {
			addrs = append(addrs, addr.Addr)
		}

		networkInterface := NetworkInterface{
			Name:             iface.Name,
			MTU:              iface.MTU,
			HardwareAddr:     iface.HardwareAddr,
			Flags:            fmt.Sprintf("%v", iface.Flags),
			Addrs:            addrs,
			BytesSent:        ioStats.BytesSent,
			BytesReceived:    ioStats.BytesRecv,
			PacketsSent:      ioStats.PacketsSent,
			PacketsReceived:  ioStats.PacketsRecv,
			ErrorsIn:         ioStats.Errin,
			ErrorsOut:        ioStats.Errout,
			DropIn:           ioStats.Dropin,
			DropOut:          ioStats.Dropout,
		}

		networkInterfaces = append(networkInterfaces, networkInterface)

		totalBytesSent += ioStats.BytesSent
		totalBytesReceived += ioStats.BytesRecv
		totalPacketsSent += ioStats.PacketsSent
		totalPacketsReceived += ioStats.PacketsRecv
	}

	info.Network = NetworkInfo{
		Interfaces:           networkInterfaces,
		TotalBytesSent:       totalBytesSent,
		TotalBytesReceived:   totalBytesReceived,
		TotalPacketsSent:     totalPacketsSent,
		TotalPacketsReceived: totalPacketsReceived,
	}

	return nil
}

// getProcessInfo collects process information
func (d *Detector) getProcessInfo(info *SystemInfo) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	var processInfos []ProcessInfo
	for _, p := range processes {
		// Get process info
		name, _ := p.Name()
		status, _ := p.Status()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		memInfo, _ := p.MemoryInfo()
		createTime, _ := p.CreateTime()
		username, _ := p.Username()
		cmdline, _ := p.Cmdline()

		processInfo := ProcessInfo{
			PID:          p.Pid,
			Name:         name,
			Status:       fmt.Sprintf("%v", status),
			CPUPercent:   cpuPercent,
			MemoryPercent: float64(memPercent),
			CreateTime:   createTime,
			Username:     username,
			Command:      cmdline,
		}

		if memInfo != nil {
			processInfo.MemoryRSS = memInfo.RSS
			processInfo.MemoryVMS = memInfo.VMS
		}

		processInfos = append(processInfos, processInfo)
	}

	info.Processes = processInfos
	return nil
}

// GetSystemStatus returns current system status
func (d *Detector) GetSystemStatus() (*SystemStatus, error) {
	info, err := d.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	status := &SystemStatus{
		LastCheck:   time.Now(),
		Uptime:      info.Uptime,
		ActiveAlerts: 0,
	}

	// Determine overall status
	status.CPUStatus = d.getCPUStatus(info.CPU.Usage)
	status.MemoryStatus = d.getMemoryStatus(info.Memory.Usage)
	status.DiskStatus = d.getDiskStatus(info.Disk.Usage)

	// Determine overall system status
	if status.CPUStatus == "critical" || status.MemoryStatus == "critical" || status.DiskStatus == "critical" {
		status.Status = "critical"
		status.Message = "System is in critical state"
	} else if status.CPUStatus == "warning" || status.MemoryStatus == "warning" || status.DiskStatus == "warning" {
		status.Status = "warning"
		status.Message = "System is in warning state"
	} else {
		status.Status = "healthy"
		status.Message = "System is healthy"
	}

	return status, nil
}

// getCPUStatus determines CPU status based on usage
func (d *Detector) getCPUStatus(usage float64) string {
	if usage >= 90 {
		return "critical"
	} else if usage >= 70 {
		return "warning"
	}
	return "healthy"
}

// getMemoryStatus determines memory status based on usage
func (d *Detector) getMemoryStatus(usage float64) string {
	if usage >= 90 {
		return "critical"
	} else if usage >= 80 {
		return "warning"
	}
	return "healthy"
}

// getDiskStatus determines disk status based on usage
func (d *Detector) getDiskStatus(usage float64) string {
	if usage >= 95 {
		return "critical"
	} else if usage >= 85 {
		return "warning"
	}
	return "healthy"
}
