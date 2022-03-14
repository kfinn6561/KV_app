package KVStore_test

import (
	"encoding/json"
	"store/KVStore"
	"testing"
)

func FindIndex(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

func handleShutdown(t *testing.T) {
	err := KVStore.Shutdown()
	if err != nil {
		t.Error("Unable to shutdown properly", err)
	}
}

func TestStartupShutdown(t *testing.T) {
	KVStore.Startup(100, 100)
	err := KVStore.Shutdown()
	if err != nil {
		t.Error("Unable to shutdown properly", err)
	}
}

func TestPutGet(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)
	key := "key"
	user := "test"
	data := "value"

	if err := KVStore.PutValue(key, user, data); err != nil {
		t.Error("unable to put a value in the kv store", err)
	}

	testData, err := KVStore.LookupValue(key, user)

	if err != nil || testData != data {
		t.Errorf("unable to retrieve value from KV store. Wanted %s, got %s. Error is %v\n", data, testData, err)
	}
}

func TestPutChange(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)
	key := "key"
	user := "test"
	data := "value"
	newData := "new data"

	if err := KVStore.PutValue(key, user, data); err != nil {
		t.Error("unable to put a value in the kv store", err)
	}
	testData, err := KVStore.LookupValue(key, user)
	if err != nil || testData != data {
		t.Errorf("unable to retrieve value from KV store. Wanted %s, got %s. Error is %v\n", data, testData, err)
	}

	errPut := KVStore.PutValue(key, user, newData)
	if errPut != nil {
		t.Error("unable to put a second value in the kv store under the same key", errPut)
	}

	testData2, err2 := KVStore.LookupValue(key, user)
	if err2 != nil || testData2 != newData {
		t.Errorf("unable to retrieve changed data from KV store. Wanted %s, got %s. Error is %v\n", newData, testData2, err)
	}

}

func TestPutChangeUnauthorised(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)
	key := "key"
	user := "test"
	wrongUser := "wrong"
	data := "value"
	newData := "new data"

	if err := KVStore.PutValue(key, user, data); err != nil {
		t.Error("unable to put a value in the kv store", err)
	}
	testData, err := KVStore.LookupValue(key, user)
	if err != nil || testData != data {
		t.Errorf("unable to retrieve value from KV store. Wanted %s, got %s. Error is %v\n", data, testData, err)
	}

	errPut := KVStore.PutValue(key, wrongUser, newData)
	if errPut != KVStore.ErrUnauthorized {
		t.Error("Able to change a value for a different user")
	}

}

func TestGetNotThere(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)
	key := "key"
	user := "test"

	testData, err := KVStore.LookupValue(key, user)
	if err != KVStore.ErrKeyNotPresent || testData != "" {
		t.Errorf("able to retrieve value from KV store when key is not present. Got %s. Error is %v\n", testData, err)
	}
}

func TestGetNotAuthorised(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)
	key := "key"
	user := "test"
	wrongUser := "wrong"
	data := "value"

	if err := KVStore.PutValue(key, user, data); err != nil {
		t.Error("unable to put a value in the kv store", err)
	}

	testData, err := KVStore.LookupValue(key, wrongUser)
	if err != KVStore.ErrUnauthorized || testData != "" {
		t.Errorf("able to retrieve value from KV store with wrong user. Got %s. Error is %v\n", testData, err)
	}

}

func TestDelete(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)
	key := "key"
	user := "test"
	data := "value"

	if err := KVStore.PutValue(key, user, data); err != nil {
		t.Error("unable to put a value in the kv store", err)
	}

	testData, err1 := KVStore.LookupValue(key, user)
	if err1 != nil || testData != data {
		t.Errorf("unable to retrieve value from KV store. Wanted %s, got %s. Error is %v\n", data, testData, err1)
	}

	err := KVStore.Delete(key, user)
	if err != nil {
		t.Error("unable to delete a key from the kv store", err)
	}
	testData, errLookup := KVStore.LookupValue(key, user)
	if errLookup != KVStore.ErrKeyNotPresent || testData != "" {
		t.Errorf("able to retrieve value from KV store after deleting. Got %s. Error is %v\n", testData, err)
	}
}

