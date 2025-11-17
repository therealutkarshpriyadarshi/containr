package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// ContainerMetrics holds comprehensive metrics for a container
type ContainerMetrics struct {
	ContainerID string    `json:"container_id"`
	Timestamp   time.Time `json:"timestamp"`

	// CPU metrics
	CPUStats CPUStats `json:"cpu_stats"`

	// Memory metrics
	MemoryStats MemoryStats `json:"memory_stats"`

	// Network metrics
	NetworkStats NetworkStats `json:"network_stats"`

	// Disk I/O metrics
	DiskStats DiskStats `json:"disk_stats"`

	// PID metrics
	PIDStats PIDStats `json:"pid_stats"`
}

// CPUStats holds CPU usage statistics
type CPUStats struct {
	UsageNanos     uint64  `json:"usage_nanos"`     // Total CPU time in nanoseconds
	SystemNanos    uint64  `json:"system_nanos"`    // System CPU time
	UserNanos      uint64  `json:"user_nanos"`      // User CPU time
	PercentCPU     float64 `json:"percent_cpu"`     // CPU usage percentage
	ThrottledTime  uint64  `json:"throttled_time"`  // Time throttled (nanoseconds)
	ThrottledCount uint64  `json:"throttled_count"` // Number of throttle periods
}

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	Usage           uint64  `json:"usage"`             // Current memory usage (bytes)
	MaxUsage        uint64  `json:"max_usage"`         // Maximum memory usage (bytes)
	Limit           uint64  `json:"limit"`             // Memory limit (bytes)
	Cache           uint64  `json:"cache"`             // Page cache memory (bytes)
	RSS             uint64  `json:"rss"`               // Resident set size (bytes)
	Swap            uint64  `json:"swap"`              // Swap usage (bytes)
	PercentUsed     float64 `json:"percent_used"`      // Percentage of limit used
	OOMKillDisabled bool    `json:"oom_kill_disabled"` // OOM killer disabled
	OOMKillCount    uint64  `json:"oom_kill_count"`    // Number of OOM kills
}

// NetworkStats holds network I/O statistics
type NetworkStats struct {
	RxBytes   uint64 `json:"rx_bytes"`   // Received bytes
	RxPackets uint64 `json:"rx_packets"` // Received packets
	RxErrors  uint64 `json:"rx_errors"`  // Receive errors
	RxDropped uint64 `json:"rx_dropped"` // Received packets dropped
	TxBytes   uint64 `json:"tx_bytes"`   // Transmitted bytes
	TxPackets uint64 `json:"tx_packets"` // Transmitted packets
	TxErrors  uint64 `json:"tx_errors"`  // Transmit errors
	TxDropped uint64 `json:"tx_dropped"` // Transmitted packets dropped
}

// DiskStats holds disk I/O statistics
type DiskStats struct {
	ReadBytes  uint64 `json:"read_bytes"`  // Bytes read
	WriteBytes uint64 `json:"write_bytes"` // Bytes written
	ReadOps    uint64 `json:"read_ops"`    // Read operations
	WriteOps   uint64 `json:"write_ops"`   // Write operations
}

// PIDStats holds PID statistics
type PIDStats struct {
	Current uint64 `json:"current"` // Current number of PIDs
	Limit   uint64 `json:"limit"`   // PID limit
}

