package logmein

import "time"

// MinimumInterval is the minimal interval for RSS feed retrieval
// You risk LogMeIn Inc's fury if you use an interval lower than one minute
var MinimumInterval = time.Minute

// Interval sets the RSS feed pull interval (or get the current interval if arg=0)
// If interval is < MinimumInterval, interval will be set to MinimumInterval
// Hint: if you want interval to be < than MinimumInterval, overwrite MinimumInterval with an appropriate lower value
func (lmi *LMI) Interval(interval time.Duration) time.Duration {
	// if we're not setting a new value, interval will be 0, which just gets the current value
	if interval != 0 {
		if interval < MinimumInterval {
			interval = MinimumInterval
		}

		// did it actually change?  if so, apply it.  otherwise don't waste time on it.
		if lmi.interval != interval {
			lmi.interval = interval
			lmi.intervalUpdateChan <- true
		}
	}

	return lmi.interval
}

// Watch runs the watch loop until the stop channel is closed
// This sends changes to Computers based on the last Fetch or last Watch loop iteration
// Your are not obligated to Fetch before Watch.  If you only Watch, all Computers and their fields will appear as new/changed
// startDelayed == no fetch until interval; !startDelayed == fetch immediately and then resume usual fetch interval
func (lmi *LMI) Watch(event chan<- *Computer, stop <-chan struct{}, startDelayed bool) {
	if !startDelayed {
		lmi.tick(event)
	}
	if lmi.interval < MinimumInterval {
		lmi.interval = MinimumInterval
	}
	ticker := time.NewTicker(lmi.interval)
	for {
		select {
		case <-stop:
			ticker.Stop()
			return
		case <-lmi.intervalUpdateChan:
			ticker.Stop()
			ticker = time.NewTicker(lmi.interval)
		case <-ticker.C:
			lmi.tick(event)
		}
	}
}

func (lmi *LMI) tick(event chan<- *Computer) {
	records := lmi.fetch(lmi.rssURL(), "")
	for _, record := range records {
		lmi.processComputerRaw(record, event)
	}
}
