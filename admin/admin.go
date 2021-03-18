package main

import (
	"io"
	"text/template"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	// Initialize the template engine
	temp := &Template{
		// The purpose of the precompilation process is to optimize the speed of later rendering of the template file
		templates: template.Must(template.ParseGlob("html/*.html")),
	}
	e.Renderer = temp
	e.GET("/index", Index)
	e.Start(":9529")
}

// Custom template engine
type Template struct {
	templates *template.Template
}

// Render Implement interface, Render function
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Invokes the template engine to render the template
	return t.templates.ExecuteTemplate(w, name, data)
}

// Index
func Index(c echo.Context) error {
	params := make(map[string]interface{})
	params["list"] = []map[string]interface{}{map[string]interface{}{
		"ip":    "127.0.0.1",
		"uuid":  "9527",
		"count": "100",
	}, map[string]interface{}{
		"ip":    "127.0.0.2",
		"uuid":  "9528",
		"count": "101",
	}}
	return c.Render(200, "index.html", params)
}
