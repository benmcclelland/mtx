package mtx

import "testing"

func TestStatus(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("Status(): %v", err)
	}
	if m.NumDrives != int(2) {
		t.Errorf("Status(): NumDrives expected 2, got %v", m.NumDrives)
	}
	if m.NumSlots != int(4) {
		t.Errorf("Status(): NumSlots expected 4, got %v", m.NumSlots)
	}
	if m.NumImportExport != int(2) {
		t.Errorf("Status(): NumImportExport expected 2, got %v", m.NumImportExport)
	}
	// Drive 0
	if m.Drives["0"].Type != DataTransferElement {
		t.Errorf("Status(): Drives[\"0\"].Type expected DataTransferElement, got %v",
			m.Drives["0"].Type)
	}
	if m.Drives["0"].ID != "0" {
		t.Errorf("Status(): Drives[\"0\"].ID expected \"0\", got \"%v\"",
			m.Drives["0"].ID)
	}
	if m.Drives["0"].Vol == nil {
		t.Errorf("Status(): Drives[\"0\"] unexpected empty slot")
	}
	if m.Drives["0"].Vol.ID != "M00001L6" {
		t.Errorf("Status(): Drives[\"0\"].Vol.ID expected M00001L6, got %v",
			m.Drives["0"].Vol.ID)
	}
	if m.Drives["0"].Vol.Home != "1" {
		t.Errorf("Status(): Drives[\"0\"].Vol.ID expected \"1\", got \"%v\"",
			m.Drives["0"].Vol.Home)
	}
	if m.Drives["0"].Vol.Drive != "0" {
		t.Errorf("Status(): Drives[\"0\"].Vol.Drive expected \"0\", got \"%v\"",
			m.Drives["0"].Vol.Drive)
	}
	// Drive 1
	if m.Drives["1"].Type != DataTransferElement {
		t.Errorf("Status(): Drives[\"1\"].Type expected DataTransferElement, got %v",
			m.Drives["1"].Type)
	}
	if m.Drives["1"].ID != "1" {
		t.Errorf("Status(): Drives[\"1\"].ID expected \"1\", got \"%v\"",
			m.Drives["1"].ID)
	}
	if m.Drives["1"].Vol != nil {
		t.Errorf("Status(): Drives[\"1\"].Vol expected nil, got %v", m.Drives["1"].Vol)
	}
	// Slot 1
	if m.Slots["1"].Type != StorageElement {
		t.Errorf("Status(): Slots[\"1\"].Type expected StorageElement, got %v",
			m.Slots["1"].Type)
	}
	if m.Slots["1"].ID != "1" {
		t.Errorf("Status(): Slots[\"1\"].ID expected \"1\", got \"%v\"",
			m.Slots["1"].ID)
	}
	if m.Slots["1"].Vol != nil {
		t.Errorf("Status(): Slots[\"1\"].Vol expected nil, got %v", m.Slots["1"].Vol)
	}
	// Slot 2
	if m.Slots["2"].Type != StorageElement {
		t.Errorf("Status(): Slots[\"2\"].Type expected StorageElement, got %v",
			m.Slots["2"].Type)
	}
	if m.Slots["2"].ID != "2" {
		t.Errorf("Status(): Slots[\"2\"].ID expected \"2\", got \"%v\"",
			m.Slots["2"].ID)
	}
	if m.Slots["2"].Vol != nil {
		t.Errorf("Status(): Slots[\"2\"].Vol expected nil, got %v", m.Slots["2"].Vol)
	}
	// Slot 3
	if m.Slots["3"].Type != StorageElement {
		t.Errorf("Status(): Slots[\"3\"].Type expected StorageElement, got %v",
			m.Slots["3"].Type)
	}
	if m.Slots["3"].ID != "3" {
		t.Errorf("Status(): Slots[\"3\"].ID expected \"3\", got \"%v\"",
			m.Slots["3"].ID)
	}
	if m.Slots["3"].Vol == nil {
		t.Errorf("Status(): Slots[\"3\"] unexpected empty slot")
	}
	if m.Slots["3"].Vol.ID != "M00003L6" {
		t.Errorf("Status(): Slots[\"3\"].Vol.ID expected M00003L6, got %v",
			m.Slots["3"].Vol.ID)
	}
	if m.Slots["3"].Vol.Home != "3" {
		t.Errorf("Status(): Slots[\"3\"].Vol.Home expected \"3\", got \"%v\"",
			m.Slots["3"].Vol.Home)
	}
	if m.Slots["3"].Vol.Drive != "" {
		t.Errorf("Status(): Slots[\"3\"].Vol.Drive expected \"\", got \"%v\"",
			m.Slots["3"].Vol.Drive)
	}
	// Slot 4
	if m.Slots["4"].Type != StorageElement {
		t.Errorf("Status(): Slots[\"4\"].Type expected StorageElement, got %v",
			m.Slots["4"].Type)
	}
	if m.Slots["4"].ID != "4" {
		t.Errorf("Status(): Slots[\"4\"].ID expected \"4\", got \"%v\"",
			m.Slots["4"].ID)
	}
	if m.Slots["4"].Vol == nil {
		t.Errorf("Status(): Slots[\"4\"] unexpected empty slot")
	}
	if m.Slots["4"].Vol.ID != "CLN004L6" {
		t.Errorf("Status(): Slots[\"4\"].Vol.ID expected CLN004L6, got %v",
			m.Slots["4"].Vol.ID)
	}
	if m.Slots["4"].Vol.Home != "4" {
		t.Errorf("Status(): Slots[\"4\"].Vol.Home expected \"4\", got \"%v\"",
			m.Slots["4"].Vol.Home)
	}
	if m.Slots["4"].Vol.Drive != "" {
		t.Errorf("Status(): Slots[\"4\"].Vol.Drive expected \"\", got \"%v\"",
			m.Slots["4"].Vol.Drive)
	}
	// (Mailbox) Slot 5
	if m.Mboxes["5"].Type != ImportExport {
		t.Errorf("Status(): Mboxes[\"5\"].Type expected ImportExport, got %v",
			m.Mboxes["5"].Type)
	}
	if m.Mboxes["5"].ID != "5" {
		t.Errorf("Status(): Mboxes[\"5\"].ID expected \"5\", got \"%v\"",
			m.Mboxes["5"].ID)
	}
	if m.Mboxes["5"].Vol == nil {
		t.Errorf("Status(): Mboxes[\"5\"] unexpected empty slot")
	}
	if m.Mboxes["5"].Vol.ID != "M00002L6" {
		t.Errorf("Status(): Mboxes[\"5\"].Vol.ID expected M00002L6, got %v",
			m.Mboxes["5"].Vol.ID)
	}
	if m.Mboxes["5"].Vol.Home != "5" {
		t.Errorf("Status(): Mboxes[\"5\"].Vol.Home expected \"5\", got \"%v\"",
			m.Mboxes["5"].Vol.Home)
	}
	if m.Mboxes["5"].Vol.Drive != "" {
		t.Errorf("Status(): Mboxes[\"5\"].Vol.Drive expected \"\", got \"%v\"",
			m.Mboxes["5"].Vol.Drive)
	}
	// (Mailbox) Slot 6
	if m.Mboxes["6"].Type != ImportExport {
		t.Errorf("Status(): Mboxes[\"6\"].Type expected ImportExport, got %v",
			m.Mboxes["6"].Type)
	}
	if m.Mboxes["6"].ID != "6" {
		t.Errorf("Status(): Mboxes[\"6\"].ID expected \"6\", got \"%v\"",
			m.Mboxes["6"].ID)
	}
	if m.Mboxes["6"].Vol != nil {
		t.Errorf("Status(): Mboxes[\"6\"].Vol expected nil, got %v", m.Mboxes["6"].Vol)
	}
}