// MetricsCollector collects container metrics
type MetricsCollector struct {
	cgroupPath string
	pid        int
	log        *logger.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(cgroupPath string, pid int) *MetricsCollector {
	return &MetricsCollector{
		cgroupPath: cgroupPath,
		pid:        pid,
		log:        logger.New("metrics"),
	}
}

// Collect collects all metrics for a container
func (mc *MetricsCollector) Collect(containerID string) (*ContainerMetrics, error) {
	metrics := &ContainerMetrics{
		ContainerID: containerID,
		Timestamp:   time.Now(),
	}

	// Collect CPU stats
	if cpuStats, err := mc.collectCPUStats(); err == nil {
		metrics.CPUStats = cpuStats
	} else {
		mc.log.Debugf("Failed to collect CPU stats: %v", err)
	}

	// Collect memory stats
	if memStats, err := mc.collectMemoryStats(); err == nil {
		metrics.MemoryStats = memStats
	} else {
		mc.log.Debugf("Failed to collect memory stats: %v", err)
	}

	// Collect network stats
	if netStats, err := mc.collectNetworkStats(); err == nil {
		metrics.NetworkStats = netStats
	} else {
		mc.log.Debugf("Failed to collect network stats: %v", err)
	}

	// Collect disk stats
	if diskStats, err := mc.collectDiskStats(); err == nil {
		metrics.DiskStats = diskStats
	} else {
		mc.log.Debugf("Failed to collect disk stats: %v", err)
	}

	// Collect PID stats
	if pidStats, err := mc.collectPIDStats(); err == nil {
		metrics.PIDStats = pidStats
	} else {
		mc.log.Debugf("Failed to collect PID stats: %v", err)
	}

	return metrics, nil
}

// collectCPUStats collects CPU usage statistics
func (mc *MetricsCollector) collectCPUStats() (CPUStats, error) {
	stats := CPUStats{}

	// Try cgroup v2 first
	cpuStatPath := filepath.Join(mc.cgroupPath, "cpu.stat")
	if data, err := os.ReadFile(cpuStatPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			value, _ := strconv.ParseUint(fields[1], 10, 64)
			switch fields[0] {
			case "usage_usec":
				stats.UsageNanos = value * 1000
			case "user_usec":
				stats.UserNanos = value * 1000
			case "system_usec":
				stats.SystemNanos = value * 1000
			case "nr_throttled":
				stats.ThrottledCount = value
			case "throttled_usec":
				stats.ThrottledTime = value * 1000
			}
		}
		return stats, nil
	}

	// Fall back to cgroup v1
	cpuacctUsage := filepath.Join(mc.cgroupPath, "cpuacct.usage")
	if data, err := os.ReadFile(cpuacctUsage); err == nil {
		if value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
			stats.UsageNanos = value
		}
	}

	return stats, nil
}

// collectMemoryStats collects memory usage statistics
func (mc *MetricsCollector) collectMemoryStats() (MemoryStats, error) {
	stats := MemoryStats{}

	// Try cgroup v2
	memStatPath := filepath.Join(mc.cgroupPath, "memory.stat")
	memCurrentPath := filepath.Join(mc.cgroupPath, "memory.current")
	memMaxPath := filepath.Join(mc.cgroupPath, "memory.max")

	// Read current usage
	if data, err := os.ReadFile(memCurrentPath); err == nil {
		if value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
			stats.Usage = value
		}
	}

	// Read limit
	if data, err := os.ReadFile(memMaxPath); err == nil {
		limitStr := strings.TrimSpace(string(data))
		if limitStr != "max" {
			if value, err := strconv.ParseUint(limitStr, 10, 64); err == nil {
				stats.Limit = value
			}
		}
	}

	// Read detailed stats
	if data, err := os.ReadFile(memStatPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			value, _ := strconv.ParseUint(fields[1], 10, 64)
			switch fields[0] {
			case "file":
				stats.Cache = value
			case "anon":
				stats.RSS = value
			case "swap":
				stats.Swap = value
			}
		}
	}

	// Calculate percentage
	if stats.Limit > 0 {
		stats.PercentUsed = float64(stats.Usage) / float64(stats.Limit) * 100
	}

	return stats, nil
}

// collectNetworkStats collects network I/O statistics
func (mc *MetricsCollector) collectNetworkStats() (NetworkStats, error) {
	stats := NetworkStats{}

	// Read from /proc/<pid>/net/dev
	netDevPath := fmt.Sprintf("/proc/%d/net/dev", mc.pid)
	data, err := os.ReadFile(netDevPath)
	if err != nil {
		return stats, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// Skip header lines
		if !strings.Contains(line, ":") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 16 {
			continue
		}

		// Skip loopback interface
		if strings.HasPrefix(fields[0], "lo:") {
			continue
		}

		// Parse RX stats
		rxBytes, _ := strconv.ParseUint(fields[1], 10, 64)
		rxPackets, _ := strconv.ParseUint(fields[2], 10, 64)
		rxErrors, _ := strconv.ParseUint(fields[3], 10, 64)
		rxDropped, _ := strconv.ParseUint(fields[4], 10, 64)

		// Parse TX stats
		txBytes, _ := strconv.ParseUint(fields[9], 10, 64)
		txPackets, _ := strconv.ParseUint(fields[10], 10, 64)
		txErrors, _ := strconv.ParseUint(fields[11], 10, 64)
		txDropped, _ := strconv.ParseUint(fields[12], 10, 64)

		// Accumulate stats from all interfaces
		stats.RxBytes += rxBytes
		stats.RxPackets += rxPackets
		stats.RxErrors += rxErrors
		stats.RxDropped += rxDropped
		stats.TxBytes += txBytes
		stats.TxPackets += txPackets
		stats.TxErrors += txErrors
		stats.TxDropped += txDropped
	}

	return stats, nil
}

