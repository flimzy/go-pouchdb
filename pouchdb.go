// +build js

package pouchdb

import (
    "github.com/gopherjs/gopherjs/js"
    "honnef.co/go/js/console"
)

type PouchDB struct {
    o   *js.Object
}

// New creates a database or opens an existing one.
// See also: http://pouchdb.com/api.html#create_database
func New(args ...interface{}) *PouchDB {
    return &PouchDB{ js.Global.Get("PouchDB").New(args...) }
}

func (db *PouchDB) Info(fn interface{}) {
    console.Log(db)
    db.o.Call("info",fn)
}
