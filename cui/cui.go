package cui

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"log"
	"os"
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

const tmpl = `var local_ui_options = {{.}}`

func Render(opts *Options) {
	const tpl = `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
    </head>
    <body>
    <script>var local_ui_options = {{.}}</script>
    </body>
</html>`
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(tpl)
	check(err)
	err = t.Execute(os.Stdout, opts)
	check(err)
}

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

type ClockResponse struct {
	XMLName      xml.Name `xml:"response"`
	Result       string   `xml:"result"`
	NewTimeLimit int      `xml:"new_timelimit"`
}

type Status struct {
	OK      int    `xml:"ok"`
	Message string `xml:"message"`
}
type MainStatus struct {
	Compile Status `xml:"compile"`
	Example Status `xml:"example"`
}
type VerifyStatus struct {
	XMLName xml.Name   `xml:"response"`
	Result  string     `xml:"result"`
	Extra   MainStatus `xml:"extra"`
	//NextTask string     `xml:"next_task"`
}

func GetVerifyStatus(final bool) *VerifyStatus {
	resp := &VerifyStatus{
		Result: "OK",
		Extra: MainStatus{
			Compile: Status{1, "The solution compiled flawlessly."},
			Example: Status{1, "OK"},
		},
	}
	if final {
		resp.Extra.Example.OK = 0
		resp.Extra.Example.Message = "Something went wrong"
	}
	return resp
}

func GetClock() *ClockResponse {
	return &ClockResponse{Result: "OK", NewTimeLimit: 3600}
}

func GetTask(tasks map[string]*Task, val *ClientGetTaskMsg) *Task {
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

const expected = `{
        ticket_id: "TICKET_ID",

        time_elapsed_sec: 15,
        time_remaining_sec: 1800,

        current_human_lang: "en",
        current_prg_lang: "c",
        current_task_name: "task1",

        task_names: ["task1", "task2", "task3"],

        human_langs: {
            "en": {"name_in_itself": "English"},
            "cn": {"name_in_itself": "\u4e2d\u6587"},
        },
        prg_langs: {
            "c": {"version": "C", "name": "C"},
            "sql": {"version": "SQL", "name": "SQL"},
            "cpp": {"version": "C++", "name": "C++"},
        },

        show_survey: true,
        show_help: false,
        show_welcome: true,
        sequential: false,
        save_often: true,

        urls: {
            "status": "/chk/status/",
            "get_task": "/c/_get_task/",
            "submit_survey": "/surveys/_ajax_submit_candidate_survey/TICKET_ID/",
            "clock": "/chk/clock/",
            "close": "/c/close/TICKET_ID",
            "verify": "/chk/verify/",
            "save": "/chk/save/",
            "timeout_action": "/chk/timeout_action/",
            "final": "/chk/final/",
            "start_ticket": "/c/_start/"
        },
        }`
