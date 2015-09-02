package logmein

import (
	"net"
	"sync"
	"time"
)

// StatusValue is the Computer status
type StatusValue int

// Status enum
const (
	StatusOffline StatusValue = 0
	StatusOnline              = 1
	StatusSession             = 4
)

type computerRecord struct {
	hostID    uint64
	timestamp time.Time
	name      string
	ipAddress net.IP
	status    StatusValue
}

// Computer is a LogMeIn computer
type Computer struct {
	changes ChangeSet
	now     *computerRecord
	prev    *computerRecord
	lock    sync.RWMutex
}

// FIXME: right now we have no way to notice that a computer has been removed, aside
//   from skating through the []Computer to find old timestamps.  We should be detecting
//   deleted Computers in some efficient manner.

func (lmi *LMI) processComputerRaw(record computerRecord, event chan<- *Computer) {
	lmi.lock.Lock()
	defer lmi.lock.Unlock()

	computer, ok := lmi.computers[record.hostID]
	if !ok {
		// a new Computer
		computer = &Computer{
			now:     &record,
			changes: record.Diff(nil),
		}
		lmi.computers[computer.now.hostID] = computer
		if event != nil {
			event <- computer
		}
	} else {
		// an existing Computer
		func() {
			computer.lock.Lock()
			defer computer.lock.Unlock()
			changes := record.Diff(computer.now)
			if changes != 0 {
				// something changed
				computer.changes = changes
				computer.prev = computer.now
				computer.now = &record
				if event != nil {
					event <- computer
				}
			}
		}()
	}
}

// ChangeSet ...
type ChangeSet uint

// ChangeType ...
type ChangeType uint

// ChangeTypes...
const (
	ChangedNone ChangeType = 0
	IsNewRecord ChangeType = 1 << iota
	IsDeletedRecord
	ChangedHostID // ??? this one probably shouldn't ever happen since we key off of this value.
	ChangedName
	ChangedIPAddress
	ChangedStatus
)

// Computers returns all computers
func (lmi *LMI) Computers() map[uint64]*Computer {
	return lmi.computers
}

// Diff returns a list of ChangeTypes based on a comparison of two computerRecords
// a = new record, b = previous record.  if b is nil, it's a new record.  if a is nil, maybe it's a deleted record?
func (a *computerRecord) Diff(b *computerRecord) ChangeSet {
	// note: we don't care about the timestamp here

	var changes ChangeSet
	if a == nil && b == nil {
		// why would this happen?
		return changes
	}
	selectAllChanges := false
	if a == nil {
		selectAllChanges = true
		changes |= ChangeSet(IsDeletedRecord)
	}
	if b == nil {
		selectAllChanges = true
		changes |= ChangeSet(IsNewRecord)
	}
	if selectAllChanges || a.hostID != b.hostID {
		changes |= ChangeSet(ChangedHostID)
	}
	if selectAllChanges || a.name != b.name {
		changes |= ChangeSet(ChangedName)
	}
	if selectAllChanges || !a.ipAddress.Equal(b.ipAddress) {
		changes |= ChangeSet(ChangedIPAddress)
	}
	if selectAllChanges || a.status != b.status {
		changes |= ChangeSet(ChangedStatus)
	}
	return changes
}

// GetChangeSet returns the ChangeSet between the current and prev version of this Computer
func (c *Computer) GetChangeSet() ChangeSet {
	return c.now.Diff(c.prev)
}

// Unchanged returns if this ChangeSet indicates no changes
func (cs ChangeSet) Unchanged() bool {
	return cs == 0
}

// Unchanged returns if this Computer is unchanged
// ???: when does this get set?  never?
func (c *Computer) Unchanged() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.changes.Unchanged()
}

// IsDeleted returns if this ChangeSet indicates a deleted record
func (cs ChangeSet) IsDeleted() bool {
	return cs&ChangeSet(IsDeletedRecord) == ChangeSet(IsDeletedRecord)
}

// IsDeleted returns if this Computer was deleted
func (c *Computer) IsDeleted() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.changes.IsDeleted()
}

// IsNew returns if this ChangeSet indicates a new record
func (cs ChangeSet) IsNew() bool {
	return cs&ChangeSet(IsNewRecord) == ChangeSet(IsNewRecord)
}

// IsNew returns if this Computer is new
func (c *Computer) IsNew() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.changes.IsNew()
}

// IsChangedHostID returns if this ChangeSet indicates a changed hostID
func (cs ChangeSet) IsChangedHostID() bool {
	return cs&ChangeSet(ChangedHostID) == ChangeSet(ChangedHostID)
}

// IsChangedHostID returns if this Computer's hostID field changed
func (c *Computer) IsChangedHostID() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.changes.IsChangedHostID()
}

// IsChangedName returns if this ChangeSet indicates a changed name
func (cs ChangeSet) IsChangedName() bool {
	return cs&ChangeSet(ChangedName) == ChangeSet(ChangedName)
}

// IsChangedName returns if this Computer's name field changed
func (c *Computer) IsChangedName() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.changes.IsChangedName()
}

// IsChangedIPAddress returns if this ChangeSet indicates a changed IP
func (cs ChangeSet) IsChangedIPAddress() bool {
	return cs&ChangeSet(ChangedIPAddress) == ChangeSet(ChangedIPAddress)
}

// IsChangedIPAddress returns if this Computer's IP field changed
func (c *Computer) IsChangedIPAddress() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.changes.IsChangedIPAddress()
}

// IsChangedStatus returns if this ChangeSet indicates a changed status
func (cs ChangeSet) IsChangedStatus() bool {
	return cs&ChangeSet(ChangedStatus) == ChangeSet(ChangedStatus)
}

// IsChangedStatus returns if this Computer's status field changed
func (c *Computer) IsChangedStatus() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.changes.IsChangedStatus()
}

// HostID returns the current HostID of the computer
func (c *Computer) HostID() uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.now.hostID
}

// OldHostID returns the former HostID of the computer
func (c *Computer) OldHostID() uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.prev == nil {
		return 0
	}
	return c.prev.hostID
}

// Name returns the current Name of the computer
func (c *Computer) Name() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.now.name
}

// OldName returns the former Name of the computer
func (c *Computer) OldName() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.prev == nil {
		return ""
	}
	return c.prev.name
}

// IPAddress returns the current IPAddress of the computer
func (c *Computer) IPAddress() net.IP {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.now.ipAddress
}

// OldIPAddress returns the former IPAddress of the computer
func (c *Computer) OldIPAddress() net.IP {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.prev == nil {
		return nil
	}
	return c.prev.ipAddress
}

// Status returns the current Status of the computer
func (c *Computer) Status() StatusValue {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.now.status
}

// OldStatus returns the former Status of the computer
func (c *Computer) OldStatus() StatusValue {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.prev == nil {
		return 0
	}
	return c.prev.status
}

// Timestamp returns the current Timestamp of the computer
func (c *Computer) Timestamp() time.Time {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.now.timestamp
}

// OldTimestamp returns the former Timestamp of the computer
func (c *Computer) OldTimestamp() time.Time {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.prev == nil {
		return time.Time{}
	}
	return c.prev.timestamp
}
