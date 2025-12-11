package main

import (
	"github.com/gin-gonic/gin"
	"github.com/raza001/go-osinfo-gin"
)

func main() {
	r := gin.Default()

	// Register under /os
	osinfo.RegisterRoutes(r, "/os")

	r.Run(":8080")
}
