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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
	// 	"honnef.co/go/js/console"
)

type PouchDB struct {
	o *js.Object
}

type Options map[string]interface{}

type Result map[string]interface{}

var GlobalPouch *js.Object

func globalPouch() *js.Object {
	if GlobalPouch != nil && jsbuiltin.TypeOf(GlobalPouch) != "undefined" {
		return GlobalPouch
	}
	GlobalPouch := js.Global.Get("PouchDB")
	if jsbuiltin.TypeOf(GlobalPouch) == "undefined" {
		// This is necessary because gopherjs runs the test from /tmp
		// rather than from the current directory, which confuses nodejs
		// as to where to search for modules
		cwd := js.Global.Get("process").Call("cwd").String()
		GlobalPouch = js.Global.Call("require", cwd+"/node_modules/pouchdb")
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
	rw := newResultWaiter()
	db.o.Call("info", rw.Done)
	return rw.ReadResult()
}

// Deestroy will delete the database.
// See: http://pouchdb.com/api.html#delete_database
func (db *PouchDB) Destroy(opts Options) error {
	rw := newResultWaiter()
	db.o.Call("destroy", opts, rw.Done)
	return rw.Error()
}

// convertJSONObject takes an intterface{} and runs it through json.Marshal()
// and json.Unmarshal() so that any struct tags will be applied.
func convertJSONObject(input, output interface{}) error {
	encoded, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, output)
}

// convertJSObject converts the provided *js.Object to an interface{} then
// calls convertJSONObject. This is necessary for objects, because json.Marshal
// ignores any unexported fields in objects, and this includes practically
// everything inside a js.Object.
func convertJSObject(jsObj *js.Object, output interface{}) error {
	return convertJSONObject(jsObj.Interface(), output)
}

// Put will create a new document or update an existing document.
// See: http://pouchdb.com/api.html#create_document
func (db *PouchDB) Put(doc interface{}) (newrev string, err error) {
	var convertedDoc interface{}
	convertJSONObject(doc, &convertedDoc)
	rw := newResultWaiter()
	db.o.Call("put", convertedDoc, rw.Done)
	return rw.ReadRev()
}

// Get retrieves a document, specified by docId.
// The document is unmarshalled into the given object.
// Some fields (like _conflicts) will only be returned if the
// options require it. Please refer to the CouchDB HTTP API documentation
// for more information.
//
// See http://pouchdb.com/api.html#fetch_document
// and http://docs.couchdb.org/en/latest/api/document/common.html?highlight=doc#get--db-docid
func (db *PouchDB) Get(docId string, doc interface{}, opts Options) error {
	rw := newResultWaiter()
	db.o.Call("get", docId, opts, rw.Done)
	obj, err := rw.ReadResult()
	if err != nil {
		return err
	}
	return convertJSONObject(obj, doc)
}

// Attachment represents document attachments.
// This structure is borrowed from http://godoc.org/github.com/fjl/go-couchdb#Attachment
type Attachment struct {
	Name string    // File name
	Type string    // MIME type of the Body
	MD5  []byte    // MD5 checksum of the Body
	Body io.Reader // The body itself
}

// PutAttachment creates or updates an attachment. To create an attachment
// on a non-existing document, pass an empty rev.
//
// See http://pouchdb.com/api.html#save_attachment and
// http://godoc.org/github.com/fjl/go-couchdb#DB.PutAttachment
func (db *PouchDB) PutAttachment(docid string, att *Attachment, rev string) (newrev string, err error) {
	rw := newResultWaiter()
	db.o.Call("putAttachment", docid, att.Name, attachmentObject(att), att.Type, rw.Done)
	return rw.ReadRev()
}

// attachmentObject converts an io.Reader to a JavaScrpit Buffer in node, or
// a Blob in the browser
func attachmentObject(att *Attachment) *js.Object {
	buf := new(bytes.Buffer)
	buf.ReadFrom(att.Body)
	if buffer := js.Global.Get("Buffer"); jsbuiltin.TypeOf(buffer) == "function" {
		// The Buffer type is supported, so we'll use that
		return buffer.New(buf.String())
	}
	// We must be in the browser, so return a Blob instead
	return js.Global.Get("Blob").New([]string{buf.String()}, att.Type)
}

