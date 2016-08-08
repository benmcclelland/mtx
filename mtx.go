package mtx

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// SlotType is the type of slot
type SlotType int

const (
	// Unknown slot type
	Unknown = iota
	// DataTransferElement is a drive slot
	DataTransferElement
	// StorageElement is a Storage slot
	StorageElement
	// ImportExport is a MailBox slot
	ImportExport
)

var (
	summaryRxp  = regexp.MustCompile(`\s*Storage Changer .*:(\d*) Drives, (\d*) Slots \( (\d*) Import/Export \)`)
	dteEmptyRxp = regexp.MustCompile(`Data Transfer Element (\d*):Empty`)
	dteFullRxp  = regexp.MustCompile(`Data Transfer Element (\d*):Full \(Storage Element (\d*) Loaded\):VolumeTag = (.*)`)
	seEmptyRxp  = regexp.MustCompile(`\s*Storage Element (\d*):Empty`)
	seFullRxp   = regexp.MustCompile(`\s*Storage Element (\d*):Full :VolumeTag=(.*)`)
	ieEmptyRxp  = regexp.MustCompile(`\s*Storage Element (\d*) IMPORT/EXPORT:Empty`)
	ieFullRxp   = regexp.MustCompile(`\s*Storage Element (\d*) IMPORT/EXPORT:Full :VolumeTag=(.*)`)
	clnRxp      = regexp.MustCompile(`(CLN.*)`)
)

var (
	// ErrSummary happens when our regex could not match the mtx status summary line
	ErrSummary = errors.New("unrecognized status summary string")
	// ErrSlotNotFound happens when we can't find the home slot for media
	ErrSlotNotFound = errors.New("unable to find requested slot for media")
	// ErrVolNotInDrive happens when trying to unload a volume not in a drive
	ErrVolNotInDrive = errors.New("media not in a drive")
	// ErrNoDrives happens when trying to load a volume in an available drive and
	// there are no empty drives
	ErrNoDrives = errors.New("no empty drives")
	// ErrInvalidDrive happens when we cant find the drive by ID
	ErrInvalidDrive = errors.New("unable to find drive")
	// ErrNoCLNMedia happens when we cant find any cleaning media in the library
	ErrNoCLNMedia = errors.New("no cleaning media found")
	// ErrNotInit happens when we try to access the Library MediaInfo before calling Status()
	ErrNotInit = errors.New("media info not initialized, please call Status() on this Library first")
	// ErrDrvNotAvail happens when trying to load media into a non-empty or unavaialble drive
	ErrDrvNotAvail = errors.New("requested media load in non empty or avaialble drive")
	// ErrNoHome happens when we can't figure out the home position for the volmue
	ErrNoHome = errors.New("no home position found for volume")
)

// Volume is a representation for the media in the Library
type Volume struct {
	// ID is the serial number for the media volume
	ID string
	// Home is the home storage slot for the media volume
	Home string
	// Drive is the string ID of the drive media is currently in
	// or "" if not in a drive
	Drive string
}

// Slot is a representation of each physical slot in the the Library
type Slot struct {
	// Type is the type of slot
	Type SlotType
	// ID is the slot identifier
	ID string
	// Volume is a pointer to the Volume currently in the slot
	// or nil if empty
	Vol *Volume
}

// DriveInfo is a map of drive IDs as strings to slot information
type DriveInfo map[string]Slot

// SlotInfo is a map of storage slot IDs as strings to slot information
type SlotInfo map[string]Slot

// MboxInfo is a map of mailbox IDs as strings to slot information
type MboxInfo map[string]Slot

// MediaInfo is a strucured representation of the state of the Library
type MediaInfo struct {
	// NumDrives is the total number of drives
	NumDrives int
	// NumSlots is the total number of storage slots
	NumSlots int
	// NumImportExport is the total number of import/export slots
	NumImportExport int
	// Drives is the drive representation
	Drives DriveInfo
	// Slots is the storage slot representation
	Slots SlotInfo
	// Mboxes is the Mbox representation
	Mboxes MboxInfo
}

// Library represents a single SCSI based media changer
type Library struct {
	// Device is the device file in use for this Library
	Device string
	// Command is the mtx command used for the Library
	Command string
	// Protects MediaInfo and command exec
	mu          sync.Mutex
	mi          MediaInfo
	initialized bool
}

