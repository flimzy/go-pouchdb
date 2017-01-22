package pouchdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/flimzy/jsblob"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
)

type PouchDB struct {
	o *js.Object
}

type Result map[string]interface{}

type DBInfo struct {
	DBName    string `json:"db_name"`
	DocCount  uint   `json:"doc_count"`
	UpdateSeq uint64 `json:"update_seq"`
}

// GlobalPouch is the global pouchdb object. The package will look for it in
// the global object (js.Global), or try to require it if it is not found. If
// this does not work for you, you ought to set it explicitly yourself:
//
//    pouchdb.GlobalPouch = js.Global.Call("require", "/path/to/your/copy/of/pouchdb")
var GlobalPouch *js.Object

func globalPouch() *js.Object {
	if GlobalPouch != nil && GlobalPouch != js.Undefined {
		return GlobalPouch
	}
	GlobalPouch = js.Global.Get("PouchDB")
	if GlobalPouch == js.Undefined {
		GlobalPouch = js.Global.Call("require", "pouchdb")
	}
	return GlobalPouch
}

// Plugin registers a loaded plugin with the global PouchDB object
func Plugin(plugin *js.Object) {
	globalPouch().Call("plugin", plugin)
}

// Debug enables debugging for the specified module. Note this only affects
// connections made after this is run.
// See: http://pouchdb.com/api.html#debug_mode
func Debug(module string) {
	globalPouch().Get("debug").Call("enable", module)
}

// DebugDisable disables debugging.
func DebugDisable() {
	globalPouch().Get("debug").Call("disable")
}

// New creates a database or opens an existing one.
// See: http://pouchdb.com/api.html#create_database
func New(db_name string) *PouchDB {
	return &PouchDB{globalPouch().New(db_name)}
}

// NewWithOpts creates a database or opens an existing one.
// See: http://pouchdb.com/api.html#create_database
func NewWithOpts(db_name string, opts Options) *PouchDB {
	return &PouchDB{globalPouch().New(db_name, opts.compile())}
}

// Info fetches information about a database.
//
// See: http://pouchdb.com/api.html#database_information
func (db *PouchDB) Info() (DBInfo, error) {
	rw := NewResultWaiter()
	db.Call("info", rw.Done)
	result, err := rw.ReadResult()
	if err != nil {
		return DBInfo{}, err
	}
	var dbinfo DBInfo
	err = ConvertJSONObject(result, &dbinfo)
	return dbinfo, err
}

// Deestroy will delete the database.
// See: http://pouchdb.com/api.html#delete_database
func (db *PouchDB) Destroy(opts Options) error {
	rw := NewResultWaiter()
	db.Call("destroy", opts.compile(), rw.Done)
	return rw.Error()
}

