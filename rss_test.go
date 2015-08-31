package logmein

import (
	"fmt"
	"testing"
	"time"
)

// FIXME: write useful tests...

func TestFileParse(t *testing.T) {
	lmi := NewLMI(0, "")

	stop := make(chan struct{})
	event := make(chan *Computer)
	go func() {
		for {
			select {
			case <-stop:
				return
			case computer := <-event:
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
	}()

	for _, record := range lmi.fetch("", "test1.xml") {
		lmi.processComputerRaw(record, event)
	}

	time.Sleep(time.Second * 2)

	for _, record := range lmi.fetch("", "test2.xml") {
		lmi.processComputerRaw(record, event)
	}

	close(stop)
}