// NewLibrary returns a Library for a given SCSI device path
func NewLibrary(device string) *Library {
	return &Library{Device: device, Command: "mtx"}
}

// NewLibraryCmd returns a Library for a given SCSI device path and mtx command
func NewLibraryCmd(device, cmd string) *Library {
	return &Library{Device: device, Command: cmd}
}

// Status returns a structured representation of the drives, slots,
// import/export, and media locations
func (l *Library) Status() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	result, err := mtxCmd(l.Command, l.Device, "status")
	if err != nil {
		return errors.Wrap(err, "status")
	}

	l.mi, err = parseStatus(bytes.NewReader(result))
	if err == nil {
		l.initialized = true
	}
	return errors.Wrap(err, "status")
}

func parseStatus(r io.Reader) (MediaInfo, error) {
	lscanner := bufio.NewScanner(r)
	var driveCnt, slotCnt, mboxCnt int
	if lscanner.Scan() {
		var err error
		line := lscanner.Text()
		match := summaryRxp.FindStringSubmatch(line)
		if len(match) != 4 {
			return MediaInfo{}, ErrSummary
		}
		driveCnt, err = strconv.Atoi(match[1])
		if err != nil {
			return MediaInfo{}, err
		}
		totslotCnt, err := strconv.Atoi(match[2])
		if err != nil {
			return MediaInfo{}, err
		}
		mboxCnt, err = strconv.Atoi(match[3])
		if err != nil {
			return MediaInfo{}, err
		}
		slotCnt = totslotCnt - mboxCnt

	}
	if err := lscanner.Err(); err != nil {
		return MediaInfo{}, err
	}

	dmap := make(map[string]Slot)
	smap := make(map[string]Slot)
	mmap := make(map[string]Slot)
	for lscanner.Scan() {
		line := lscanner.Text()
		// We are going to test these in the order of what we are
		// more likely to hit a match on for a larger library
		// to try to minimize the searching
		// Full storage slot
		match := seFullRxp.FindStringSubmatch(line)
		if match != nil {
			newVol := Volume{
				ID:    match[2],
				Home:  match[1],
				Drive: "",
			}
			smap[match[1]] = Slot{
				Type: StorageElement,
				ID:   match[1],
				Vol:  &newVol,
			}
			continue
		}
		// Empty storage slot
		match = seEmptyRxp.FindStringSubmatch(line)
		if match != nil {
			smap[match[1]] = Slot{
				Type: StorageElement,
				ID:   match[1],
				Vol:  nil,
			}
			continue
		}
		// Full drive slot
		match = dteFullRxp.FindStringSubmatch(line)
		if match != nil {
			newVol := Volume{
				ID:    match[3],
				Home:  match[2],
				Drive: match[1],
			}
			dmap[match[1]] = Slot{
				Type: DataTransferElement,
				ID:   match[1],
				Vol:  &newVol,
			}
			continue
		}
		// Empty drive slot
		match = dteEmptyRxp.FindStringSubmatch(line)
		if match != nil {
			dmap[match[1]] = Slot{
				Type: DataTransferElement,
				ID:   match[1],
				Vol:  nil,
			}
			continue
		}
		// Empty mailbox slot
		match = ieEmptyRxp.FindStringSubmatch(line)
		if match != nil {
			mmap[match[1]] = Slot{
				Type: ImportExport,
				ID:   match[1],
				Vol:  nil,
			}
			continue
		}
		// Full mailbox slot
		match = ieFullRxp.FindStringSubmatch(line)
		if match != nil {
			newVol := Volume{
				ID:    match[2],
				Home:  match[1],
				Drive: "",
			}
			mmap[match[1]] = Slot{
				Type: ImportExport,
				ID:   match[1],
				Vol:  &newVol,
			}
			continue
		}
	}
	if err := lscanner.Err(); err != nil {
		return MediaInfo{}, err
	}

	m := MediaInfo{
		NumDrives:       driveCnt,
		NumSlots:        slotCnt,
		NumImportExport: mboxCnt,
		Drives:          dmap,
		Slots:           smap,
		Mboxes:          mmap,
	}
	return m, nil

}

