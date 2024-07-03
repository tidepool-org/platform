package utils

import (
	"encoding/json"
	"log"
	"os"
)

func writeFileData(data interface{}, fileName string) {
	if data == nil || fileName == "" {
		return
	}

	var handleErr = func(err error) {
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	err := os.MkdirAll(fileName, os.ModePerm)
	handleErr(err)
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	handleErr(err)

	defer f.Close()
	jsonData, err := json.Marshal(data)
	handleErr(err)
	f.WriteString(string(jsonData) + "\n")
}
