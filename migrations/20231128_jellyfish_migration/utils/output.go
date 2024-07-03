package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func writeFileData(data interface{}, path string, name string) {
	if data == nil || path == "" || name == "" {
		return
	}

	var handleErr = func(err error) {
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	err := os.MkdirAll(path, os.ModePerm)
	handleErr(err)
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", path, name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	handleErr(err)

	defer f.Close()
	jsonData, err := json.Marshal(data)
	handleErr(err)
	f.WriteString(string(jsonData) + "\n")
}
