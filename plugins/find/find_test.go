// +build js

package pouchdb_find_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/kr/pretty"
	"github.com/pmezard/go-difflib/difflib"
)

type myDB struct {
	*pouchdb.PouchDB
	*pouchdb_find.PouchPluginFind
}

func init() {
	// This is necessary because gopherjs runs the test from /tmp
	// rather than from the current directory, which confuses nodejs
	// as to where to search for modules
	cwd := strings.TrimSuffix(js.Global.Get("process").Call("cwd").String(), "plugins/find")
	pouchdb.GlobalPouch = js.Global.Call("require", cwd+"/node_modules/pouchdb")
	find := js.Global.Call("require", cwd+"/node_modules/pouchdb-find")
	pouchdb.Plugin(find)
}

func TestFind(t *testing.T) {
	mainDB := pouchdb.New("finddb")
	mainDB.Destroy(pouchdb.Options{}) // to ensure a clean slate
	mainDB = pouchdb.New("finddb")

	db := myDB{
		mainDB,
		pouchdb_find.New(mainDB),
	}

	err := db.CreateIndex(pouchdb_find.Index{
		Fields: []string{"name", "size"},
	})
	if err != nil {
		t.Fatalf("Error from CreateIndex: %s\n", err)
	}

	// Create the same index again; we should be notified it exists
	err = db.CreateIndex(pouchdb_find.Index{
		Fields: []string{"name", "size"},
	})
	if err != nil && !err.IndexExists() {
		t.Fatalf("Error re-creating index: %s\n", err)
	}
	if !err.IndexExists() {
		t.Fatalf("We were not notified that the index already existed\n")
	}

	expected := []*pouchdb_find.IndexDef{
		&pouchdb_find.IndexDef{
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
		&pouchdb_find.IndexDef{
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

	idxs, err2 := db.GetIndexes()
	if err2 != nil {
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
