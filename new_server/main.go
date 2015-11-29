package main

import (
	"encoding/json"
	"encoding/xml"
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

var cui_html []byte
var tasks map[string]*Task

type ClientGetTaskMsg struct {
	Task                 string
	Ticket               string
	ProgLang             string
	HumanLang            string
	PreferServerProgLang bool
}

type Task struct {
	XMLName          xml.Name `xml:"response"`
	Status           string   `xml:"task_status" json: "task_status"`
	Description      string   `xml:"task_description"`
	Type             string   `xml:"task_type"`
	SolutionTemplate string   `xml:"solution_template"`
	CurrentSolution  string   `xml:"current_solution"`
	ExampleInput     string   `xml:"example_input"`
	ProgLangList     string   `xml:"prg_lang_list"`
	HumanLangList    string   `xml:"human_lang_list"`
	ProgLang         string   `xml:"prg_lang"`
	HumanLang        string   `xml:"human_lang"`
}

func getTask(val *ClientGetTaskMsg) *Task {
	prg_lang_list, _ := json.Marshal([]string{"c", "cpp"})
	human_lang_list, _ := json.Marshal([]string{"en", "cn"})
	task := tasks[val.Task]
	if task == nil {
		task = &Task{
			Status:           "open",
			Description:      "Description: task1,en,c",
			Type:             "algo",
			SolutionTemplate: "",
			CurrentSolution:  "",
			ExampleInput:     "",
			ProgLangList:     string(prg_lang_list),
			HumanLangList:    string(human_lang_list),
			ProgLang:         val.ProgLang,
			HumanLang:        val.HumanLang,
		}
		tasks[val.Task] = task
	}
	task.ProgLang = val.ProgLang
	task.HumanLang = val.HumanLang
	return task
}

func main() {
	tasks = map[string]*Task{}
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

	a.Post("/_get_task", func(c *echo.Context) error {
		val := &ClientGetTaskMsg{
			Task:                 c.Form("task"),
			Ticket:               c.Form("ticket"),
			ProgLang:             c.Form("prg_lang"),
			HumanLang:            c.Form("human_lang"),
			PreferServerProgLang: c.Form("prefer_server_prg_lang") == "false",
		}
		fmt.Println(c.Request().Form)
		j, _ := json.Marshal(val)
		fmt.Printf("%s\n", string(j))
		task := getTask(val)
		return c.XML(http.StatusOK, task)
	})

	// Start server
	e.Run(":1323")
}
