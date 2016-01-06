package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/maddyonline/goonj/cui"
	"github.com/maddyonline/goonj/utils"
	"golang.org/x/oauth2"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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

const DEFAULT_PORT = "3000"

var (
	throttle = time.Tick(1 * time.Second)
)

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

type UserContext struct {
	githubClient *github.Client
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

var TMP_DIR string

func getTmpWorkDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return utils.CreateDirIfReqd("goonj-workdir")
	}
	return utils.CreateDirIfReqd(filepath.Join(u.HomeDir, "goonj-workdir"))
}

func saveSolution(c *echo.Context) *cui.Task {
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
		return nil
	}

	log.Info("%s %s task.ProgLang: %s, solnReq.ProgLang: %s", c.Request().Method, c.Request().URL, task.ProgLang, solnReq.ProgLang)

	ext := map[string]string{"cpp": "cpp", "c": "c", "py2": "py", "py3": "py", "go": "go"}[solnReq.ProgLang]
	file := fmt.Sprintf("%s/%s/%s/main.%s", TMP_DIR, solnReq.Ticket, solnReq.Task, ext)
	log.Info("Writing soln to %s", file)
	err := utils.UpdateFile(file, solnReq.Solution)
	if err != nil {
		panic(err)
	}
	//saveAsGist(githubClient, file, solnReq.Solution)
	log.Info("Storing solution as gist: %s", solnReq.Solution)
	storeKey, filename, filecontent := solnReq.Ticket, strings.Join([]string{solnReq.Task, filepath.Base(file)}, "-"), string(solnReq.Solution)
	user, ok := userContexts[solnReq.Ticket]
	log.Info("ticket, user, ok: %s, %v, %v", solnReq.Ticket, user, ok)
	saveAsGist(user.githubClient, storeKey, filename, filecontent)
	task.Src = file
	task.CurrentSolution = solnReq.Solution
	task.ProgLang = solnReq.ProgLang
	return task
}

func addCuiHandlers(e *echo.Echo) {
	c := e.Group("/c")
	c.Post("/_start", func(c *echo.Context) error {
		session, ok := cuiSessions[c.Form("ticket")]
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "Attempt to start an invalid session")
		}
		session.StartTime = time.Now()
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
		saveSolution(c)
		return c.String(http.StatusOK, "Finished saving")
	})

	chk.Post("/verify", func(c *echo.Context) error {
		c.Form("task")
		log.Info("/verify: %#v", c.Request().Form)
		task := saveSolution(c)
		return c.XML(http.StatusOK, cui.GetVerifyStatus(task))
	})

	chk.Post("/final", func(c *echo.Context) error {
		task := saveSolution(c)
		return c.XML(http.StatusOK, cui.GetVerifyStatus(task))
	})
}

var (
	githubClient *github.Client
	gistStore    = map[string]*github.Gist{}
)

func NewGitHubClient(secret string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: secret},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc)
}

func saveAsGist(client *github.Client, key, filename, filecontent string) {
	savedGist, ok := gistStore[key]
	if !ok {
		description := "Think Hike App"
		metaFilename := "meta.json"
		metaFilecontent := fmt.Sprintf(`{"key": "%s"}`, key)
		gfiles := map[github.GistFilename]github.GistFile{
			github.GistFilename(filename):     github.GistFile{Filename: &filename, Content: &filecontent},
			github.GistFilename(metaFilename): github.GistFile{Filename: &metaFilename, Content: &metaFilecontent},
		}
		log.Info("Finally sending this across: %#v", gfiles)
		g, _, err := client.Gists.Create(&github.Gist{
			Description: &description,
			Files:       gfiles,
		})
		if err != nil {
			log.Info("Got error while creating gist: %s", err)
		} else {
			log.Info("Created gist: %s", g)
			gistStore[key] = g
		}
	} else {
		savedGist.Files[github.GistFilename(filename)] = github.GistFile{Filename: &filename, Content: &filecontent}
		g, _, err := client.Gists.Edit(*savedGist.ID, savedGist)
		if err != nil {
			log.Info("Got error while saving gist: %s", err)
		} else {
			log.Info("Saved gist: %s", g)
			gistStore[key] = g
		}
	}
}

func readDotEnv(root string) (map[string]string, error) {
	env := map[string]string{}
	read, err := ioutil.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		return map[string]string{}, err
	}
	err = json.Unmarshal(read, &env)
	return env, err
}

var (
	userContexts = map[string]*UserContext{}
)