func TestInventory(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	err := lib.Inventory()
	if err != nil {
		t.Errorf("Inventory: %v", err)
	}
}

func TestLoad(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("Load: Status(): %v", err)
	}
	err = lib.Load(lib.mi.Slots["3"].Vol, lib.mi.Drives["1"])
	if err != nil {
		t.Errorf("Load: %v", err)
	}
	// check changes to m
	if m.Slots["3"].Vol != nil {
		t.Errorf("Load: expected to empty slot 3 (nil Vol), got %v",
			m.Slots["3"].Vol.ID)
	}
	if m.Drives["1"].Vol == nil {
		t.Error("Load: expected non-empty Drive, got nil Vol")
	}
	if m.Drives["1"].Vol.ID != "M00003L6" {
		t.Errorf("Load: expected Vol ID in drive expected M00003L6, got %v",
			m.Drives["1"].Vol.ID)
	}
	if m.Drives["1"].Vol.Drive != "1" {
		t.Errorf("Load: expected Drive ID 1 for Vol M00003L6, got %v", m.Drives["1"].Vol.Drive)
	}
}

func TestLoadCln(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("LoadCln: Status(): %v", err)
	}
	err = lib.LoadCln(m.Drives["1"])
	if err != nil {
		t.Errorf("LoadCln: LoadCln(): %v", err)
	}
	if m.Slots["4"].Vol != nil {
		t.Errorf("LoadCln: expected to empty slot 4 (nil Vol), got %v",
			m.Slots["4"].Vol.ID)
	}
	if m.Drives["1"].Vol == nil {
		t.Error("LoadCln: expected non-empty Drive, got nil Vol")
	}
	if m.Drives["1"].Vol.ID != "CLN004L6" {
		t.Errorf("LoadCln: expected Vol ID in drive expected CLN004L6, got %v",
			m.Drives["1"].Vol.ID)
	}
	if m.Drives["1"].Vol.Drive != "1" {
		t.Errorf("LoadCln: expected Drive ID 1 for Vol M00003L6, got %v", m.Drives["1"].Vol.Drive)
	}
}

