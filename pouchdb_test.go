// +build js

package pouchdb

import (
	"testing"
	"honnef.co/go/js/console"
)

func TestNew(t *testing.T) {
console.Log("foo")
	db := New("testdb")
	info,err := db.Info()
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
	info,err := db.Info()
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
