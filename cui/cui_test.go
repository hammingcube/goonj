package cui

import (
	"fmt"
	"testing"
)

// type Options struct {
// 	TicketId         string            `json:"ticket_id"`
// 	TimeElapsed      int               `json:"time_elpased_sec"`
// 	TimeRemaining    int               `json:"time_remaining_sec"`
// 	CurrentHumanLang string            `json:"current_human_lang"`
// 	CurrentProgLang  string            `json:"current_prg_lang"`
// 	CurrentTaskName  string            `json:"current_task_name"`
// 	TaskNames        []string          `json:"task_names"`
// 	HumanLangList    []HumanLang       `json:"human_langs"`
// 	ProgLangList     []ProgLang        `json:"prog_langs"`
// 	ShowSurvey       bool              `json:"show_survey"`
// 	ShowHelp         bool              `json:"show_help"`
// 	ShowWelcome      bool              `json:"show_welcome"`
// 	Sequential       bool              `json:"sequential"`
// 	SaveOften        bool              `json:"save_often"`
// 	Urls             map[string]string `json:"urls"`
// }

// ticket_id: "TICKET_ID",

// time_elapsed_sec: 15,
// time_remaining_sec: 1800,

// current_human_lang: "en",
// current_prg_lang: "c",
// current_task_name: "task1",

// task_names: ["task1", "task2", "task3"],

// human_langs: {
//     "en": {"name_in_itself": "English"},
//     "cn": {"name_in_itself": "\u4e2d\u6587"},
// },
// prg_langs: {
//     "c": {"version": "C", "name": "C"},
//     "sql": {"version": "SQL", "name": "SQL"},
//     "cpp": {"version": "C++", "name": "C++"},
// },

// show_survey: true,
// show_help: false,
// show_welcome: true,
// sequential: false,
// save_often: true,

// urls: {
//     "status": "/chk/status/",
//     "get_task": "/c/_get_task/",
//     "submit_survey": "/surveys/_ajax_submit_candidate_survey/TICKET_ID/",
//     "clock": "/chk/clock/",
//     "close": "/c/close/TICKET_ID",
//     "verify": "/chk/verify/",
//     "save": "/chk/save/",
//     "timeout_action": "/chk/timeout_action/",
//     "final": "/chk/final/",
//     "start_ticket": "/c/_start/"
// },
// }`

func TestOptions(t *testing.T) {
	opts := &Options{
		TicketId:         "my-ticket-id",
		TimeElapsed:      15,
		TimeRemaining:    1800,
		CurrentHumanLang: "en",
		CurrentProgLang:  "c++",
		CurrentTaskName:  "task1",
		TaskNames:        []string{"task1", "task2"},
		HumanLangList:    map[string]HumanLang{"en": HumanLang{"English"}, "cn": HumanLang{"\u4e2d\u6587"}},
		ProgLangList:     map[string]ProgLang{"c": ProgLang{"C", "C"}, "sql": ProgLang{"SQL", "SQL"}, "cpp": ProgLang{"C++", "C++"}},
		ShowSurvey:       true,
		ShowHelp:         false,
		ShowWelcome:      true,
		Sequential:       false,
		SaveOften:        true,
		Urls: map[string]string{
			"status":         "/chk/status/",
			"get_task":       "/c/_get_task/",
			"submit_survey":  "/surveys/_ajax_submit_candidate_survey/TICKET_ID/",
			"clock":          "/chk/clock/",
			"close":          "/c/close/TICKET_ID",
			"verify":         "/chk/verify/",
			"save":           "/chk/save/",
			"timeout_action": "/chk/timeout_action/",
			"final":          "/chk/final/",
			"start_ticket":   "/c/_start/",
		},
	}

	fmt.Println(opts)
	Render(opts)
}
