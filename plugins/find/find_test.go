// +build js

package find_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/kr/pretty"
	"github.com/pmezard/go-difflib/difflib"
)

type myDB struct {
	*pouchdb.PouchDB
	*find.PouchPluginFind
}

var memdown *js.Object

func init() {
	pouchdb.GlobalPouch = js.Global.Call("require", "pouchdb")
	find := js.Global.Call("require", "pouchdb-find")

	pouchdb.Plugin(find)
	memdown = js.Global.Call("require", "memdown")
}

func TestFind(t *testing.T) {
	mainDB := pouchdb.NewWithOpts("finddb", pouchdb.Options{
		DB: memdown,
	})
	mainDB.Destroy(pouchdb.Options{}) // to ensure a clean slate
	mainDB = pouchdb.NewWithOpts("finddb", pouchdb.Options{
		DB: memdown,
	})

	db := myDB{
		mainDB,
		find.New(mainDB),
	}

	ferr := db.CreateIndex(find.Index{
		Fields: []string{"name", "size"},
	})
	if ferr != nil {
		t.Fatalf("Error from CreateIndex: %s\n", ferr)
	}

	// Create the same index again; we should be notified it exists
	ferr = db.CreateIndex(find.Index{
		Fields: []string{"name", "size"},
	})
	if ferr != nil && !find.IsIndexExists(ferr) {
		t.Fatalf("Error re-creating index: %s\n", ferr)
	}
	if !find.IsIndexExists(ferr) {
		t.Fatalf("We were not notified that the index already existed\n")
	}

	expected := []*find.IndexDef{
		&find.IndexDef{
			Ddoc: "",
			Name: "_all_docs",
			Type: "special",
			Def: struct {
				Fields []map[string]string "json:\"fields\""
			}{
				Fields: []map[string]string{
					map[string]string{"_id": "asc"},
				},
			},
		},
		&find.IndexDef{
			Ddoc: "_design/idx-1c1850c82e1b5105c94a267ec61322ce",
			Name: "idx-1c1850c82e1b5105c94a267ec61322ce",
			Type: "json",
			Def: struct {
				Fields []map[string]string "json:\"fields\""
			}{
				Fields: []map[string]string{
					map[string]string{"name": "asc"},
					map[string]string{"size": "asc"},
				},
			},
		},
	}

	idxs, err := db.GetIndexes()
	if err != nil {
		t.Fatalf("Error running GetIndexes: %s", err)
	}
	if !reflect.DeepEqual(idxs, expected) {
		DumpDiff(expected, idxs)
		t.Fatal()
	}

	// Retrieval
	doc := map[string]interface{}{
		"_id":  "12345",
		"name": "Bob",
		"size": 48,
	}
	_, err = db.Put(doc)
	if err != nil {
		t.Fatalf("Error calling Put(): %s", err)
	}
	doc = map[string]interface{}{
		"_id":  "23456",
		"name": "Alice",
	}
	_, err = db.Put(doc)
	if err != nil {
		t.Fatalf("Error calling Put(): %s", err)
	}

	var resultDoc map[string]interface{}
	req := find.Request{
		Selector: map[string]interface{}{
			"name": "Bob",
		},
		Fields: []string{
			"_id",
			"name",
			"size",
		},
	}
	expectedResult := map[string]interface{}{
		"docs": []interface{}{
			map[string]interface{}{
				"_id":  "12345",
				"name": "Bob",
				"size": float64(48),
			},
		},
	}
	err = db.Find(req, &resultDoc)
	if err != nil {
		t.Fatalf("Error executing Find(): %s", err)
	}
	if !reflect.DeepEqual(expectedResult, resultDoc) {
		DumpDiff(expectedResult, resultDoc)
		t.Fatal()
	}

	err = db.DeleteIndex(idxs[1])
	if err != nil {
		t.Fatalf("Error running DeleteIndex: %s", err)
	}

	expected = []*find.IndexDef{
		&find.IndexDef{
			Ddoc: "",
			Name: "_all_docs",
			Type: "special",
			Def: struct {
				Fields []map[string]string "json:\"fields\""
			}{
				Fields: []map[string]string{
					map[string]string{"_id": "asc"},
				},
			},
		},
	}
	idxs, err = db.GetIndexes()
	if err != nil {
		t.Fatalf("Error running GetIndexes: %s", err)
	}
	if !reflect.DeepEqual(idxs, expected) {
		DumpDiff(expected, idxs)
		t.Fatal()
	}
}

func DumpDiff(expectedObj, actualObj interface{}) {
	expected := pretty.Sprintf("%# v\n", expectedObj)
	actual := pretty.Sprintf("%# v\n", actualObj)
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		B:        difflib.SplitLines(actual),
		FromFile: "expected",
		ToFile:   "actual",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	fmt.Print(text)
}
