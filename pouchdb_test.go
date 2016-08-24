// +build js

package pouchdb

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/gopherjs/gopherjs/js"
)

type TestDoc struct {
	DocId      string `json:"_id"`
	DocRev     string `json:"_rev,omitempty"`
	DocDeleted bool   `json:"_deleted"`
	Value      string `json:"foo"`
}

var memdown *js.Object

func init() {
	GlobalPouch = js.Global.Call("require", "pouchdb")
	memdown = js.Global.Call("require", "memdown")
}

func newPouch(dbname string) *PouchDB {
	return NewWithOpts(dbname, Options{
		DB: memdown,
	})
}

func TestNew(t *testing.T) {
	db := newPouch("testdb")
	info, err := db.Info()
	if err != nil {
		t.Fatalf("Info() returned error: %s", err)
	}
	if info.DBName != "testdb" {
		t.Fatalf("Info() returned unexpected db_name '%s'", info.DBName)
	}
	db.Destroy(Options{})
}

func TestNewFromOpts(t *testing.T) {
	db := NewWithOpts("testdb", Options{})
	info, err := db.Info()
	if err != nil {
		t.Fatalf("Info() returned error: %s", err)
	}
	if info.DBName != "testdb" {
		t.Fatalf("Info() returned unexpected db_name '%s'", info.DBName)
	}
	db.Destroy(Options{})
}

func TestDestory(t *testing.T) {
	db := newPouch("testdb")
	err := db.Destroy(Options{})
	if err != nil {
		t.Fatalf("Destroy() resulted in an error: %s", err)
	}
}

func TestPutGet(t *testing.T) {
	db := newPouch("testdb")
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
	db := newPouch("testdb")
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
		IncludeDocs: true,
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
	db := newPouch("testdb")
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
	err = db.Get("foo", &deletedDoc, Options{Rev: delRev})
	if err != nil {
		t.Fatalf("Error fetching deleted doc: %s", err)
	}
	if !deletedDoc.DocDeleted {
		t.Fatalf("Remove() did not properly delete the document")
	}
	db.Destroy(Options{})
}

func TestViewCleanup(t *testing.T) {
	db := newPouch("testdb")
	err := db.ViewCleanup()
	if err != nil {
		t.Fatalf("Error cleaning up views: %s", err)
	}
	db.Destroy(Options{})
}

func TestCompact(t *testing.T) {
	db := newPouch("testdb")
	err := db.Compact(Options{})
	if err != nil {
		t.Fatalf("Error compacting database: %s", err)
	}
	db.Destroy(Options{})
}

func TestAttachments(t *testing.T) {
	db := newPouch("testdb")
	body1 := "A légpárnás hajóm tele van angolnákkal"
	att1 := &Attachment{
		Name: "foo.txt",
		Type: "text/plain",
		Body: strings.NewReader(body1),
	}
	rev, err := db.PutAttachment("foo", att1, "")
	if err != nil {
		t.Fatalf("Error putting attachment: %s", err)
	}
	if len(rev) == 0 {
		t.Fatal("PutAttachment() returned a 0-byte rev")
	}
	att2, err := db.Attachment("foo", "foo.txt", "")
	buf := new(bytes.Buffer)
	buf.ReadFrom(att2.Body)
	body2 := buf.String()
	if body1 != body2 {
		t.Fatalf("The fetched body doesn't match. Got '%s' instead of '%s'", body2, body1)
	}
	rev, err = db.DeleteAttachment("foo", "foo.txt", rev)
	if err != nil {
		t.Fatalf("Error deleting attachment: %s", err)
	}
	if len(rev) == 0 {
		t.Fatal("DeleteAttachment() returned a 0-byte rev")
	}
	db.Destroy(Options{})
}

func TestReplicate(t *testing.T) {
	newPouch("db1").Destroy(Options{})
	newPouch("db2").Destroy(Options{})
	db1 := newPouch("db1")
	doc1 := TestDoc{
		DocId: "oink",
		Value: "foo",
	}
	_, err := db1.Put(doc1)
	if err != nil {
		t.Fatalf("Error putting document: %s", err)
	}
	err = db1.Get(doc1.DocId, &doc1, Options{})
	if err != nil {
		t.Fatalf("Error re-reading doc1: %s", err)
	}
	db2 := newPouch("db2")
	results, err := Replicate(db1, db2, Options{})
	if err != nil {
		t.Fatalf("Error replicating: %s", err)
	}
	if x := int(results["docs_read"].(float64)); x != 1 {
		t.Fatalf("Unexpected number of docs read: %d", x)
	}
	if x := int(results["docs_written"].(float64)); x != 1 {
		t.Fatalf("Unexpected number of docs written: %d", x)
	}
	if x := int(results["doc_write_failures"].(float64)); x != 0 {
		t.Fatalf("Unexpected number of failures: %d", x)
	}
	doc2 := TestDoc{}
	db2.Get("oink", &doc2, Options{})
	if !reflect.DeepEqual(doc1, doc2) {
		t.Fatalf("Document is different after replication")
	}
	db1.Destroy(Options{})
	db2.Destroy(Options{})
}