func TestUnLoad(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("Unload: Status(): %v", err)
	}
	err = lib.Unload(lib.mi.Drives["0"].Vol)
	if err != nil {
		t.Errorf("Unload: %v", err)
	}
	// check changes to m
	if m.Slots["1"].Vol == nil {
		t.Errorf("Unload: expected non-empty slot 1")
	}
	if m.Slots["1"].Vol.ID != "M00001L6" {
		t.Errorf("Unload: Vol ID in slot expected M00001L6, got %v",
			m.Slots["1"].Vol.ID)
	}
	if m.Slots["1"].Vol.Drive != "" {
		t.Errorf("Unload: Vol Drive in slot expected empty string, got %v",
			m.Slots["1"].Vol.Drive)
	}
	if m.Drives["0"].Vol != nil {
		t.Errorf("Unload: expected empty Drive, got Vol %v", m.Drives["0"].Vol.ID)
	}
}

func TestTransfer(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("Unload: Status(): %v", err)
	}
	err = lib.Transfer(lib.mi.Slots["3"].Vol, lib.mi.Slots["2"])
	if err != nil {
		t.Errorf("Unload: %v", err)
	}
	// check changes to m
	if m.Slots["2"].Vol == nil {
		t.Errorf("Transfer: expected non-empty slot 1")
	}
	if m.Slots["2"].Vol.ID != "M00003L6" {
		t.Errorf("Transfer: Vol ID in slot expected M00003L6, got %v", m.Slots["2"].Vol.ID)
	}
	if m.Slots["2"].Vol.Drive != "" {
		t.Errorf("Transfer: Vol Drive in slot expected empty string, got %v", m.Slots["2"].Vol.Drive)
	}
	if m.Slots["3"].Vol != nil {
		t.Errorf("Unload: expected empty Slot, got Vol %v\n%+v",
			m.Slots["3"].Vol.ID, m.Slots)
	}
}

func TestStatusFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmockerr")
	_, err := lib.Status()
	if err == nil {
		t.Errorf("Status(): expected error, got success")
	}
}

func TestInventoryFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmockerr")
	err := lib.Inventory()
	if err == nil {
		t.Errorf("Inventory(): expected error, got success")
	}
}

func TestLoadFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmockerr")
	err := lib.Load(&Volume{ID: "ABC", Home: "1"},
		Slot{Type: DataTransferElement, ID: "0"})
	if err == nil {
		t.Errorf("Load(): expected error, got success")
	}
}

func TestLoadClnFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmockerr")
	err := lib.LoadCln(Slot{Type: DataTransferElement, ID: "0", Vol: nil})
	if err == nil {
		t.Errorf("LoadCln(): expected error, got success")
	}
}

func TestUnloadFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmockerr")
	err := lib.Unload(&Volume{ID: "ABC", Home: "1", Drive: "0"})
	if err == nil {
		t.Errorf("Unload(): expected error, got success")
	}
}

func TestTransferFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmockerr")
	err := lib.Transfer(&Volume{ID: "ABC", Home: "1"},
		Slot{Type: StorageElement, ID: "2"})
	if err == nil {
		t.Errorf("Transfer(): expected error, got success")
	}
}

func TestGetDriveByID(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("GetDriveByID: Status(): %v", err)
	}
	d, err := GetDriveByID("1", *m)
	if err != nil {
		t.Errorf("GetDriveByID(): %v", err)
	}
	if d.ID != "1" {
		t.Errorf("drive ID expected \"1\", got \"%v\"", d.ID)
	}
}

func TestGetDriveByIDFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("GetDriveByID: Status(): %v", err)
	}
	_, err = GetDriveByID("2", *m)
	if err == nil {
		t.Errorf("GetDriveByID(): expected error, got nil")
	}
}

func TestGetSlotByID(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("GetSlotByID: Status(): %v", err)
	}
	s, err := GetSlotByID("1", *m)
	if err != nil {
		t.Errorf("GetSlotByID(): %v", err)
	}
	if s.ID != "1" {
		t.Errorf("slot ID expected \"1\", got  \"%v\"", s.ID)
	}
}

func TestGetSlotByIDFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("GetSlotByID: Status(): %v", err)
	}
	_, err = GetSlotByID("10", *m)
	if err == nil {
		t.Errorf("GetSlotByID(): expected error, got nil")
	}
}

func TestGetMboxByID(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("GetMboxByID: Status(): %v", err)
	}
	s, err := GetMboxByID("5", *m)
	if err != nil {
		t.Errorf("GetMboxByID(): %v", err)
	}
	if s.ID != "5" {
		t.Errorf("mbox ID expected \"5\", got  \"%v\"", s.ID)
	}
}

func TestGetMboxByIDFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("GetMboxByID: Status(): %v", err)
	}
	_, err = GetMboxByID("10", *m)
	if err == nil {
		t.Errorf("GetMboxByID(): expected error, got nil")
	}
}

func TestFindHomeSlot(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("FindHomeSlot: Status(): %v", err)
	}
	s, err := FindHomeSlot(*m.Slots["3"].Vol, *m)
	if err != nil {
		t.Errorf("FindHomeSlot(): %v", err)
	}
	if s.ID != "3" {
		t.Errorf("home slot expected \"3\", got \"%v\"", s.ID)
	}
}

func TestFindHomeSlotFail(t *testing.T) {
	lib := NewLibraryCmd("/dev/sga", "./mtxmock")
	m, err := lib.Status()
	if err != nil {
		t.Errorf("FindHomeSlot: Status(): %v", err)
	}
	_, err = FindHomeSlot(Volume{ID: "abc", Home: "100", Drive: ""}, *m)
	if err == nil {
		t.Errorf("FindHomeSlot(): expected error, but got nil")
	}
}
