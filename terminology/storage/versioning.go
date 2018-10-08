package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Descriptor provides a simple structure for file-backed datastore versioning
// and configuration.
type Descriptor struct {
	Version   float32
	StoreType string
	path      string
}

const (
	descriptorName = "sctdb.json"
)

// CreateOrOpenDescriptor either opens or creates a Descriptor at the specified
// path. If creating a new Descriptor the Version and Store type specified are
// used, otherwise the values are ignored
func CreateOrOpenDescriptor(path string, currentVersion float32, storeType string) (*Descriptor, error) {
	descriptorFilename := filepath.Join(path, descriptorName)
	if _, err := os.Stat(descriptorFilename); os.IsNotExist(err) {
		desc := &Descriptor{Version: currentVersion, StoreType: storeType, path: path}
		return desc, desc.Save()
	}
	data, err := ioutil.ReadFile(descriptorFilename)
	if err != nil {
		return nil, err
	}
	var desc Descriptor
	desc.path = path
	return &desc, json.Unmarshal(data, &desc)
}

// Save writes the Descriptor to the filesystem
func (d *Descriptor) Save() error {
	descriptorFilename := filepath.Join(d.path, descriptorName)
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(descriptorFilename, data, 0644)
}
