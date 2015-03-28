package main

import (
	"encoding/json"
	"encoding/xml"
)

func getClock() []byte {
	/*r := type struct {
		Result string `xml:"result"`
		NewTimeLimit int64 `xml:"new_timelimit"`
		}{ 
			Result: 'OK',
			NewTimeLimit: -1,
		}
	xmlResp, err := xml.MarshalIndent(r, " ", "    ")
	if err != nil {
		return []byte{}
	}*/
	return []byte{}
}

func getTask() []byte {
	type Task struct {
		XMLName   			xml.Name 		`xml:"response"`
		Status				string			`xml:"task_status" json: "task_status"`
		Description			string			`xml:"task_description"`
		Type				string			`xml:"task_type"`
		SolutionTemplate	string			`xml:"solution_template"`
		CurrentSolution		string			`xml:"current_solution"`
		ExampleInput		string			`xml:"example_input"`
		ProgLangList		string			`xml:"prg_lang_list"`
		HumanLangList		string			`xml:"human_lang_list"`
		ProgLang			string			`xml:"prg_lang"`
		HumanLang			string			`xml:"human_lang"`
	}
	prg_lang_list, err := json.Marshal([]string{"c", "cpp"})
	human_lang_list, err := json.Marshal([]string{"en", "cn"})

	task := Task{
		Status: "open",
		Description: "Description: task1,en,c",
		Type: "algo",
		SolutionTemplate: "",
		CurrentSolution: "",
		ExampleInput: "",
		ProgLangList: string(prg_lang_list),
		HumanLangList: string(human_lang_list),
		ProgLang: "c",
		HumanLang: "en",
	}
	xmlResp, err := xml.MarshalIndent(task, " ", "    ")
	if err != nil {
		return []byte{}
	}
	return xmlResp
}