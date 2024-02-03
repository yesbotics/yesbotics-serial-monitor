package main

import (
	"log"
	"os"
	"yesbotics/ysm/cmd"
)

func main() {
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	cmd.Execute()
}
