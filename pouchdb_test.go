// +build js

package pouchdb

import (
	"testing"
)

func TestNew(t *testing.T) {
	// console.Log("foo")
	db := New("testdb")
	info, err := db.Info()
	if err != nil {
		t.Fatalf("Info() returned error: %s", err)
	}
	if info["db_name"] != "testdb" {
		t.Fatalf("Info() returned unexpected db_name '%s'", info["db_name"])
	}
}

func TestNewFromOpts(t *testing.T) {
	db := NewFromOpts(Options{
		"name": "testdb",
	})
	info, err := db.Info()
	if err != nil {
		t.Fatalf("Info() returned error: %s", err)
	}
	if info["db_name"] != "testdb" {
		t.Fatalf("Info() returned unexpected db_name '%s'", info["db_name"])
	}
}

func TestDestory(t *testing.T) {
	db := New("testdb")
	err := db.Destroy()
	if err != nil {
		t.Fatalf("Destroy() resulted in an error: %s", err)
	}
}

func TestPutGet(t *testing.T) {
	db := New("testdb")
	doc := map[string]interface{}{
		"_id": "foobar",
		"foo": "bar",
	}
	err := db.Put(doc)
	if err != nil {
		t.Fatalf("Error calling Put(): %s", err)
	}
	var got map[string]interface{}
	err = db.Get("foobar", &got, Options{})
	if err != nil {
		t.Fatalf("Error calling Get(): %s", err)
	}
	if got["_id"] != doc["_id"] {
		t.Fatalf("Retrieved unexpected _id: %s instead of %s", got["_id"], doc["_id"])
	}
	if doc["foo"] != doc["foo"] {
		t.Fatalf("Retrieved unexpected payload 'foo': %s instead of %s", got["foo"], doc["foo"])
	}
	rev, ok := got["_rev"].(string)
	if !ok {
		t.Fatal("_rev is not a string")
	}
	if len(rev) == 0 {
		t.Fatal("_rev is empty")
	}
}
