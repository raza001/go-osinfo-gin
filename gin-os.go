package osinfo

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	cpu "github.com/shirou/gopsutil/v3/cpu"
	disk "github.com/shirou/gopsutil/v3/disk"
	host "github.com/shirou/gopsutil/v3/host"
	mem "github.com/shirou/gopsutil/v3/mem"
)

// Metrics tracks request statistics
type Metrics struct {
	mu                sync.RWMutex
	TotalRequests     int64
	TotalResponseTime int64 // in milliseconds
	StatusCodes       map[int]int64
	StartTime         time.Time
}

var metrics = &Metrics{
	StatusCodes: make(map[int]int64),
	StartTime:   time.Now(),
}

// RegisterRoutes registers a set of endpoints under the provided router group or engine.
// Example: RegisterRoutes(r, "/os") will create /os/health, /os/info, /os/mem, etc.
func RegisterRoutes(r gin.IRouter, prefix string) {
	// Add metrics middleware to all routes
	r.Use(metricsMiddleware())

	grp := r.Group(prefix)
	grp.GET("/health", healthHandler)
	grp.GET("/info", infoHandler)
	grp.GET("/uptime", uptimeHandler)
	grp.GET("/mem", memHandler)
	grp.GET("/cpu", cpuHandler)
	grp.GET("/disk", diskHandler)
	grp.GET("/env", envHandler)
	grp.GET("/metrics", metricsHandler)
	grp.GET("/server-uptime", serverUptimeHandler)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func infoHandler(c *gin.Context) {
	h, _ := host.Info()
	c.JSON(http.StatusOK, gin.H{
		"hostname":        h.Hostname,
		"uptime":          h.Uptime,
		"platform":        h.Platform,
		"platformFamily":  h.PlatformFamily,
		"platformVersion": h.PlatformVersion,
		"kernelVersion":   h.KernelVersion,
		"architecture":    h.KernelArch,
	})
}

func uptimeHandler(c *gin.Context) {
	u, err := host.Uptime()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"uptime_seconds": u})
}

func memHandler(c *gin.Context) {
	m, err := mem.VirtualMemory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total":       m.Total,
		"available":   m.Available,
		"used":        m.Used,
		"usedPercent": m.UsedPercent,
	})
}

func cpuHandler(c *gin.Context) {
	percent, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cpu_percent": percent})
}

func diskHandler(c *gin.Context) {
	parts, err := disk.Partitions(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(parts))
	for _, p := range parts {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		out = append(out, gin.H{
			"device":      p.Device,
			"mountpoint":  p.Mountpoint,
			"fstype":      p.Fstype,
			"total":       usage.Total,
			"free":        usage.Free,
			"used":        usage.Used,
			"usedPercent": usage.UsedPercent,
		})
	}
	c.JSON(http.StatusOK, out)
}

func envHandler(c *gin.Context) {
	envs := os.Environ()
	c.JSON(http.StatusOK, gin.H{"env": envs})
}

// metricsMiddleware tracks request metrics
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Milliseconds()

		metrics.mu.Lock()
		metrics.TotalRequests++
		metrics.TotalResponseTime += duration
		metrics.StatusCodes[c.Writer.Status()]++
		metrics.mu.Unlock()
	}
}

// metricsHandler returns collected request metrics
func metricsHandler(c *gin.Context) {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	avgResponseTime := float64(0)
	if metrics.TotalRequests > 0 {
		avgResponseTime = float64(metrics.TotalResponseTime) / float64(metrics.TotalRequests)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_requests":       metrics.TotalRequests,
		"total_response_time":  metrics.TotalResponseTime,
		"avg_response_time_ms": avgResponseTime,
		"status_codes":         metrics.StatusCodes,
	})
}

// serverUptimeHandler returns how long the server has been running
func serverUptimeHandler(c *gin.Context) {
	uptime := time.Since(metrics.StartTime).Seconds()
	c.JSON(http.StatusOK, gin.H{
		"server_uptime_seconds": uptime,
		"server_start_time":     metrics.StartTime,
	})
}
