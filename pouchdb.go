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
//     "honnef.co/go/js/console"
)

type PouchDB struct {
    o   *js.Object
}

type pouchResult struct {
    result  *js.Object
    err     *js.Object
}

func (pr *pouchResult) Result() (*js.Object,error) {
    if pr.err == nil {
        return pr.result,nil
    }
    return pr.result,&js.Error{pr.err}
}

// New creates a database or opens an existing one.
// See: http://pouchdb.com/api.html#create_database
func New(db_name string) *PouchDB {
    return &PouchDB{ js.Global.Get("PouchDB").New(db_name) }
}

// NewFromOpts creates a database or opens an existing one.
// See: http://pouchdb.com/api.html#create_database
func NewFromOpts(opts *js.Object) *PouchDB {
    return &PouchDB{ js.Global.Get("PouchDB").New(opts) }
}

// Info fetches information about a database.
// See: http://pouchdb.com/api.html#database_information
func (db *PouchDB) Info() (*js.Object,error) {
    infoChan := make(chan *pouchResult)
    db.o.Call("info",func(err *js.Object, info *js.Object) {
        infoChan <- &pouchResult{info, err}
    })
    result := <-infoChan
    return result.Result()
}

// Deestroy will delete the database.
// See: http://pouchdb.com/api.html#delete_database
func (db *PouchDB) Destroy(args ...interface{}) {
    db.o.Call("destroy", args...)
}

// Put will create a new document or update an existing document.
// See: http://pouchdb.com/api.html#create_document
func (db *PouchDB) Put(args ...interface{}) {
    db.o.Call("put", args...)
}

// Get retrieves a document, specified by docId.
// Seehttp://pouchdb.com/api.html#fetch_document
func (db *PouchDB) Get(args ...interface{}) {
    db.o.Call("get", args...)
}

// Delete will delete the document.
// See: http://pouchdb.com/api.html#delete_document
func (db *PouchDB) Delete(args ...interface{}) {
    db.o.Call("delete", args...)
}

// BulkDocs will create, update or delete multiple documents.
// See: http://pouchdb.com/api.html#batch_create
func (db *PouchDB) BulkDocs(args ...interface{}) {
    db.o.Call("bulkDocs", args...)
}

// AllDocs will fetch multiple documents.
// See http://pouchdb.com/api.html#batch_fetch
func (db *PouchDB) AllDocs(args ...interface{}) {
    db.o.Call("allDocs", args...)
}

// Replicate will replicate data from source to target
// See: http://pouchdb.com/api.html#replication
// func (db *PouchDB) Replicate(source, target string, args ...*js.Object) {
//     db.o.Call("replicate", source, target, args...)
// }

// ViewCleanup cleans up any stale map/reduce indexes.
// See: http://pouchdb.com/api.html#view_cleanup
func (db *PouchDB) ViewCleanup(fn interface{}) {
    db.o.Call("viewCleanup", fn)
}

// Compact triggers a compaction operation in the local or remote database.
// See: http://pouchdb.com/api.html#compaction
func (db *PouchDB) Compact(args ...interface{}) {
    db.o.Call("compact", args...)
}

// RevsDiff will, given a set of document/revision IDs return the subset of
// those that do not correspond to revisions stored in the database.
// See: http://pouchdb.com/api.html#revisions_diff
func (db *PouchDB) RevsDiff(diff *js.Object, fn interface{}) {
    db.o.Call("revsDiff", diff, fn)
}

// Defaults sets default options.
// See: http://pouchdb.com/api.html#defaults
func (db *PouchDB) Defaults(opts *js.Object) {
    db.o.Call("defaults", opts)
}

// DebugEnable enables debugging for the specified module.
// See: http://pouchdb.com/api.html#debug_mode
func (db *PouchDB) DebugEnable(module string) {
    db.o.Get("debug").Call("enable", module)
}

// DebugDisable disables debugging.
func (db *PouchDB) DebugDisable() {
    db.o.Get("debug").Call("disable")
}
