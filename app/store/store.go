package store

import "time"

type DataValue struct {
	value      string
	expiration *time.Time
}

type DataStore struct {
	items map[string]*DataValue
}

func NewDataStore() *DataStore {
	return &DataStore{items: make(map[string]*DataValue)}
}

func (d *DataStore) Set(key string, val string) {
	d.SetWithExpiration(key, val, nil)
}

func (d *DataStore) SetWithExpiration(key string, val string, expiration *time.Time) {
	d.items[key] = &DataValue{value: val, expiration: expiration}
}

func (d *DataStore) Get(key string) *string {
	val, found := d.items[key]
	if !found {
		return nil
	}
	if val.expiration != nil && time.Now().After(*val.expiration) {
		delete(d.items, key)
		return nil
	}
	return &val.value
}