func attachmentFromPouch(name string, obj *js.Object) *Attachment {
	att := &Attachment{
		Name: name,
	}
	var body string
	if jsbuiltin.TypeOf(obj.Get("write")) == "function" {
		// This looks like a Buffer object; we're in node
		body = obj.Call("toString", "utf-8").String()
	} else {
		// We're in the browser
		body = obj.String()
		att.Type = obj.Get("type").String()
	}
	att.Body = strings.NewReader(body)
	return att
}

// Attachment retrieves an attachment. The rev argument can be left empty to
// retrieve the latest revision. The caller is responsible for closing the
// attachment's Body if the returned error is nil.
//
// Note that PouchDB's getDocument() does not fetch meta data (except for the
// MIME type in the browser only), so the MD5 sum and (in node) the content
// type files will be empty.
//
// See http://pouchdb.com/api.html#get_attachment and
// http://godoc.org/github.com/fjl/go-couchdb#Attachment
func (db *PouchDB) Attachment(docid, name, rev string) (*Attachment, error) {
	opts := Options{}
	if len(rev) > 0 {
		opts["rev"] = rev
	}
	rw := newResultWaiter()
	db.o.Call("getAttachment", docid, name, opts, rw.Done)
	obj, err := rw.Read()
	if err != nil {
		return nil, err
	}
	x := attachmentFromPouch(name, obj)
	return x, nil
}

func (db *PouchDB) DeleteAttachment(docid, name, rev string) (newrev string, err error) {
	rw := newResultWaiter()
	db.o.Call("removeAttachment", docid, name, rev, rw.Done)
	return rw.ReadRev()
}

// Remove will delete the document. The document must specify both _id and
// _rev. On success, it returns the _rev of the new document with _delete set
// to true.
//
// See: http://pouchdb.com/api.html#delete_document
func (db *PouchDB) Remove(doc interface{}, opts Options) (newrev string, err error) {
	var convertedDoc interface{}
	convertJSONObject(doc, &convertedDoc)
	rw := newResultWaiter()
	db.o.Call("remove", convertedDoc, opts, rw.Done)
	return rw.ReadRev()
}

// BulkDocs will create, update or delete multiple documents.
// See: http://pouchdb.com/api.html#batch_create
func (db *PouchDB) BulkDocs(docs interface{}, opts Options) ([]Result, error) {
	s := reflect.ValueOf(docs)
	if s.Kind() != reflect.Slice {
		return nil, errors.New("docs must be a slice")
	}
	convertedDocs := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		convertJSONObject(s.Index(i).Interface(), &(convertedDocs[i]))
	}
	rw := newResultWaiter()
	db.o.Call("bulkDocs", convertedDocs, opts, rw.Done)
	return rw.ReadBulkResults()
}

// AllDocs will fetch multiple documents.
// The output of the query is unmarshalled into the given result. The format
// of the result depends on the options. Please refer to the CouchDB HTTP API
//  documentation for all the possible options that can be set
//
// See http://pouchdb.com/api.html#batch_fetch and
// http://docs.couchdb.org/en/latest/api/database/bulk-api.html#db-all-docs
func (db *PouchDB) AllDocs(result interface{}, opts Options) error {
	rw := newResultWaiter()
	db.o.Call("allDocs", opts, rw.Done)
	obj, err := rw.Read()
	if err != nil {
		return err
	}
	return convertJSObject(obj, &result)
}

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
//
// See: http://pouchdb.com/api.html#view_cleanup
func (db *PouchDB) ViewCleanup() error {
	rw := newResultWaiter()
	db.o.Call("viewCleanup", rw.Done)
	return rw.Error()
}

// Compact triggers a compaction operation in the local or remote database.
//
// See: http://pouchdb.com/api.html#compaction
func (db *PouchDB) Compact(opts Options) error {
	rw := newResultWaiter()
	db.o.Call("compact", opts, rw.Done)
	return rw.Error()
}

// RevsDiff will, given a set of document/revision IDs return the subset of
// those that do not correspond to revisions stored in the database.
// See: http://pouchdb.com/api.html#revisions_diff
// func (db *PouchDB) RevsDiff(diff *js.Object, fn interface{}) {
// 	db.o.Call("revsDiff", diff, fn)
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
