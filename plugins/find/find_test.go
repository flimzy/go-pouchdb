// +build js

package pouchdb_find_test

import (
	"testing"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
)

type myDB struct {
	*pouchdb.PouchDB
	*pouchdb_find.PouchPluginFind
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
	if err != nil && ! err.IndexExists() {
		t.Fatalf("Error re-creating index: %s\n", err)
	}
	if ! err.IndexExists() {
		t.Fatalf("We were not notified that the index already existed\n")
	}

// 	idxs, err := db.GetIndexes()
// 	fmt.Printf("X:%v", idxs)
// 	t.Fatalf("ouch")
// 	fmt.Printf("Yuppers\n")
}