// collectDiskStats collects disk I/O statistics
func (mc *MetricsCollector) collectDiskStats() (DiskStats, error) {
	stats := DiskStats{}

	// Try cgroup v2
	ioStatPath := filepath.Join(mc.cgroupPath, "io.stat")
	if data, err := os.ReadFile(ioStatPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			for _, field := range fields[1:] {
				kv := strings.Split(field, "=")
				if len(kv) != 2 {
					continue
				}

				value, _ := strconv.ParseUint(kv[1], 10, 64)
				switch kv[0] {
				case "rbytes":
					stats.ReadBytes += value
				case "wbytes":
					stats.WriteBytes += value
				case "rios":
					stats.ReadOps += value
				case "wios":
					stats.WriteOps += value
				}
			}
		}
	}

	return stats, nil
}

// collectPIDStats collects PID statistics
func (mc *MetricsCollector) collectPIDStats() (PIDStats, error) {
	stats := PIDStats{}

	// Try cgroup v2
	pidCurrentPath := filepath.Join(mc.cgroupPath, "pids.current")
	pidMaxPath := filepath.Join(mc.cgroupPath, "pids.max")

	// Read current PID count
	if data, err := os.ReadFile(pidCurrentPath); err == nil {
		if value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
			stats.Current = value
		}
	}

	// Read PID limit
	if data, err := os.ReadFile(pidMaxPath); err == nil {
		limitStr := strings.TrimSpace(string(data))
		if limitStr != "max" {
			if value, err := strconv.ParseUint(limitStr, 10, 64); err == nil {
				stats.Limit = value
			}
		}
	}

	return stats, nil
}

// FormatMetrics formats metrics for human-readable output
func FormatMetrics(m *ContainerMetrics) string {
	return fmt.Sprintf(`Container: %s
Timestamp: %s

CPU:
  Usage: %.2f%%
  User: %d ns
  System: %d ns
  Throttled: %d times (%d ns)

Memory:
  Usage: %s / %s (%.2f%%)
  RSS: %s
  Cache: %s
  Swap: %s

Network:
  RX: %s (%d packets, %d errors, %d dropped)
  TX: %s (%d packets, %d errors, %d dropped)

Disk I/O:
  Read: %s (%d ops)
  Write: %s (%d ops)

PIDs:
  Current: %d
  Limit: %d
`,
		m.ContainerID,
		m.Timestamp.Format(time.RFC3339),
		m.CPUStats.PercentCPU,
		m.CPUStats.UserNanos,
		m.CPUStats.SystemNanos,
		m.CPUStats.ThrottledCount,
		m.CPUStats.ThrottledTime,
		formatBytes(m.MemoryStats.Usage),
		formatBytes(m.MemoryStats.Limit),
		m.MemoryStats.PercentUsed,
		formatBytes(m.MemoryStats.RSS),
		formatBytes(m.MemoryStats.Cache),
		formatBytes(m.MemoryStats.Swap),
		formatBytes(m.NetworkStats.RxBytes),
		m.NetworkStats.RxPackets,
		m.NetworkStats.RxErrors,
		m.NetworkStats.RxDropped,
		formatBytes(m.NetworkStats.TxBytes),
		m.NetworkStats.TxPackets,
		m.NetworkStats.TxErrors,
		m.NetworkStats.TxDropped,
		formatBytes(m.DiskStats.ReadBytes),
		m.DiskStats.ReadOps,
		formatBytes(m.DiskStats.WriteBytes),
		m.DiskStats.WriteOps,
		m.PIDStats.Current,
		m.PIDStats.Limit,
	)
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
