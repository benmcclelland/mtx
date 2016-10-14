# mtx
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/benmcclelland/mtx) [![build](https://img.shields.io/travis/benmcclelland/mtx.svg?style=flat)](https://travis-ci.org/benmcclelland/mtx)

golang mtx wrapper library for scsi media changers

Tested with Centos6/RHEL6 version of mtx (mtx-1.3.12)

Example of loading and unloading a volume:

```go
// Initialize new library session using "/dev/sg0"
// as the media changer device
lib := mtx.NewLibrary("/dev/sg0")

// Calling status will parse the mtx output and
// fill out the MediaInfo struct
mi, err := lib.Status()
if err != nil {
	return err
}

// FindStorageVolume will look for a volume with
// a matching barcode in the storage slots
vol, err := mtx.FindStorageVolume("MyBarcode", mi)
if err != nil {
	return err
}

// Load the volume that we found into the first
// drive element, drive elements start with 0
err = lib.Load(vol, mi.Drives["0"])
if err != nil {
	return err
}

// Unload the volume from the drive when finished
err = lib.Unload(vol)
if err != nil {
	return err
}
```