func main() {
	initializeConfig()
	port := opts.Port
	staticDir := filepath.Join(opts.StaticFilesRoot, "static_cui/cui/static/cui")
	templatesDir := filepath.Join(opts.StaticFilesRoot, "static_cui/cui/templates")
	log.Info("Using Port=%s", port)
	log.Info("Using Static Directory=%s", staticDir)
	log.Info("Using Templates Directory=%s", templatesDir)

	env, err := readDotEnv(opts.StaticFilesRoot)
	if err != nil {
		log.Fatal("Got error while reading dotenv: %v", err)
		return
	} else {
		log.Info("Read env: %s", env)
	}
	THINK_GISTS_KEY, ok := env["THINK_GISTS_KEY"]
	if !ok {
		log.Fatal("Need github secret to proceed")
		return
	}
	AUTH0_TOKEN, ok := env["AUTH0_TOKEN"]
	if !ok {
		log.Fatal("Need AUTH0 token to proceed")
		return
	}

	//initializeGitClient(secret)
	//saveAsGist(githubClient, "abc.txt", "this is cool")
	//saveAsGist(githubClient, "abc.txt", "this is fun")

	schemaDecoder = schema.NewDecoder()
	cuiSessions = map[string]*cui.Session{}
	tasks = map[cui.TaskKey]*cui.Task{}

	TMP_DIR, err = getTmpWorkDir()
	if err != nil {
		log.Fatal("Failed to initialize tmp_dir: %v", err)
		return
	}

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
	e.Index(filepath.Join(opts.StaticFilesRoot, "client-app/index.html"))
	e.Static("/static/", filepath.Join(opts.StaticFilesRoot, "client-app/static"))

	// Initial API call
	e.Get("/secured/ping", func(c *echo.Context) error {
		user_id := c.Query("user_id")
		log.Info("user_id: %s", user_id)
		url := fmt.Sprintf("https://thinkhike.auth0.com/api/v2/users/%s", user_id)
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", AUTH0_TOKEN))
		q := req.URL.Query()
		q.Add("fields", "identities")
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		log.Info("Request: %v", resp.StatusCode)
		errorStr := ""
		if err != nil {
			errorStr = fmt.Sprintf("Error : %s", err)
			log.Error("Req error: %v", err)
		}
		defer resp.Body.Close()
		type Auth0Data struct {
			Identities []struct {
				AccessToken *string `json:"access_token"`
			}
		}
		expected := &struct {
			Ticket string `json:"ticket_id"`
			Error  string
			Data   *Auth0Data
		}{
			Error: "",
			Data:  &Auth0Data{},
		}
		err = json.NewDecoder(resp.Body).Decode(expected.Data)
		if err != nil {
			log.Error("Decode error: %v", err)
			errorStr += fmt.Sprintf(", %s", err)
		}
		expected.Error = errorStr
		log.Info("Got: access token: %s", *expected.Data.Identities[0].AccessToken)
		USER_GH_TOKEN := *expected.Data.Identities[0].AccessToken
		user := &UserContext{githubClient: NewGitHubClient(USER_GH_TOKEN)}
		ticket := cui.NewTicket(nil)
		cuiSessions[ticket.Id] = &cui.Session{TimeLimit: 3600, Created: time.Now(), Ticket: ticket}
		userContexts[ticket.Id] = user
		expected.Ticket = ticket.Id
		return c.JSON(http.StatusOK, expected)
	})

	// Remaining routes
	e.Get("/hello", hello)
	e.Static("/static/cui", staticDir)
	e.Get("/cui/:ticket_id", func(c *echo.Context) error {
		ticket_id := c.Param("ticket_id")
		log.Info("Ticket: %s", ticket_id)
		session, ok := cuiSessions[ticket_id]
		if !ok {
			return echo.NewHTTPError(http.StatusNotFound, "No valid session found")
		}
		if time.Now().Sub(session.Created) > time.Duration(5*time.Minute) {
			return echo.NewHTTPError(http.StatusNotFound, "Session Expired")
		}
		if !session.Started {
			session.Started = true
		}
		log.Info("Session Started? %v", session.Started)
		return c.Render(http.StatusOK, "cui.html", map[string]interface{}{"Title": "Goonj", "Ticket": session.Ticket})
	})
	e.Get("/cui/new", func(c *echo.Context) error {
		user := &UserContext{githubClient: NewGitHubClient(THINK_GISTS_KEY)}
		ticket := cui.NewTicket(nil)
		cuiSessions[ticket.Id] = &cui.Session{TimeLimit: 3600, Created: time.Now(), Ticket: ticket}
		userContexts[ticket.Id] = user
		return c.JSON(http.StatusOK, map[string]string{"ticket_id": ticket.Id})
	})

	addCuiHandlers(e)

	// Start server
	e.Run(fmt.Sprintf(":%s", port))
}
