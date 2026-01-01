# go-osinfo-gin

A tiny Go package that exposes system/OS information and runtime metrics as HTTP endpoints using Gin.
`github.com/raza001/go-osinfo-gin`

## Features


- `/os/health` - simple health check
- `/os/info` - host info (platform, kernel, hostname)
- `/os/uptime` - uptime in seconds
- `/os/mem` - memory stats
- `/os/cpu` - CPU percent
- `/os/disk` - disk partitions and usage
- `/os/env` - environment variables


## Quick start


1. `go run example/main.go`
2. `curl http://localhost:8080/os/info`


## Notes


- Uses `github.com/shirou/gopsutil/v3` for system metrics. Works cross-platform but some fields depend on OS support.
- Keep in mind exposing environment variables and detailed host info is sensitive — protect these endpoints behind auth when running in production.
