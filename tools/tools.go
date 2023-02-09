package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}
