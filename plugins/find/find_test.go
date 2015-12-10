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
	db := myDB{
		mainDB,
		pouchdb_find.New(mainDB),
	}
	
	db.Compact(pouchdb.Options{})
}
