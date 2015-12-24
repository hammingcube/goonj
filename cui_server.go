package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/maddyonline/goonj/cui"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var schemaDecoder *schema.Decoder

var opts struct {
	Port            string
	StaticFilesRoot string
}

func assignString(v *string, args ...string) {
	for _, arg := range args {
		if arg != "" {
			*v = arg
			break
		}
	}
}

const ENV_PORT_NAME = "CUI_PORT"
const ENV_STATIC_FILES_DIR = "CUI_STATIC_FILES_DIR"

const DEFAULT_PORT = "1323"
const DEFAULT_STATIC_DIR = "."

func initializeConfig() {
	var oldUsage = flag.Usage
	var newUsage = func() {
		oldUsage()
		fmt.Fprintf(os.Stderr, "Alternatively, you may set environment variables %s and %s\n", ENV_PORT_NAME, ENV_STATIC_FILES_DIR)
	}
	flag.Usage = newUsage
	flag.StringVar(&opts.Port, "port", "", "Port on which server runs")
	flag.StringVar(&opts.StaticFilesRoot, "static", "", "Path to static directory")
	flag.Parse()
	assignString(&opts.Port, opts.Port, os.Getenv(ENV_PORT_NAME), DEFAULT_PORT)
	defaultDir, err := filepath.Abs(DEFAULT_STATIC_DIR)
	if err != nil {
		log.Error("Got error while taking absolute path of %s: %v", DEFAULT_STATIC_DIR, err)
		defaultDir = "."
	}
	assignString(&opts.StaticFilesRoot, opts.StaticFilesRoot, os.Getenv(ENV_STATIC_FILES_DIR), defaultDir)
}

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

func loadTemplates(templatesDir string) *Template {
	t := &Template{
		// Cached templates
		templates: template.Must(template.ParseFiles(filepath.Join(templatesDir, "cui.html"))),
	}
	return t
}

var cui_html []byte

var tasks map[cui.TaskKey]*cui.Task
var cuiSessions map[string]*cui.Session
var toggle bool

func addCuiHandlers(e *echo.Echo) {
	c := e.Group("/c")
	c.Post("/_start", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Started")
	})
	c.Post("/_get_task", func(c *echo.Context) error {
		val := &cui.ClientGetTaskMsg{
			Task:                 c.Form("task"),
			Ticket:               c.Form("ticket"),
			ProgLang:             c.Form("prg_lang"),
			HumanLang:            c.Form("human_lang"),
			PreferServerProgLang: c.Form("prefer_server_prg_lang") == "false",
		}
		return c.XML(http.StatusOK, cui.GetTask(tasks, val))
	})

	chk := e.Group("/chk")
	chk.Post("/clock", func(c *echo.Context) error {
		c.Request().ParseForm()
		clkReq := &cui.ClockRequest{}
		schemaDecoder.Decode(clkReq, c.Request().Form)
		log.Info("Clock Request: %v", clkReq)
		oldlimit := time.Duration(clkReq.OldTimeLimit) * time.Second
		resp := cui.GetClock(cuiSessions, clkReq)
		newlimit := time.Duration(resp.NewTimeLimit) * time.Second
		log.Info("Clock Request: OldLimit=%s", oldlimit)
		log.Info("Clock Response: NewLimit=%s", newlimit)
		return c.XML(http.StatusOK, resp)
	})
	chk.Post("/save", func(c *echo.Context) error {
		val := struct {
			Task     string
			Ticket   string
			ProgLang string
			Solution string
		}{
			Task:     c.Form("task"),
			Ticket:   c.Form("ticket"),
			ProgLang: c.Form("prg_lang"),
			Solution: c.Form("solution"),
		}
		v, _ := json.Marshal(val)
		log.Info("%s", string(v))
		key := cui.TaskKey{val.Ticket, val.Task}
		tasks[key].CurrentSolution = val.Solution
		tasks[key].ProgLang = val.ProgLang
		return c.String(http.StatusOK, "Finished saving")
	})

	chk.Post("/verify", func(c *echo.Context) error {
		toggle = !toggle
		return c.XML(http.StatusOK, cui.GetVerifyStatus(toggle))
	})
	chk.Post("/final", func(c *echo.Context) error {
		return c.XML(http.StatusOK, cui.GetVerifyStatus(true))
	})
}

func main() {
	initializeConfig()
	port := opts.Port
	staticDir := filepath.Join(opts.StaticFilesRoot, "static_cui/cui/static/cui")
	templatesDir := filepath.Join(opts.StaticFilesRoot, "static_cui/cui/templates")
	log.Info("Using Port=%s", port)
	log.Info("Using Static Directory=%s", staticDir)
	log.Info("Using Templates Directory=%s", templatesDir)

	schemaDecoder = schema.NewDecoder()
	cuiSessions = map[string]*cui.Session{}
	tasks = map[cui.TaskKey]*cui.Task{}

	// Echo instance
	e := echo.New()
	e.Hook(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		l := len(path) - 1
		if path != "/" && path[l] == '/' {
			r.URL.Path = path[:l]
		}
	})
	t := loadTemplates(templatesDir)
	e.SetRenderer(t)

	// Middleware
	e.Use(mw.Logger())
	//e.Use(mw.Recover())

	// Routes
	e.Get("/hello", hello)
	e.Static("/static/cui", staticDir)
	e.Get("/cui", func(c *echo.Context) error {
		type Ticket struct {
			Id string
		}
		ticket := &Ticket{RandId()}
		cuiSessions[ticket.Id] = &cui.Session{StartTime: time.Now(), TimeLimit: 3600}
		return c.Render(http.StatusOK, "cui.html", map[string]interface{}{"Title": "Goonj", "Ticket": ticket})
	})

	addCuiHandlers(e)

	// Start server
	e.Run(fmt.Sprintf(":%s", port))
}