// Inventory tells the Library to (re)inventory all the media
// which usually involves a lot of robotic movement and barcode reading
func (l *Library) Inventory() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := mtxCmd(l.Command, l.Device, "inventory")
	return errors.Wrap(err, "inventory")
}

// Load will attempt to move media from a storage slot to a drive
func (l *Library) Load(slot, drive string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.mi.Drives[drive].Vol != nil {
		return ErrDrvNotAvail
	}
	_, err := mtxCmd(l.Command, l.Device, "load", slot, drive)
	if err == nil && l.initialized {
		d := l.mi.Drives[drive]
		l.mi.Drives[drive] = Slot{
			Type: d.Type,
			ID:   d.ID,
			Vol:  l.mi.Slots[slot].Vol,
		}
		s := l.mi.Slots[slot]
		l.mi.Slots[slot] = Slot{
			Type: s.Type,
			ID:   s.ID,
		}
	}
	return errors.Wrap(err, "load")
}

// LoadVol will attempt to move volume from home slot to an availble drive
// The slot returned if no error will be the drive that it got loaded in
func (l *Library) LoadVol(v Volume) (Slot, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.initialized == false {
		return Slot{}, ErrNotInit
	}
	edrives := GetEmptyDrives(l.mi)
	if len(edrives) == 0 {
		return Slot{}, ErrNoDrives
	}

	var ok bool
	if _, ok = l.mi.Drives[edrives[0]]; !ok {
		return Slot{}, ErrInvalidDrive
	}

	_, err := mtxCmd(l.Command, l.Device, "load", v.Home, edrives[0])
	if err == nil {
		l.mi.Slots[v.Home].Vol.Drive = edrives[0]
		d := l.mi.Drives[edrives[0]]
		l.mi.Drives[edrives[0]] = Slot{
			Type: d.Type,
			ID:   d.ID,
			Vol:  l.mi.Slots[v.Home].Vol,
		}
		s := l.mi.Slots[v.Home]
		l.mi.Slots[v.Home] = Slot{
			Type: s.Type,
			ID:   s.ID,
		}
	}
	return l.mi.Drives[edrives[0]], errors.Wrap(err, "loadvol")
}

// LoadCln will attempt to move randomized cleaning media to specified drive
func (l *Library) LoadCln(d Slot) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	clns := FindCleaningMedia(l.mi)
	if len(clns) == 0 {
		return ErrNoCLNMedia
	}

	// Pick random cleaning media to load balance them
	v := clns[rand.Intn(len(clns))]

	_, err := mtxCmd(l.Command, l.Device, "load", v.Home, d.ID)
	if err == nil && l.initialized {
		d := l.mi.Drives[d.ID]
		l.mi.Drives[d.ID] = Slot{
			Type: d.Type,
			ID:   d.ID,
			Vol:  l.mi.Slots[v.Home].Vol,
		}
		s := l.mi.Slots[v.Home]
		l.mi.Slots[v.Home] = Slot{
			Type: s.Type,
			ID:   s.ID,
		}
	}
	return errors.Wrap(err, "loadcln")
}

// Unload will attempt to move media from drive to a storage slot
func (l *Library) Unload(slot, drive string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := mtxCmd(l.Command, l.Device, "unload", slot, drive)
	if err == nil && l.initialized {
		s := l.mi.Slots[slot]
		l.mi.Slots[slot] = Slot{
			Type: s.Type,
			ID:   s.ID,
			Vol:  l.mi.Drives[drive].Vol,
		}
		d := l.mi.Drives[drive]
		l.mi.Drives[drive] = Slot{
			Type: d.Type,
			ID:   d.ID,
		}
	}
	return errors.Wrap(err, "unload")
}

// UnloadVol will attempt to move volume from drive to home slot
func (l *Library) UnloadVol(v Volume) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if v.Drive == "" {
		return ErrVolNotInDrive
	}
	if v.Home == "" {
		return ErrNoHome
	}

	_, err := mtxCmd(l.Command, l.Device, "unload", v.Home, v.Drive)
	if err == nil && l.initialized {
		s := l.mi.Slots[v.Home]
		l.mi.Slots[v.Home] = Slot{
			Type: s.Type,
			ID:   s.ID,
			Vol:  l.mi.Drives[v.Drive].Vol,
		}
		d := l.mi.Drives[v.Drive]
		l.mi.Drives[v.Drive] = Slot{
			Type: d.Type,
			ID:   d.ID,
		}
	}
	return errors.Wrap(err, "unloadvol")
}

