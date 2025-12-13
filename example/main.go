package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raza001/go-osinfo-gin"
)

func main() {
	r := gin.Default()

	// Register under /os
	osinfo.RegisterRoutes(r, "")

	r.GET("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	r.Run(":8080")
}
