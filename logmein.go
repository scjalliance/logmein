package logmein

import (
	"fmt"
	"sync"
	"time"
)

// LMI is the LogMeIn profile watcher instance
type LMI struct {
	profileID          uint64
	secretKey          string
	computers          map[uint64]*Computer
	interval           time.Duration
	intervalUpdateChan chan bool
	lock               sync.RWMutex
}

// NewLMI returns an LMI
func NewLMI(profileID uint64, secretKey string) *LMI {
	lmi := &LMI{
		profileID:          profileID,
		secretKey:          secretKey,
		computers:          make(map[uint64]*Computer),
		intervalUpdateChan: make(chan bool),
	}
	return lmi
}

func (lmi *LMI) rssURL() string {
	return fmt.Sprintf("https://secure.logmein.com/usershortcut.asp?key=%s&profileid=%d&showoffline=1&lmiextensions=1", lmi.secretKey, lmi.profileID)
}
