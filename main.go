package main

import (
	"fmt"
	"log/slog"
	"os"
	"yesbotics/ysm/cmd"
)

func main() {

	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Sprintf("Error opening file: %v", err))
	}
	defer f.Close()

	opts := &slog.HandlerOptions{
		//Level: slog.LevelDebug,
		Level: slog.LevelError,
	}

	logger := slog.New(slog.NewTextHandler(f, opts))
	slog.SetDefault(logger)

	cmd.Execute()
}
