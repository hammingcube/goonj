package main

import (
	"encoding/json"
	"encoding/xml"
	"log"
)

func getClock() []byte {
	resp := struct {
		XMLName      xml.Name `xml:"response"`
		Result       string   `xml:"result"`
		NewTimeLimit int      `xml:"new_timelimit"`
	}{
		Result:       "OK",
		NewTimeLimit: 3600,
	}
	xmlResp, err := xml.MarshalIndent(resp, " ", "    ")
	if err != nil {
		log.Printf("Error: %v", err)
		return []byte{}
	}
	return xmlResp
}

func getTask(val *ClientGetTaskMsg) []byte {
	prg_lang_list, err := json.Marshal([]string{"c", "cpp"})
	human_lang_list, err := json.Marshal([]string{"en", "cn"})
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
	log.Printf("Sending %s", task)
	xmlResp, err := xml.MarshalIndent(task, " ", "    ")
	if err != nil {
		return []byte{}
	}
	return xmlResp
}
