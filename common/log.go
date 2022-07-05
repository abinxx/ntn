package common

import (
	"log"
	"os"
)

func SetLog(path string) {
	if path == "" {
		return //为空则输出到控制台
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("CREATE LOG FILE ERROR:", err)
	}
	log.SetOutput(file)
}
