package main

import (
	"fmt"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"html/template"
	"io"
	"net/http"
)

type (
	// Template provides HTML template rendering
	Template struct {
		templates *template.Template
	}
)

// Render HTML
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Handler
func hello(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func cui(c *echo.Context) error {
	return c.Render(http.StatusOK, "cui.html", map[string]string{"Title": "Goonj"})
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	// Routes
	e.Get("/", hello)

	_, err := template.ParseFiles("../static_cui/cui/templates/cui.html")
	fmt.Printf("Error is: %v\n", err)
	t := &Template{
		// Cached templates
		templates: template.Must(template.ParseFiles("../static_cui/cui/templates/cui.html")),
	}
	e.SetRenderer(t)
	e.Get("/cui", cui)
	e.Static("/static/cui", "../static_cui/cui/static/cui")

	// Start server
	e.Run(":1323")
}
