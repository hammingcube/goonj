package cui

import (
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
