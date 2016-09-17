/*
Package mtx is a Go library for interacting with SCSI media changers.
It wraps the mtx executable and parses the output.  The mtx executable
is readily available in most distros.
*/
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
			return MediaInfo{}, errors.Errorf("no summary output found")
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
		// more likely to hit a match on to try to minimize the
		// false matches in a large library
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

// Load will attempt to load volume into specified drive
func (l *Library) Load(vol *Volume, drive Slot) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.mi.Drives[drive.ID].Vol != nil {
		return errors.Errorf("attempting to load vol %v into non-epmty drive %v", vol.ID, drive.ID)
	}

	if vol.Drive != "" {
		return errors.Errorf("attempting to load vol %v that is already in drive %v", vol.ID, vol.Drive)
	}
	_, err := mtxCmd(l.Command, l.Device, "load", vol.Home, drive.ID)
	if err == nil && l.initialized {
		d := l.mi.Drives[drive.ID]
		l.mi.Drives[drive.ID] = Slot{
			Type: d.Type,
			ID:   d.ID,
			Vol:  vol,
		}
		s := l.mi.Slots[vol.Home]
		l.mi.Slots[vol.Home] = Slot{
			Type: s.Type,
			ID:   s.ID,
		}
		vol.Drive = drive.ID
	}
	return errors.Wrap(err, "load")
}

// LoadCln will attempt to move a randomized cleaning media to specified drive
func (l *Library) LoadCln(d Slot) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	clns := FindCleaningMedia(l.mi)
	if len(clns) == 0 {
		return errors.Errorf("no cleaning media avaiable")
	}

	// Pick random cleaning media to load balance them
	v := clns[rand.Intn(len(clns))]

	_, err := mtxCmd(l.Command, l.Device, "load", v.Home, d.ID)
	if err == nil && l.initialized {
		d := l.mi.Drives[d.ID]
		l.mi.Slots[v.Home].Vol.Drive = d.ID
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

// Unload will attempt to move volume from drive to home slot
func (l *Library) Unload(vol *Volume) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if vol.Drive == "" {
		return errors.Errorf("attmepting to unload volume %v not currently in drive", vol.ID)
	}
	if vol.Home == "" {
		return errors.Errorf("no home slot found for volume %v, can't unlaod", vol.ID)
	}

	_, err := mtxCmd(l.Command, l.Device, "unload", vol.Home, vol.Drive)
	if err == nil && l.initialized {
		s := l.mi.Slots[vol.Home]
		l.mi.Slots[vol.Home] = Slot{
			Type: s.Type,
			ID:   s.ID,
			Vol:  vol,
		}
		d := l.mi.Drives[vol.Drive]
		l.mi.Drives[vol.Drive] = Slot{
			Type: d.Type,
			ID:   d.ID,
		}
		vol.Drive = ""
	}
	return errors.Wrap(err, "unloadvol")
}

// Transfer will attempt to move volume to specified slot
func (l *Library) Transfer(vol *Volume, slot Slot) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := mtxCmd(l.Command, l.Device, "transfer", vol.ID, slot.ID)
	if err == nil && l.initialized {
		s := l.mi.Slots[slot.ID]
		l.mi.Slots[slot.ID] = Slot{
			Type: s.Type,
			ID:   s.ID,
			Vol:  vol,
		}
		s = l.mi.Slots[vol.Home]
		l.mi.Slots[vol.Home] = Slot{
			Type: s.Type,
			ID:   s.ID,
		}
		vol.Home = slot.ID
	}
	return errors.Wrap(err, "transfer")
}

// Info returns a copy of the MediaInfo for the library
// Will be an empty struct if Status() hasn't been called yet
func (l *Library) Info() MediaInfo {
	return l.mi
}

// String representation for a Library is the device path
func (l *Library) String() string {
	return l.Device
}

// GetDriveByID returns the drive Slot for a given string ID
func GetDriveByID(id string, m MediaInfo) (Slot, error) {
	if s, ok := m.Drives[id]; ok {
		return s, nil
	}
	return Slot{}, errors.Errorf("no slot found for id %v", id)
}

// GetSlotByID returns the storage Slot for a given string ID
func GetSlotByID(id string, m MediaInfo) (Slot, error) {
	if s, ok := m.Slots[id]; ok {
		return s, nil
	}
	return Slot{}, errors.Errorf("no slot found for id %v", id)
}

// GetMboxByID returns the mailbox Slot for a given string ID
func GetMboxByID(id string, m MediaInfo) (Slot, error) {
	if s, ok := m.Mboxes[id]; ok {
		return s, nil
	}
	return Slot{}, errors.Errorf("no slot found for id %v", id)
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
func FindHomeSlot(vol Volume, mi MediaInfo) (Slot, error) {
	if s, ok := mi.Slots[vol.Home]; ok {
		return s, nil
	}
	if s, ok := mi.Mboxes[vol.Home]; ok {
		return s, nil
	}
	return Slot{}, errors.Errorf("no home slot found for volume %v", vol.ID)
}

func mtxCmd(mtxcmd, dev string, args ...string) ([]byte, error) {
	cmdargs := append([]string{"-f", dev}, args...)
	cmd := exec.Command(mtxcmd, cmdargs...)
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		err = errors.Wrap(err, "mtx command setup stderr pipe")
		return []byte{}, err
	}
	if err := cmd.Start(); err != nil {
		err = errors.Wrap(err, "mtx start command")
		return []byte{}, err
	}
	cmdout, err := ioutil.ReadAll(stdout)
	if err != nil {
		err = errors.Wrap(err, "mtx read stdout output")
		return []byte{}, err
	}
	cmderr, err := ioutil.ReadAll(stderr)
	if err != nil {
		err = errors.Wrap(err, "mtx read stderr output")
		return []byte{}, err
	}
	if err := cmd.Wait(); err != nil {
		err = errors.Wrap(err, "mtx wait command")
		err = errors.Wrap(err, strings.TrimSuffix(string(cmderr), "\n"))
		return []byte{}, err
	}
	return cmdout, nil
}
