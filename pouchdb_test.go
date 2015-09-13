// +build js

package pouchdb

import (
	"testing"
)

type TestDoc struct {
	DocId      string `json:"_id"`
	DocRev     string `json:"_rev,omitempty"`
	DocDeleted bool   `json:"_deleted"`
	Value      string `json:"foo"`
}

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
	err := db.Destroy(Options{})
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
	rev, err := db.Put(doc)
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
	if got["_rev"] != rev {
		t.Fatalf("Retrieved unexpected rev: %s instead of %s", got["_rev"], rev)
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
	db.Destroy(Options{})
}

type TestRow struct {
	Id  string  `json:"id"`
	Key string  `json:"key"`
	Doc TestDoc `json:"doc"`
}

type TestDocCollection struct {
	TotalRows int       `json:"total_rows"`
	Offset    int       `json:"offset"`
	Rows      []TestRow `json:"rows"`
}

func TestBulkDocs(t *testing.T) {
	db := New("testdb")
	docs := []TestDoc{
		TestDoc{
			DocId: "foo",
			Value: "foo",
		},
		TestDoc{
			DocId: "bar",
			Value: "bar",
		},
	}
	results, err := db.BulkDocs(docs, Options{})
	if err != nil {
		t.Fatalf("Received error from BulkDocs: %s", err)
	}
	for i, doc := range docs {
		if !results[i]["ok"].(bool) {
			t.Fatalf("BulkDocs() failed")
		}
		if doc.DocId != results[i]["id"] {
			t.Fatalf("BulkDocs() returned _id %s, expected %s", results[i]["id"], doc.DocId)
		}
	}
	// test AllDocs()
	allDocs := TestDocCollection{}
	db.AllDocs(&allDocs, Options{
		"include_docs": true,
	})
	if allDocs.TotalRows != 2 {
		t.Fatalf("Got an unexpected number of results: %d", allDocs.TotalRows)
	}
	if allDocs.Offset != 0 {
		t.Fatalf("Got an unexpected offset: %d", allDocs.Offset)
	}
	for _, row := range allDocs.Rows {
		doc := row.Doc
		if doc.DocId != "foo" && doc.DocId != "bar" {
			t.Fatalf("Unexpected _id in result set: %s", doc.DocId)
		}
	}
	db.Destroy(Options{})
}

func TestRemove(t *testing.T) {
	db := New("testdb")
	doc := TestDoc{
		DocId: "foo",
	}
	rev, err := db.Put(doc)
	if err != nil {
		t.Fatalf("Failed to create document: %s", err)
	}
	doc.DocRev = rev
	delRev, err := db.Remove(doc, Options{})
	if err != nil {
		t.Fatalf("Received error from Delete: %s", err)
	}
	var deletedDoc TestDoc
	err = db.Get("foo", &deletedDoc, Options{"rev": delRev})
	if err != nil {
		t.Fatalf("Error fetching deleted doc: %s", err)
	}
	if !deletedDoc.DocDeleted {
		t.Fatalf("Remove() did not properly delete the document")
	}
	db.Destroy(Options{})
}
