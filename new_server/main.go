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

func loadTemplates() *Template {
	_, err := template.ParseFiles("../static_cui/cui/templates/cui.html")
	fmt.Printf("Error is: %v\n", err)
	t := &Template{
		// Cached templates
		templates: template.Must(template.ParseFiles("../static_cui/cui/templates/cui.html")),
	}
	return t
}

func main() {

	// Echo instance
	e := echo.New()
	e.Hook(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		l := len(path) - 1
		if path != "/" && path[l] == '/' {
			r.URL.Path = path[:l]
		}
	})
	t := loadTemplates()
	e.SetRenderer(t)

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	// Routes
	e.Get("/hello", hello)
	e.Static("/static/cui", "../static_cui/cui/static/cui")
	e.Get("/cui", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "cui.html", map[string]string{"Title": "Goonj"})
	})

	a := e.Group("/c")

	a.Post("/_start", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Started")
	})

	// Start server
	e.Run(":1323")
}
