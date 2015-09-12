// +build js

// Package pouchdb provides GopherJS bindings for PouchDB.
// Whenever possible, the PouchDB function calls have been made more
// Go idiomatic. This means:
//  - They don't take optional arguments. Where appropriate, multiple
//    versions of a function exist with different argument lists.
//  - They don't use call backs or return Promises
//  - They have been made synchronous. If you need asynchronous operation,
//    wrap your calls in goroutines
//  - They return errors as the last return value (Go style) rather than the
//    first (JS style)
package pouchdb

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
	"honnef.co/go/js/console"
)

type PouchDB struct {
	o *js.Object
}

type Options map[string]interface{}

type Result map[string]interface{}

var GlobalPouch *js.Object

func globalPouch() *js.Object {
console.Log("foo")
	if GlobalPouch != nil && jsbuiltin.Typeof(GlobalPouch) != "undefined" {
		return GlobalPouch
	}
	GlobalPouch := js.Global.Get("PouchDB")
	if jsbuiltin.Typeof(GlobalPouch) == "undefined" {
		// This is necessary because gopherjs runs the test from /tmp
		// rather than from the current directory, which confuses nodejs
		// as to where to search for modules
		cwd := js.Global.Get("process").Call("cwd").String()
		GlobalPouch = js.Global.Call("require",  cwd + "/node_modules/pouchdb")
	}
	return GlobalPouch
}

// New creates a database or opens an existing one.
// See: http://pouchdb.com/api.html#create_database
func New(db_name string) *PouchDB {
	return &PouchDB{globalPouch().New(db_name)}
}

// NewFromOpts creates a database or opens an existing one.
// See: http://pouchdb.com/api.html#create_database
func NewFromOpts(opts Options) *PouchDB {
	return &PouchDB{globalPouch().New(opts)}
}

// Info fetches information about a database.
// See: http://pouchdb.com/api.html#database_information
func (db *PouchDB) Info() (Result, error) {
	result := newResult()
	db.o.Call("info", result.Done)
	return result.ReadResult()
}

// Deestroy will delete the database.
// See: http://pouchdb.com/api.html#delete_database
func (db *PouchDB) Destroy() error {
	result := newResult()
	db.o.Call("destroy", result.Done)
	_,err := result.Read()
	return err
}

// Put will create a new document or update an existing document.
// See: http://pouchdb.com/api.html#create_document
// func (db *PouchDB) Put(args ...interface{}) {
// 	db.o.Call("put", args...)
// }

// Get retrieves a document, specified by docId.
// Seehttp://pouchdb.com/api.html#fetch_document
// func (db *PouchDB) Get(args ...interface{}) {
// 	db.o.Call("get", args...)
// }

// Delete will delete the document.
// See: http://pouchdb.com/api.html#delete_document
// func (db *PouchDB) Delete(args ...interface{}) {
// 	db.o.Call("delete", args...)
// }

// BulkDocs will create, update or delete multiple documents.
// See: http://pouchdb.com/api.html#batch_create
// func (db *PouchDB) BulkDocs(args ...interface{}) {
// 	db.o.Call("bulkDocs", args...)
// }

// AllDocs will fetch multiple documents.
// See http://pouchdb.com/api.html#batch_fetch
// func (db *PouchDB) AllDocs(args ...interface{}) {
// 	db.o.Call("allDocs", args...)
// }

// Replicate will replicate data from source to target in the foreground.
// For "live" replication use ReplicateLive()
// See: http://pouchdb.com/api.html#replication
// func Replicate(source, target string, options Options) {
// 	options["live"] = false
// 	js.Global.Get("PouchDB").Call("replicate", options)
// }

// Replicate will replicate data from source to target in the background.
// This method returns a *ChangeFeed which can be used to monitor progress
// in a Go routine. For foreground sync, use Replicate().
// See: http://pouchdb.com/api.html#replication

// ViewCleanup cleans up any stale map/reduce indexes.
// See: http://pouchdb.com/api.html#view_cleanup
// func (db *PouchDB) ViewCleanup(fn interface{}) {
// 	db.o.Call("viewCleanup", fn)
// }

// Compact triggers a compaction operation in the local or remote database.
// See: http://pouchdb.com/api.html#compaction
// func (db *PouchDB) Compact(args ...interface{}) {
// 	db.o.Call("compact", args...)
// }

// RevsDiff will, given a set of document/revision IDs return the subset of
// those that do not correspond to revisions stored in the database.
// See: http://pouchdb.com/api.html#revisions_diff
// func (db *PouchDB) RevsDiff(diff *js.Object, fn interface{}) {
// 	db.o.Call("revsDiff", diff, fn)
// }

// Defaults sets default options.
// See: http://pouchdb.com/api.html#defaults
// func (db *PouchDB) Defaults(opts *js.Object) {
// 	db.o.Call("defaults", opts)
// }

// Debug enables debugging for the specified module.
// See: http://pouchdb.com/api.html#debug_mode
func Debug(module string) {
	globalPouch().Get("debug").Call("enable", module)
}

// DebugDisable disables debugging.
func DebugDisable() {
	globalPouch().Get("debug").Call("disable")
}
