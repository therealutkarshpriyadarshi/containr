package rbac

import (
	"fmt"
)

// ResourceQuota defines resource limits for a user
type ResourceQuota struct {
	MaxContainers int     `json:"max_containers" yaml:"max_containers"`
	MaxCPU        float64 `json:"max_cpu" yaml:"max_cpu"`
	MaxMemory     int64   `json:"max_memory" yaml:"max_memory"`
	MaxStorage    int64   `json:"max_storage" yaml:"max_storage"`
}

// ResourceUsage tracks current resource usage
type ResourceUsage struct {
	Containers int     `json:"containers"`
	CPU        float64 `json:"cpu"`
	Memory     int64   `json:"memory"`
	Storage    int64   `json:"storage"`
}

// NewResourceQuota creates a new resource quota
func NewResourceQuota(containers int, cpu float64, memory, storage int64) *ResourceQuota {
	return &ResourceQuota{
		MaxContainers: containers,
		MaxCPU:        cpu,
		MaxMemory:     memory,
		MaxStorage:    storage,
	}
}

// Check checks if usage is within quota
func (q *ResourceQuota) Check(usage *ResourceUsage) error {
	if usage.Containers > q.MaxContainers {
		return fmt.Errorf("container quota exceeded: %d/%d", usage.Containers, q.MaxContainers)
	}

	if usage.CPU > q.MaxCPU {
		return fmt.Errorf("CPU quota exceeded: %.2f/%.2f", usage.CPU, q.MaxCPU)
	}

	if usage.Memory > q.MaxMemory {
		return fmt.Errorf("memory quota exceeded: %d/%d bytes", usage.Memory, q.MaxMemory)
	}

	if usage.Storage > q.MaxStorage {
		return fmt.Errorf("storage quota exceeded: %d/%d bytes", usage.Storage, q.MaxStorage)
	}

	return nil
}

// Percentage returns the percentage of quota used
func (q *ResourceQuota) Percentage(usage *ResourceUsage) QuotaPercentage {
	containerPct := float64(usage.Containers) / float64(q.MaxContainers) * 100
	cpuPct := usage.CPU / q.MaxCPU * 100
	memoryPct := float64(usage.Memory) / float64(q.MaxMemory) * 100
	storagePct := float64(usage.Storage) / float64(q.MaxStorage) * 100

	return QuotaPercentage{
		Containers: containerPct,
		CPU:        cpuPct,
		Memory:     memoryPct,
		Storage:    storagePct,
	}
}

// QuotaPercentage represents quota usage percentages
type QuotaPercentage struct {
	Containers float64 `json:"containers"`
	CPU        float64 `json:"cpu"`
	Memory     float64 `json:"memory"`
	Storage    float64 `json:"storage"`
}

// IsOverQuota checks if any resource is over quota
func (qp *QuotaPercentage) IsOverQuota() bool {
	return qp.Containers > 100 || qp.CPU > 100 || qp.Memory > 100 || qp.Storage > 100
}

// IsNearQuota checks if any resource is near quota (>80%)
func (qp *QuotaPercentage) IsNearQuota() bool {
	return qp.Containers > 80 || qp.CPU > 80 || qp.Memory > 80 || qp.Storage > 80
}
