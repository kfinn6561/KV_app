package KVStore

import (
	"encoding/json"
	"fmt"
	"time"
)

//Data stores all information relevant to a key
type Data struct {
	owner        string
	value        string
	reads        int
	writes       int
	lastAccessed time.Time
}

//Data.isAuthorised checks if the user is authorised to access that data
func (d *Data) isAuthorised(user string) bool {
	return user == d.owner || user == adminUser
}

//Data.getValue returns the current value of the data and also updates the access timestamp
func (d *Data) getValue() string {
	d.reads++
	d.lastAccessed = time.Now()
	return d.value
}

//Data.setValue sets the current value of the data and also updates the access timestamp
func (d *Data) setValue(value string) {
	d.writes++
	d.lastAccessed = time.Now()
	d.value = value
}

//NewData initialises a (pointer to) an instance of a Data struct, initialising the number of writes to 1 and reads to 0
//it also initialises the timestamp to now
func NewData(owner string, value string) *Data {
	d := Data{
		owner:        owner,
		value:        value,
		lastAccessed: time.Now(),
		writes:       1,
		reads:        0,
	}
	return &d
}

//Key is a struct used to return information about a key, excluding its value. Used by the list functions
type Key struct {
	Key    string `json:"key"`
	Owner  string `json:"owner"`
	Writes int    `json:"writes"`
	Reads  int    `json:"reads"`
	Age    int64  `json:"age"`
}

func directRemoveOldKeys() {
	for len(kvStore) > MaxDepth { //Should only run once, but no harm in being certain
		var oldestKey string
		oldestTime := time.Now()
		for key, data := range kvStore {
			if data.lastAccessed.Before(oldestTime) {
				oldestTime = data.lastAccessed
				oldestKey = key
			}
		}
		delete(kvStore, oldestKey)
	}
}

func directLookupValue(key, user string) (string, error) {
	value, present := kvStore[key]
	if !present {
		return "", ErrKeyNotPresent
	}
	if !value.isAuthorised(user) {
		return "", ErrUnauthorized
	}
	return value.getValue(), nil
}

func directPutValue(key, user, value string) error {
	data, present := kvStore[key]
	if present {
		if data.isAuthorised(user) {
			data.setValue(value)
			return nil
		} else {
			return ErrUnauthorized
		}
	}
	data = NewData(user, value)
	kvStore[key] = data
	directRemoveOldKeys()
	return nil
}

func directDelete(key, user string) error {
	value, present := kvStore[key]
	if !present {
		return ErrKeyNotPresent
	}
	if !value.isAuthorised(user) {
		return ErrUnauthorized
	}
	delete(kvStore, key)
	return nil
}

func directGetKeyInfo(key string) (*Key, error) {
	data, err := kvStore[key]
	if !err {
		return nil, ErrKeyNotPresent
	}
	output := Key{
		Key:    key,
		Owner:  data.owner,
		Writes: data.writes,
		Reads:  data.reads,
		Age:    time.Since(data.lastAccessed).Milliseconds(),
	}
	return &output, nil
}

func directListStore() ([]byte, error) {
	output := []*Key{}
	for key := range kvStore {
		data, err := directGetKeyInfo(key)
		if err != nil {
			fmt.Println("something has gone terribly wrong")
			fmt.Println("list store cannot find a key")
			return nil, ErrKeyNotPresent //this should never fire
		}
		output = append(output, data)
	}
	jsonOut, errJSON := json.Marshal(output)
	if errJSON != nil {
		return nil, errJSON
	}
	return jsonOut, nil
}

func directListKey(key string) ([]byte, error) {
	data, err := directGetKeyInfo(key)
	if err != nil {
		return nil, err
	}
	jsonOut, errJSON := json.Marshal(data)
	if errJSON != nil {
		return nil, errJSON
	}
	return jsonOut, nil
}
