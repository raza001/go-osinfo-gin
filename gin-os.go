package osinfo

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cpu "github.com/shirou/gopsutil/v3/cpu"
	disk "github.com/shirou/gopsutil/v3/disk"
	host "github.com/shirou/gopsutil/v3/host"
	mem "github.com/shirou/gopsutil/v3/mem"
)

// Metrics tracks request statistics
type Metrics struct {
	mu                sync.RWMutex
	TotalRequests     int64
	TotalResponseTime int64
	StatusCodes       map[int]int64
	StartTime         time.Time
}

var metrics = &Metrics{
	StatusCodes: make(map[int]int64),
	StartTime:   time.Now(),
}

// RegisterRoutes registers all OS endpoints and dashboard
func RegisterRoutes(r gin.IRouter, prefix string) {

	// Middleware for metrics
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

	// Prometheus handler
	grp.GET("/gui-metrics", gin.WrapH(promhttp.Handler()))

	// Dashboard UI
	grp.GET("/dashboard", serveDashboard)

	// Static files
	grp.GET("/static/*filepath", staticHandler)
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
	out := []gin.H{}
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
	c.JSON(http.StatusOK, gin.H{"env": os.Environ()})
}

// ===== METRICS =====

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

func metricsHandler(c *gin.Context) {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	avg := float64(0)
	if metrics.TotalRequests > 0 {
		avg = float64(metrics.TotalResponseTime) / float64(metrics.TotalRequests)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_requests":       metrics.TotalRequests,
		"avg_response_time_ms": avg,
		"status_codes":         metrics.StatusCodes,
	})
}

func serverUptimeHandler(c *gin.Context) {
	uptime := time.Since(metrics.StartTime).Seconds()
	c.JSON(http.StatusOK, gin.H{
		"server_uptime_seconds": uptime,
		"server_start_time":     metrics.StartTime,
	})
}
