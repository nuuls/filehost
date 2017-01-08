package filehost

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

// JSONDB implements a simple Database that stores the
// data in a json file and caches it in memory
type JSONDB struct {
	Path string

	mu    sync.RWMutex
	index map[string]*FileInfo
}

// NewDB initializes a JSONDB
func NewDB(path string) *JSONDB {
	db := &JSONDB{
		Path:  path,
		index: map[string]*FileInfo{},
	}
	db.index = db.load()
	return db
}

// SaveFileInfo implements the Database interface
func (db *JSONDB) SaveFileInfo(info *FileInfo) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.index[info.Name] = info
	db.save()
}

// GetFileInfo implements the Database interface
func (db *JSONDB) GetFileInfo(name string) *FileInfo {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.index[name]
}

// save saves the index to disk, mu must be at least RLocked
func (db *JSONDB) save() {
	bs, err := json.MarshalIndent(db.index, "", "    ")
	if err != nil {
		log.WithError(err).Error("cannot marshal database index")
		return
	}
	file, err := os.Create(db.Path)
	if err != nil {
		log.WithError(err).Error("cannot create database file")
		return
	}
	_, err = file.Write(bs)
	if err != nil {
		log.WithError(err).Error("cannot write to database file")
	}
}

func (db *JSONDB) load() map[string]*FileInfo {
	index := map[string]*FileInfo{}
	bs, err := ioutil.ReadFile(db.Path)
	if err != nil {
		log.WithError(err).Error("cannot load database file")
		return index
	}
	err = json.Unmarshal(bs, &index)
	if err != nil {
		log.WithError(err).Error("failed to read database file")
		return index
	}
	return index
}
