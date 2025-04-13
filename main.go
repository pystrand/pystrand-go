package main

import (
	"os"
	"os/signal"

	"github.com/pystrand/pystrand-server/bridge"
)

func main() {
	bridge := bridge.NewBridge()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go bridge.Start()
	<-signalChan
	bridge.Stop()
}
