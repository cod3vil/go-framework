package service

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// serverStart 进程启动时间，用于计算运行时长。
var serverStart = time.Now()

// ServerStat 服务监控指标。
type ServerStat struct {
	Host    HostStat    `json:"host"`
	CPU     CPUStat     `json:"cpu"`
	Memory  MemoryStat  `json:"memory"`
	Disk    []DiskStat  `json:"disk"`
	Runtime RuntimeStat `json:"runtime"`
}

// HostStat 主机信息。
type HostStat struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
	Arch     string `json:"arch"`
	BootTime string `json:"bootTime"`
}

// CPUStat CPU 信息。
type CPUStat struct {
	Cores       int     `json:"cores"`
	UsedPercent float64 `json:"usedPercent"`
}

// MemoryStat 物理内存信息（字节）。
type MemoryStat struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"usedPercent"`
}

// DiskStat 磁盘分区信息（字节）。
type DiskStat struct {
	Path        string  `json:"path"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

// RuntimeStat Go 运行时信息。
type RuntimeStat struct {
	GoVersion  string `json:"goVersion"`
	Goroutines int    `json:"goroutines"`
	NumGC      uint32 `json:"numGC"`
	AllocMB    uint64 `json:"allocMB"`
	SysMB      uint64 `json:"sysMB"`
	Uptime     string `json:"uptime"`
}

// ServerStat 采集服务运行指标。个别子项采集失败时以零值返回，不影响整体。
func (s *Service) ServerStat(ctx context.Context) *ServerStat {
	stat := &ServerStat{Runtime: runtimeStat()}

	if info, err := host.InfoWithContext(ctx); err == nil {
		stat.Host = HostStat{
			Hostname: info.Hostname, OS: info.OS, Platform: info.Platform,
			Arch:     info.KernelArch,
			BootTime: time.Unix(int64(info.BootTime), 0).Format(time.RFC3339),
		}
	}

	stat.CPU.Cores = runtime.NumCPU()
	// 采样 200ms 得到整体使用率。
	if percents, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, false); err == nil && len(percents) > 0 {
		stat.CPU.UsedPercent = round2(percents[0])
	}

	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		stat.Memory = MemoryStat{
			Total: vm.Total, Used: vm.Used, Available: vm.Available,
			UsedPercent: round2(vm.UsedPercent),
		}
	}

	if parts, err := disk.PartitionsWithContext(ctx, false); err == nil {
		seen := make(map[string]bool)
		for _, p := range parts {
			if seen[p.Mountpoint] {
				continue
			}
			seen[p.Mountpoint] = true
			if u, err := disk.UsageWithContext(ctx, p.Mountpoint); err == nil {
				stat.Disk = append(stat.Disk, DiskStat{
					Path: u.Path, Total: u.Total, Used: u.Used,
					UsedPercent: round2(u.UsedPercent),
				})
			}
		}
	}
	return stat
}

func runtimeStat() RuntimeStat {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return RuntimeStat{
		GoVersion:  runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		NumGC:      m.NumGC,
		AllocMB:    m.Alloc / 1024 / 1024,
		SysMB:      m.Sys / 1024 / 1024,
		Uptime:     time.Since(serverStart).Round(time.Second).String(),
	}
}

func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100
}
