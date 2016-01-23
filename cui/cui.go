package cui

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/maddyonline/code"
	"github.com/maddyonline/goonj/utils"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

type HumanLang struct {
	Name string `json:"name_in_itself"`
}
type ProgLang struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Options struct {
	TicketId         string               `json:"ticket_id"`
	TimeElapsed      int                  `json:"time_elpased_sec"`
	TimeRemaining    int                  `json:"time_remaining_sec"`
	CurrentHumanLang string               `json:"current_human_lang"`
	CurrentProgLang  string               `json:"current_prg_lang"`
	CurrentTaskName  string               `json:"current_task_name"`
	TaskNames        []string             `json:"task_names"`
	HumanLangList    map[string]HumanLang `json:"human_langs"`
	ProgLangList     map[string]ProgLang  `json:"prg_langs"`
	ShowSurvey       bool                 `json:"show_survey"`
	ShowHelp         bool                 `json:"show_help"`
	ShowWelcome      bool                 `json:"show_welcome"`
	Sequential       bool                 `json:"sequential"`
	SaveOften        bool                 `json:"save_often"`
	Urls             map[string]string    `json:"urls"`
}

type Ticket struct {
	Id      string
	Options *Options
}

type Session struct {
	Ticket    *Ticket
	StartTime time.Time
	Created   time.Time
	Started   bool
	TimeLimit int
}

func addToTask(tasks map[TaskKey]*Task, ticketId string, input *code.Input, prefix string) *Task {
	if len(input.Files) < 1 {
		return nil
	}
	task := NewTask()
	file := input.Files[0]
	taskName := strings.Join([]string{"task", prefix, "main"}, "-")
	task.Id = taskName
	task.CurrentSolution = file.Content
	task.ProgLang = LanguageFromRunner(input.Language)
	tasks[TaskKey{ticketId, taskName}] = task
	return task
}

func LoadTicket(tasks map[TaskKey]*Task, opts *Options) *Ticket {
	ticketId := utils.RandId()
	gistId := "4f1bae999b5fbea43624"
	evalContext := code.GistFetch(gistId)
	task := addToTask(tasks, ticketId, evalContext.Test, "test")
	task.JudgeSolution = evalContext.Solution
	task.Generator = evalContext.Generator
	taskIds := []string{task.Id}
	if opts == nil {
		opts = DefaultOptions()
	}
	opts.TicketId = ticketId
	opts.TaskNames = taskIds
	opts.CurrentTaskName = task.Id
	opts.CurrentProgLang = task.ProgLang
	return &Ticket{Id: ticketId, Options: opts}
}

func NewTicket(tasks map[TaskKey]*Task, opts *Options) *Ticket {
	ticketId := utils.RandId()
	gistId := "4f1bae999b5fbea43624"
	evalContext := code.GistFetch(gistId)
	addToTask(tasks, ticketId, evalContext.Generator, "generator")
	addToTask(tasks, ticketId, evalContext.Solution, "solution")
	addToTask(tasks, ticketId, evalContext.Test, "test")

	taskIds := []string{}
	for key := range tasks {
		taskIds = append(taskIds, key.TaskId)
	}

	if opts == nil {
		opts = DefaultOptions()
	}
	opts.TicketId = ticketId
	opts.TaskNames = taskIds
	return &Ticket{Id: ticketId, Options: opts}
}

func DefaultOptions() *Options {
	opts := &Options{
		TicketId:         "",
		TimeElapsed:      5,
		TimeRemaining:    3600,
		CurrentHumanLang: "en",
		CurrentProgLang:  "c",
		CurrentTaskName:  "task1",
		TaskNames:        []string{"task1", "task2", "task3"},
		HumanLangList: map[string]HumanLang{
			"en": HumanLang{Name: "English"},
			"cn": HumanLang{Name: "\u4e2d\u6587"},
		},
		ProgLangList: map[string]ProgLang{
			"c":   ProgLang{Version: "C", Name: "C"},
			"cpp": ProgLang{Version: "C++", Name: "C++"},
			"py2": ProgLang{Version: "py2", Name: "Python 2"},
			"py3": ProgLang{Version: "py3", Name: "Python 3"},
			"go":  ProgLang{Version: "go", Name: "Go"},
			"js":  ProgLang{Version: "js", Name: "Javascript"},
		},
		ShowSurvey:  false,
		ShowWelcome: false,
		Sequential:  false,
		SaveOften:   true,
		Urls: map[string]string{
			"status":         "/chk/status/",
			"get_task":       "/c/_get_task/",
			"submit_survey":  "/surveys/_ajax_submit_candidate_survey/TICKET_ID/",
			"clock":          "/chk/clock/",
			"close":          "/c/close/TICKET_ID",
			"verify":         "/chk/verify/",
			"judge":          "/chk/judge/",
			"save":           "/chk/save/",
			"timeout_action": "/chk/timeout_action/",
			"final":          "/chk/final/",
			"start_ticket":   "/c/_start/",
		},
	}
	return opts
}

type TaskKey struct {
	TicketId string
	TaskId   string
}

