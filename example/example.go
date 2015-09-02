package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/scjalliance/logmein"
	"gopkg.in/gcfg.v1"
)

type config struct {
	RSS struct {
		ProfileID uint64 `gcfg:"profile"`
		Key       string `gcfg:"key"`
	} `gcfg:"rss"`
}

var configFile = flag.String("c", "config.conf", "Config file")

func main() {
	flag.Parse()

	cfg := new(config)
	err := gcfg.ReadFileInto(cfg, *configFile)
	if err != nil {
		fmt.Println("Config Error: ", err)
		os.Exit(1)
	}

	lmi := logmein.NewLMI(cfg.RSS.ProfileID, cfg.RSS.Key)
	computers := lmi.Fetch()
	for _, computer := range computers {
		log.Println("FETCH: ", computer.Name())
	}

	recv := make(chan *logmein.Computer)
	stop := make(chan struct{})
	go lmi.Watch(recv, stop, true)

	for {
		computer := <-recv
		fmt.Printf("EVENT  [%d]\n\tName: %s\n\tIP: %s\n\tStatus: %d\n\tTimestamp: %s\n", computer.HostID(), computer.Name(), computer.IPAddress(), computer.Status(), computer.Timestamp())
		if computer.Unchanged() {
			fmt.Println("\t• Unchanged")
		}
		if computer.IsDeleted() {
			fmt.Println("\t• IsDeleted")
		}
		if computer.IsNew() {
			fmt.Println("\t• IsNew")
		}
		if computer.IsChangedHostID() {
			fmt.Printf("\t• IsChangedHostID [%d -> %d]\n", computer.OldHostID(), computer.HostID())
		}
		if computer.IsChangedName() {
			fmt.Printf("\t• IsChangedName [%s -> %s]\n", computer.OldName(), computer.Name())
		}
		if computer.IsChangedIPAddress() {
			fmt.Printf("\t• IsChangedIPAddress [%s -> %s]\n", computer.OldIPAddress(), computer.IPAddress())
		}
		if computer.IsChangedStatus() {
			fmt.Printf("\t• IsChangedStatus [%d -> %d]\n", computer.OldStatus(), computer.Status())
		}
		fmt.Println("")
	}
}
