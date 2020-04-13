package kvbase

import (
	"encoding/json"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"strings"
)

// LevelBackend acts as a wrapper around a Backend interface
type LevelBackend struct {
	Backend
	Connection *leveldb.DB
	Source     string
}

// NewLevelDB initialises a new database using the LevelDB driver
func NewLevelDB(source string) (Backend, error) {
	if source == "" {
		source = "data"
	}

	db, err := leveldb.OpenFile(source, nil)
	if err != nil {
		return nil, err
	}

	database := LevelBackend{
		Connection: db,
		Source:     source,
	}

	return &database, nil
}

// Count returns the total number of records inside of the provided bucket
func (database *LevelBackend) Count(bucket string) (int, error) {
	db := database.Connection
	counter := 0

	iter := db.NewIterator(util.BytesPrefix([]byte(bucket+"_")), nil)
	for iter.Next() {
		counter++
	}
	iter.Release()

	if err := iter.Error(); err != nil {
		return 0, err
	}

	return counter, nil
}

// Create inserts a record into the backend
func (database *LevelBackend) Create(bucket string, key string, model interface{}) error {
	db := database.Connection

	if _, err := db.Get([]byte(bucket+"_"+key), nil); err == nil {
		return errors.New("key already exists")
	}

	data, err := json.Marshal(&model)
	if err != nil {
		return err
	}

	if err := db.Put([]byte(bucket+"_"+key), data, nil); err != nil {
		return err
	}

	return nil
}

// Delete removes a record from the backend
func (database *LevelBackend) Delete(bucket string, key string) error {
	db := database.Connection

	if _, err := db.Get([]byte(bucket+"_"+key), nil); err != nil {
		return err
	}

	if err := db.Delete([]byte(bucket+"_"+key), nil); err != nil {
		return err
	}

	return nil
}

// Drop deletes a bucket (and all of its contents) from the backend
func (database *LevelBackend) Drop(bucket string) error {
	db := database.Connection

	iter := db.NewIterator(util.BytesPrefix([]byte(bucket+"_")), nil)
	for iter.Next() {
		if err := db.Delete(iter.Key(), nil); err != nil {
			return err
		}
	}
	iter.Release()

	if err := iter.Error(); err != nil {
		return err
	}

	return nil
}

// Get returns all records inside of the provided bucket
func (database *LevelBackend) Get(bucket string, model interface{}) (*map[string]interface{}, error) {
	db := database.Connection
	results := make(map[string]interface{})

	iter := db.NewIterator(util.BytesPrefix([]byte(bucket+"_")), nil)
	for iter.Next() {
		key := strings.TrimPrefix(string(iter.Key()), bucket+"_")

		if err := json.Unmarshal(iter.Value(), &model); err != nil {
			return nil, err
		}

		results[key] = model
	}
	iter.Release()

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return &results, nil
}

// Read returns a single struct from the provided bucket, using the provided key
func (database *LevelBackend) Read(bucket string, key string, model interface{}) error {
	db := database.Connection

	data, err := db.Get([]byte(bucket+"_"+key), nil)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &model)
}

// Update modifies an existing record from the backend, inside of the provided bucket, using the provided key
func (database *LevelBackend) Update(bucket string, key string, model interface{}) error {
	db := database.Connection

	if _, err := db.Get([]byte(bucket+"_"+key), nil); err != nil {
		return err
	}

	data, err := json.Marshal(&model)
	if err != nil {
		return err
	}

	if err := db.Put([]byte(bucket+"_"+key), data, nil); err != nil {
		return err
	}

	return nil
}
