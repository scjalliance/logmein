package logmein

import (
	"log"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/jteeuwen/go-pkg-xmlx"
)

var matchHostID = regexp.MustCompile(`[\?&]hostid=(\d+)(&.*)?$`)

// Fetch gets the most current data from the LMI RSS feed and replaces the condition that Watch is based on
func (lmi *LMI) Fetch() map[uint64]*Computer {
	records := lmi.fetch(lmi.rssURL(), "")
	for _, record := range records {
		lmi.processComputerRaw(record, nil)
	}
	return lmi.computers
}

func (lmi *LMI) fetch(uri string, file string) []computerRecord {
	var records []computerRecord

	doc := xmlx.New()
	if len(uri) > 0 {
		if err := doc.LoadUri(uri, nil); err != nil {
			log.Println("ERROR: ", err)
			return nil
		}
	} else if len(file) > 0 {
		if err := doc.LoadFile(file, nil); err != nil {
			log.Println("ERROR: ", err)
			return nil
		}
	} else {
		return nil
	}

	rssNode := doc.SelectNode("", "rss")
	timestamp, err := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", rssNode.S("", "lastBuildDate"))
	if err != nil {
		// FIXME: is this the best thing to do?
		timestamp = time.Now()
	}
	items := rssNode.SelectNodes("", "item")
	for _, item := range items {
		info := item.SelectNode("", "lmihostinfo")
		record := computerRecord{
			timestamp: timestamp,
			name:      info.S("", "description"),
			ipAddress: net.ParseIP(info.S("", "ip")),
			status:    StatusValue(info.I("", "status")),
		}
		hostMatches := matchHostID.FindStringSubmatch(info.S("", "link"))
		if len(hostMatches) < 2 {
			// FIXME: we should probably complain?
			continue
		}
		hostID, err := strconv.ParseUint(hostMatches[1], 10, 64)
		if err != nil {
			// FIXME: we should probably complain?
			continue
		}
		record.hostID = hostID

		records = append(records, record)
	}

	return records
}