type ClientGetTaskMsg struct {
	Task                 string
	Ticket               string
	ProgLang             string
	HumanLang            string
	PreferServerProgLang bool
}

type Task struct {
	XMLName          xml.Name    `xml:"response"`
	Id               string      `xml:"id" json:"id"`
	Status           string      `xml:"task_status" json: "task_status"`
	Description      string      `xml:"task_description"`
	Type             string      `xml:"task_type"`
	SolutionTemplate string      `xml:"solution_template"`
	CurrentSolution  string      `xml:"current_solution"`
	ExampleInput     string      `xml:"example_input"`
	ProgLangList     string      `xml:"prg_lang_list"`
	HumanLangList    string      `xml:"human_lang_list"`
	ProgLang         string      `xml:"prg_lang"`
	HumanLang        string      `xml:"human_lang"`
	Src              string      `xml:"-"`
	Generator        *code.Input `xml:"-"`
	JudgeSolution    *code.Input `xml:"-"`
}

type ClockRequest struct {
	TicketId     string `schema:"ticket"`
	OldTimeLimit int    `schema:"old_timelimit"`
}
type ClockResponse struct {
	XMLName      xml.Name `xml:"response"`
	Result       string   `xml:"result"`
	NewTimeLimit int      `xml:"new_timelimit"`
}

type SolutionRequest struct {
	Ticket    string `schema:"ticket"`
	Task      string `schema:"task"`
	ProgLang  string `scheam:"prg_lang"`
	Solution  string `schema:"solution"`
	TestData0 string `schema:"test_data0"`
	TestData1 string `schema:"test_data1"`
	TestData2 string `schema:"test_data2"`
	TestData3 string `schema:"test_data3"`
	TestData4 string `schema:"test_data4"`
}

type Status struct {
	OK      int    `xml:"ok"`
	Message string `xml:"message"`
}
type MainStatus struct {
	Compile   Status `xml:"compile"`
	Example   Status `xml:"example"`
	TestData0 Status `xml:"test_data0"`
	TestData1 Status `xml:"test_data1"`
	TestData2 Status `xml:"test_data2"`
	TestData3 Status `xml:"test_data3"`
	TestData4 Status `xml:"test_data4"`
}
type VerifyStatus struct {
	XMLName xml.Name   `xml:"response"`
	Result  string     `xml:"result"`
	Message string     `xml:"message"`
	Id      string     `xml:"id"`
	Delay   int        `xml:"delay"`
	Extra   MainStatus `xml:"extra"`
	//NextTask string     `xml:"next_task"`
}

func laterReply() *VerifyStatus {
	resp := &VerifyStatus{
		Result:  "LATER",
		Message: "We are still evaluating the solution",
		Id:      "submission_id: 23e3",
		Delay:   60,
	}
	return resp
}

type Mode int

const (
	VERIFY Mode = iota
	JUDGE
	FINAL
)

func (t Mode) String() string {
	var val string
	switch t {
	case VERIFY:
		val = "VERIFY"
	case JUDGE:
		val = "JUDGE"
	case FINAL:
		val = "FINAL"
	}
	return val
}

func FileNameForCode(progLang string) string {
	ext := map[string]string{
		"cpp": "cpp",
		"c":   "cpp",
		"py2": "py",
		"py3": "py",
		"go":  "go",
		"js":  "js",
	}[progLang]
	return fmt.Sprintf("main.%s", ext)
}

func LanguageForRunner(progLang string) string {
	return map[string]string{
		"c":          "cpp",
		"cpp":        "cpp",
		"go":         "go",
		"javascript": "javascript",
		"js":         "javascript",
		"py2":        "python",
		"py3":        "python",
	}[progLang]
}

func LanguageFromRunner(progLang string) string {
	return map[string]string{
		"c":          "cpp",
		"cpp":        "cpp",
		"go":         "go",
		"javascript": "javascript",
		"py3":        "python",
	}[progLang]
}

func errorResponse(err error, v *VerifyStatus) *VerifyStatus {
	v.Extra.Compile.OK = 0
	v.Extra.Compile.Message = fmt.Sprintf("Something went wrong: %v", err)
	v.Extra.Example.OK = 0
	v.Extra.Example.Message = "Something went wrong"
	return v
}

