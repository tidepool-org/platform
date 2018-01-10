package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	quitChannel := make(chan os.Signal)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}