// ConvertJSONObject takes an intterface{} and runs it through json.Marshal()
// and json.Unmarshal() so that any struct tags will be applied.
func ConvertJSONObject(input, output interface{}) error {
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
func ConvertJSObject(jsObj *js.Object, output interface{}) error {
	return ConvertJSONObject(jsObj.Interface(), output)
}

// Put will create a new document or update an existing document.
// See: http://pouchdb.com/api.html#create_document
func (db *PouchDB) Put(doc interface{}) (newrev string, err error) {
	var convertedDoc interface{}
	ConvertJSONObject(doc, &convertedDoc)
	rw := NewResultWaiter()
	db.Call("put", convertedDoc, rw.Done)
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
	rw := NewResultWaiter()
	db.Call("get", docId, opts.compile(), rw.Done)
	obj, err := rw.ReadResult()
	if err != nil {
		return err
	}
	return ConvertJSONObject(obj, doc)
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
	rw := NewResultWaiter()
	db.Call("putAttachment", docid, att.Name, rev, attachmentObject(att), att.Type, rw.Done)
	return rw.ReadRev()
}

// attachmentObject converts an io.Reader to a JavaScript Buffer in node, or
// a Blob in the browser
func attachmentObject(att *Attachment) *js.Object {
	buf := new(bytes.Buffer)
	buf.ReadFrom(att.Body)
	if buffer := js.Global.Get("Buffer"); jsbuiltin.TypeOf(buffer) == "function" {
		// The Buffer type is supported, so we'll use that
		return buffer.New(buf.String())
	}
	// We must be in the browser, so return a Blob instead
	return js.Global.Get("Blob").New([]interface{}{buf.Bytes()}, map[string]string{"type": att.Type})
}

func attachmentFromPouch(name string, obj *js.Object) *Attachment {
	att := &Attachment{
		Name: name,
	}
	var body string
	if jsbuiltin.TypeOf(obj.Get("write")) == "function" {
		// This looks like a Buffer object; we're in node
		body = obj.Call("toString", "utf-8").String()
		att.Body = strings.NewReader(body) // FIXME: bytes, not string
	} else {
		// We're in the browser
		att.Type = obj.Get("type").String()
		blob := jsblob.Blob{*obj}
		att.Body = bytes.NewReader(blob.Bytes())
	}
	return att
}

// Attachment retrieves an attachment. The rev argument can be left empty to
// retrieve the latest revision. The caller is responsible for closing the
// attachment's Body if the returned error is nil.
//
// Note that PouchDB's getDocument() does not fetch meta data (except for the
// MIME type in the browser only), so the MD5 sum and (in node) the content
// type fields will be empty.
//
// See http://pouchdb.com/api.html#get_attachment and
// http://godoc.org/github.com/fjl/go-couchdb#Attachment
func (db *PouchDB) Attachment(docid, name, rev string) (*Attachment, error) {
	opts := Options{
		Rev: rev,
	}
	rw := NewResultWaiter()
	db.Call("getAttachment", docid, name, opts.compile(), rw.Done)
	obj, err := rw.Read()
	if err != nil {
		return nil, err
	}
	x := attachmentFromPouch(name, obj)
	return x, nil
}

func (db *PouchDB) DeleteAttachment(docid, name, rev string) (newrev string, err error) {
	rw := NewResultWaiter()
	db.Call("removeAttachment", docid, name, rev, rw.Done)
	return rw.ReadRev()
}

// Remove will delete the document. The document must specify both _id and
// _rev. On success, it returns the _rev of the new document with _delete set
// to true.
//
// See: http://pouchdb.com/api.html#delete_document
func (db *PouchDB) Remove(doc interface{}, opts Options) (newrev string, err error) {
	var convertedDoc interface{}
	ConvertJSONObject(doc, &convertedDoc)
	rw := NewResultWaiter()
	db.Call("remove", convertedDoc, opts.compile(), rw.Done)
	return rw.ReadRev()
}

// BulkDocs will create, update or delete multiple documents.
//
// See: http://pouchdb.com/api.html#batch_create
func (db *PouchDB) BulkDocs(docs interface{}, opts Options) ([]Result, error) {
	s := reflect.ValueOf(docs)
	if s.Kind() != reflect.Slice {
		return nil, errors.New("docs must be a slice")
	}
	convertedDocs := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ConvertJSONObject(s.Index(i).Interface(), &(convertedDocs[i]))
	}
	rw := NewResultWaiter()
	db.Call("bulkDocs", convertedDocs, opts.compile(), rw.Done)
	return rw.ReadBulkResults()
}

// AllDocs will fetch multiple documents.
// The output of the query is unmarshalled into the given result. The format
// of the result depends on the options. Please refer to the CouchDB HTTP API
// documentation for all the possible options that can be set.
//
// See http://pouchdb.com/api.html#batch_fetch and
// http://docs.couchdb.org/en/latest/api/database/bulk-api.html#db-all-docs
func (db *PouchDB) AllDocs(result interface{}, opts Options) error {
	rw := NewResultWaiter()
	db.Call("allDocs", opts.compile(), rw.Done)
	obj, err := rw.Read()
	if err != nil {
		return err
	}
	return ConvertJSObject(obj, &result)
}

