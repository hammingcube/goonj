package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/maddyonline/goonj/cui"
	"github.com/maddyonline/hey/utils"
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

	// Configure opts.Port
	assignString(&opts.Port, opts.Port, os.Getenv(ENV_PORT_NAME), DEFAULT_PORT)

	// Configure opts.StaticFilesRoot
	defaultDir := "."
	if GOPATH := os.Getenv("GOPATH"); GOPATH != "" {
		srcDir, err := filepath.Abs(filepath.Join(GOPATH, "src/github.com/maddyonline/goonj"))
		if err == nil {
			defaultDir = srcDir
		}
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

const TMP_DIR = "/Users/madhavjha/src/github.com/maddyonline/tempdir"

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
		solnReq := &cui.SolutionRequest{
			Ticket:   c.Form("ticket"),
			Task:     c.Form("task"),
			ProgLang: c.Form("prg_lang"),
			Solution: c.Form("solution"),
		}
		log.Info("%s %s Form: %#v", c.Request().Method, c.Request().URL, solnReq)
		log.Info("%s %s prg_lang: %s", c.Request().Method, c.Request().URL, c.Form("prg_lang"))
		task, ok := tasks[cui.TaskKey{solnReq.Ticket, solnReq.Task}]
		if !ok {
			return c.String(http.StatusOK, "Finished saving")
		}

		log.Info("%s %s task.ProgLang: %s, solnReq.ProgLang: %s", c.Request().Method, c.Request().URL, task.ProgLang, solnReq.ProgLang)

		ext := map[string]string{"cpp": "cpp", "c": "c"}[solnReq.ProgLang]
		file := fmt.Sprintf("%s/%s/%s/main.%s", TMP_DIR, solnReq.Ticket, solnReq.Task, ext)
		log.Info("Writing soln to %s", file)
		err := utils.UpdateFile(file, solnReq.Solution)
		if err != nil {
			panic(err)
		}
		task.Src = file
		task.CurrentSolution = solnReq.Solution
		task.ProgLang = solnReq.ProgLang
		return c.String(http.StatusOK, "Finished saving")
	})

	chk.Post("/verify", func(c *echo.Context) error {
		solnReq := &cui.SolutionRequest{
			Ticket:   c.Form("ticket"),
			Task:     c.Form("task"),
			ProgLang: c.Form("prg_lang"),
			Solution: c.Form("solution"),
		}
		log.Info("%s %s Form: %#v", c.Request().Method, c.Request().URL, solnReq)
		log.Info("%s %s prg_lang: %s", c.Request().Method, c.Request().URL, c.Form("prg_lang"))
		task, ok := tasks[cui.TaskKey{solnReq.Ticket, solnReq.Task}]
		if !ok {
			return c.String(http.StatusOK, "Finished saving")
		}
		log.Info("%s %s task.Src: %s", c.Request().Method, c.Request().URL, task.Src)
		return c.XML(http.StatusOK, cui.GetVerifyStatus(task.Src))
	})
	chk.Post("/final", func(c *echo.Context) error {
		return c.XML(http.StatusOK, cui.GetVerifyStatus(""))
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