func TestDeleteUnauthorised(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)

	key := "key"
	user := "test"
	wrongUser := "wrong"
	data := "value"

	if err := KVStore.PutValue(key, user, data); err != nil {
		t.Error("unable to put a value in the kv store", err)
	}

	testData, err1 := KVStore.LookupValue(key, user)
	if err1 != nil || testData != data {
		t.Errorf("unable to retrieve value from KV store. Wanted %s, got %s. Error is %v\n", data, testData, err1)
	}

	err := KVStore.Delete(key, wrongUser)
	if err != KVStore.ErrUnauthorized {
		t.Error("able to delete a key from another user", err)
	}
	testData, errLookup := KVStore.LookupValue(key, user)
	if errLookup != nil || testData != data {
		t.Errorf("attempt to delete by unauthorised user corrupted the data. Wanted %s, got %s. Error is %v\n", data, testData, errLookup)
	}
}

func TestDeleteNotThere(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)
	key := "key"
	user := "test"

	err := KVStore.Delete(key, user)
	if err != KVStore.ErrKeyNotPresent {
		t.Error("able to delete a non-existent key", err)
	}
}

func TestListing(t *testing.T) {
	KVStore.Startup(100, 100)
	defer handleShutdown(t)
	keys := []string{"key1", "key2", "key3"}
	users := []string{"user1", "user2", "user3"}
	values := []string{"value1", "value2", "value3"}
	wrongKey := "wrong"

	for i := 0; i < len(keys); i++ {
		if err := KVStore.PutValue(keys[i], users[i], values[i]); err != nil {
			t.Error("unable to put a value in the kv store", err)
		}
	}
	t.Run("TestListStore", func(t *testing.T) {
		testJSON, err := KVStore.ListStore()
		if err != nil {
			t.Error("unable to list store contents", err)
		}
		var testData []KVStore.Key
		errJSON := json.Unmarshal(testJSON, &testData)
		if errJSON != nil {
			t.Error("error unmarshaling store list json", errJSON)
		}
		if len(testData) != len(keys) {
			t.Errorf("did not list enough objects. Expected %d, got %d\n", len(keys), len(testData))
		} else {
			for j := 0; j < len(testData); j++ {
				data := testData[j]
				k := FindIndex(keys, data.Key) //list store may not return keys in the same order
				if data.Key != keys[k] || data.Owner != users[k] {
					t.Errorf("returned key had incorrect values. expected key %s and owner %s, got %v\n", keys[k], users[k], data)
				}
			}
		}
	})

	t.Run("TestListKey", func(t *testing.T) {
		testJSON, err := KVStore.ListKey(keys[0])
		if err != nil {
			t.Error("unable to list store key", err)
		}
		var testData KVStore.Key
		errJSON := json.Unmarshal(testJSON, &testData)
		if errJSON != nil {
			t.Error("error unmarshaling json", errJSON)
		}
		if testData.Key != keys[0] || testData.Owner != users[0] {
			t.Errorf("returned key had incorrect values. expected key %s and owner %s, got %v\n", keys[0], users[0], testData)
		}
	})

	t.Run("TestListKeyNotThere", func(t *testing.T) {
		testData, err := KVStore.ListKey(wrongKey)
		if err != KVStore.ErrKeyNotPresent || testData != nil {
			t.Errorf("able to list non-existent key. Got %s. Error is %v\n", testData, err)
		}
	})

}

func TestDepth(t *testing.T) {
	keys := []string{"key1", "key2", "key3", "key4"}
	users := []string{"user1", "user2", "user3", "user4"}
	values := []string{"value1", "value2", "value3", "value4"}

	KVStore.Startup(100, len(keys)-1)
	defer handleShutdown(t)

	for i := 0; i < len(keys); i++ {
		if err := KVStore.PutValue(keys[i], users[i], values[i]); err != nil {
			t.Error("unable to put a value in the kv store", err)
		}
	} //key1 should be ejected from the store

	testData, errLookup := KVStore.LookupValue(keys[0], users[0])
	if errLookup != KVStore.ErrKeyNotPresent || testData != "" {
		t.Errorf("able to retrieve value from KV store that should have been rejected due to depth. Got %s. Error is %v\n", testData, errLookup)
	}

}