// Invoke a map/reduce function, which allows you to perform more complex
// queries on PouchDB than what you get with allDocs().
//
// See http://pouchdb.com/api.html#query_database
func (db *PouchDB) Query(view string, result interface{}, opts Options) error {
	rw := NewResultWaiter()
	db.Call("query", view, opts.compile(), rw.Done)
	obj, err := rw.Read()
	if err != nil {
		return err
	}
	return ConvertJSObject(obj, &result)
}

type MapFunc func(string)

func (db *PouchDB) QueryFunc(fn MapFunc, result interface{}, opts Options) error {
	rw := NewResultWaiter()
	db.Call("query", fn, opts.compile(), rw.Done)
	obj, err := rw.Read()
	if err != nil {
		return err
	}
	return ConvertJSObject(obj, &result)
}

// Replicate will replicate data from source to target in the foreground.
// For "live" replication use ReplicateLive()
// See: http://pouchdb.com/api.html#replication
func Replicate(source, target *PouchDB, opts Options) (Result, error) {
	rw := NewResultWaiter()
	repl := globalPouch().Call("replicate", source, target, opts.compile())
	repl.Call("then", func(r *js.Object) {
		rw.Done(nil, r)
	})
	repl.Call("catch", func(e *js.Object) {
		rw.Done(e, nil)
	})
	return rw.ReadResult()
}

// Sync data from src to target and target to src. This is a convenience method for bidirectional data replication.
//
// See http://pouchdb.com/api.html#sync
func Sync(source, target *PouchDB, opts Options) ([]Result, error) {
	results := make([]Result, 2)
	result, err := Replicate(source, target, opts)
	results[0] = result
	if err != nil {
		return results, err
	}
	result, err = Replicate(target, source, opts)
	results[1] = result
	return results, err
}

// Replicate will replicate data from source to target in the background.
// This method returns a *ChangeFeed which can be used to monitor progress
// in a Go routine. For foreground sync, use Replicate().
// See: http://pouchdb.com/api.html#replication

// ViewCleanup cleans up any stale map/reduce indexes.
//
// See: http://pouchdb.com/api.html#view_cleanup
func (db *PouchDB) ViewCleanup() error {
	rw := NewResultWaiter()
	db.Call("viewCleanup", rw.Done)
	return rw.Error()
}

// Compact triggers a compaction operation in the local or remote database.
//
// See: http://pouchdb.com/api.html#compaction
func (db *PouchDB) Compact(opts Options) error {
	rw := NewResultWaiter()
	db.Call("compact", opts, rw.Done)
	return rw.Error()
}

// RevsDiff will, given a set of document/revision IDs return the subset of
// those that do not correspond to revisions stored in the database.
// See: http://pouchdb.com/api.html#revisions_diff
// func (db *PouchDB) RevsDiff(diff *js.Object, fn interface{}) {
// 	db.Call("revsDiff", diff, fn)
// }

// Call calls the underlying PouchDB object's method with the given name and
// arguments. This method is used internally, and may also facilitate the use
// of plugins which may add methods to PouchDB which are not implemented in
// the GopherJS bindings.
func (db *PouchDB) Call(name string, args ...interface{}) *js.Object {
	return db.o.Call(name, args...)
}

// GetJS gets the requested key from the underlying PouchDB object
func (db *PouchDB) GetJS(name string) *js.Object {
	return db.o.Get(name)
}

// OnCreate registers the function as an event listener for the 'created' event.
// See https://pouchdb.com/api.html#events
func OnCreate(fn func(dbName string)) {
	globalPouch().Call("on", "created", func(dbname string) {
		go fn(dbname)
	})
}

// OnDestroy registers the function as an event listener for the 'destroyed'
// event. See https://pouchdb.com/api.html#events
func OnDestroy(fn func(dbName string)) {
	globalPouch().Call("on", "destroyed", func(dbname string) {
		go fn(dbname)
	})
}
