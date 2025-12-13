package osinfo

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed templates
var embeddedFiles embed.FS

var dashboardTemplate *template.Template

func init() {
	tmpl, err := template.ParseFS(embeddedFiles, "templates/*.html")
	if err != nil {
		panic(err)
	}
	dashboardTemplate = tmpl
}

// Serve dashboard HTML
func serveDashboard(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")

	err := dashboardTemplate.ExecuteTemplate(c.Writer, "dashboard.html", gin.H{
		"title": "OS Metrics Dashboard",
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
	}
}

// Serve static files
func staticHandler(c *gin.Context) {
	file := c.Param("filepath")
	c.FileFromFS(file, http.FS(embeddedFiles))
}