// Transfer will attempt to move media from one stroage slot to another
func (l *Library) Transfer(slotA, slotB string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := mtxCmd(l.Command, l.Device, "transfer", slotA, slotB)
	if err == nil && l.initialized {
		s := l.mi.Slots[slotB]
		l.mi.Slots[slotB] = Slot{
			Type: s.Type,
			ID:   s.ID,
			Vol:  l.mi.Slots[slotA].Vol,
		}
		s = l.mi.Slots[slotA]
		l.mi.Slots[slotA] = Slot{
			Type: s.Type,
			ID:   s.ID,
		}
	}
	return errors.Wrap(err, "transfer")
}

// Info returns a copy of the MediaInfo for the library
// Will be an empty struct if Status() hasn't been called yet
func (l Library) Info() MediaInfo {
	return l.mi
}

// String representation for a Library is the device path
func (l Library) String() string {
	return l.Device
}

// GetDriveByID returns the drive Slot for a given string ID
func GetDriveByID(id string, m MediaInfo) (Slot, error) {
	if s, ok := m.Drives[id]; ok {
		return s, nil
	}
	return Slot{}, ErrSlotNotFound
}

// GetSlotByID returns the storage Slot for a given string ID
func GetSlotByID(id string, m MediaInfo) (Slot, error) {
	if s, ok := m.Slots[id]; ok {
		return s, nil
	}
	return Slot{}, ErrSlotNotFound
}

// GetMboxByID returns the mailbox Slot for a given string ID
func GetMboxByID(id string, m MediaInfo) (Slot, error) {
	if s, ok := m.Mboxes[id]; ok {
		return s, nil
	}
	return Slot{}, ErrSlotNotFound
}

// GetEmptyDrives returns a slice of string drive IDs with no volumes set
func GetEmptyDrives(m MediaInfo) []string {
	var result []string
	for _, value := range m.Drives {
		if value.Vol == nil {
			result = append(result, value.ID)
		}
	}
	return result
}

// FindCleaningMedia returns a slice of Volumes that have serial
// numbers begining with CLN that are not currently in a drive
func FindCleaningMedia(m MediaInfo) []Volume {
	var result []Volume
	for _, value := range m.Slots {
		if value.Vol != nil {
			if clnRxp.MatchString(value.Vol.ID) {
				result = append(result, *value.Vol)
			}
		}
	}
	return result
}

// FindHomeSlot returns the home slot for the volume
func FindHomeSlot(v Volume, m MediaInfo) (Slot, error) {
	if s, ok := m.Slots[v.Home]; ok {
		return s, nil
	}
	if s, ok := m.Mboxes[v.Home]; ok {
		return s, nil
	}
	return Slot{}, ErrSlotNotFound
}

// FindHomeID returns the string ID for the home slot for the volume
func FindHomeID(v Volume) string {
	return v.Home
}

func mtxCmd(mtxcmd, dev string, args ...string) ([]byte, error) {
	cmdargs := append([]string{"-f", dev}, args...)
	cmd := exec.Command(mtxcmd, cmdargs...)
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		err = errors.Wrap(err, "mtx command")
		return []byte{}, err
	}
	if err := cmd.Start(); err != nil {
		err = errors.Wrap(err, "mtx command")
		return []byte{}, err
	}
	cmdout, err := ioutil.ReadAll(stdout)
	if err != nil {
		err = errors.Wrap(err, "mtx command")
		return []byte{}, err
	}
	cmderr, err := ioutil.ReadAll(stderr)
	if err != nil {
		err = errors.Wrap(err, "mtx command")
		return []byte{}, err
	}
	if err := cmd.Wait(); err != nil {
		err = errors.Wrap(err, "mtx command")
		err = errors.Wrap(err, strings.TrimSuffix(string(cmderr), "\n"))
		return []byte{}, err
	}
	return cmdout, nil
}