func GetVerifyStatus(runner *code.Runner, task *Task, solnReq *SolutionRequest, mode Mode) *VerifyStatus {
	//return laterReply()
	resp := &VerifyStatus{
		Result: "OK",
		Extra: MainStatus{
			Compile:   Status{1, "The solution compiled flawlessly."},
			Example:   Status{1, "OK"},
			TestData0: Status{1, "OK"},
			TestData1: Status{1, "OK"},
			TestData2: Status{1, "OK"},
			TestData3: Status{1, "OK"},
			TestData4: Status{1, "OK"},
		},
	}
	if task == nil {
		return resp
	}
	content, err := ioutil.ReadFile(task.Src)
	if err != nil {
		return errorResponse(err, resp)
	}
	filename := filepath.Base(task.Src)
	language := LanguageForRunner(task.ProgLang)
	log.Info("Got testData:=>%s<=", solnReq.TestData0)
	input := code.MakeInput(language, filename, string(content), code.StdinFile(solnReq.TestData0))
	log.Info("In VerifyStatus, input: %s", input)
	log.Info("In mode %s", mode)
	switch mode {
	case VERIFY:
		out, err := runner.Run(input)
		if err != nil {
			return errorResponse(err, resp)
		}
		log.Info("In VerifyStatus, mode=%v, got stdout=%q, stderr=%q, err=%v", mode, out.Stdout, out.Stderr, err)
		if out.Stderr != "" || err != nil {
			err := errors.New(fmt.Sprintf("stderr: %s, err: %v", out.Stderr, err))
			return errorResponse(err, resp)
		}
		resp.Extra.Example.OK = 0
		resp.Extra.Example.Message = out.Stdout
	case JUDGE, FINAL:
		log.Info("Judge called")
		log.Info("In VerifyStatus, mode=%s", mode)
		mysoln := code.MakeInput(language, filename, string(content), code.StdinFile(""))
		log.Info("My soln: %#v", mysoln)
		result := code.Evaluate(task.Generator, mysoln, task.JudgeSolution, runner)
		log.Info("Got result of evaluation: %#v", result)
		resp.Extra.Example.Message = fmt.Sprintf("%#v", result)
	}
	return resp
}

func GetClock(sessions map[string]*Session, clkReq *ClockRequest) *ClockResponse {
	session, ok := sessions[clkReq.TicketId]
	if !ok {
		return &ClockResponse{Result: "OK", NewTimeLimit: clkReq.OldTimeLimit}
	}
	elapsed := int(time.Since(session.StartTime) / time.Second)
	remaining := session.TimeLimit - elapsed
	log.Info("elapsed: %s, remaining: %s", time.Duration(elapsed)*time.Second, time.Duration(remaining)*time.Second)
	if remaining < 0 {
		remaining = 0
	}
	log.Info("newTimeLimit: %v, that is, %s", remaining, time.Duration(remaining)*time.Second)
	return &ClockResponse{Result: "OK", NewTimeLimit: remaining}
}

const SOLN_TEMPL_CPP = `# include <iostream>
using namespace std;
int main() {
  string s;
  while(cin >> s) {
    cout << s.size() << endl;
  }
}
`

const DESC_TEMPL = `### Can a string be a palindrome?

Given a string (say appease) be converted into a palindrome by permuting it?
In this case, we see appease --> apesepa is possible.

What is a good algorithm for this?

#### Test Cases

#### Input

	appease
	appeal

#### Output

	1
	0


### Code Example

	func getTrue() bool {
	    return true
	}
`

func getDescFromMarkdown(input []byte) []byte {
	unsafe := blackfriday.MarkdownCommon(input)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return html
}

func NewTask() *Task {
	prg_lang_list, _ := json.Marshal([]string{"c", "cpp", "py2", "py3", "go", "js"})
	human_lang_list, _ := json.Marshal([]string{"en", "cn"})
	task := &Task{
		Id:               "",
		Status:           "open",
		Description:      "Placeholder",
		Type:             "algo",
		SolutionTemplate: "This is just a template",
		CurrentSolution:  "",
		ExampleInput:     "",
		ProgLangList:     string(prg_lang_list),
		HumanLangList:    string(human_lang_list),
		ProgLang:         "cpp",
		HumanLang:        "en",
	}
	return task
}

func GetTask(tasks map[TaskKey]*Task, val *ClientGetTaskMsg) *Task {
	key := TaskKey{val.Ticket, val.Task}
	prg_lang_list, _ := json.Marshal([]string{"c", "cpp", "py2", "py3", "go", "js"})
	human_lang_list, _ := json.Marshal([]string{"en", "cn"})
	task, ok := tasks[key]
	log.Info("Looking for %s in tasks: %v", key, ok)
	keys := []TaskKey{}
	for key := range tasks {
		keys = append(keys, key)
	}
	log.Info("len(tasks): %d, keys: %v", len(tasks), keys)
	if task == nil {
		log.Info("Serving task based on nil request")
		task = &Task{
			Id:               val.Task,
			Status:           "open",
			Description:      string(getDescFromMarkdown([]byte(DESC_TEMPL))),
			Type:             "algo",
			SolutionTemplate: "This is just a template",
			CurrentSolution:  SOLN_TEMPL_CPP,
			ExampleInput:     "",
			ProgLangList:     string(prg_lang_list),
			HumanLangList:    string(human_lang_list),
			ProgLang:         val.ProgLang,
			HumanLang:        val.HumanLang,
		}
		tasks[key] = task
	}
	log.Info("PREFER-SERVER-LANG: %v", val.PreferServerProgLang)
	if val.PreferServerProgLang {
		log.Info("Updating task %s prog-lang form %s to %s", task.Id, task.ProgLang, val.ProgLang)
		task.ProgLang = val.ProgLang
	}
	log.Info("Updating task %s prog-lang form %s to %s", task.Id, task.HumanLang, val.HumanLang)
	task.HumanLang = val.HumanLang
	return task
}
