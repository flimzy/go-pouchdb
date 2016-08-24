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

	"github.com/flimzy/jsblob"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
)

type PouchDB struct {
	o *js.Object
}

type Options struct {
	// This turns on auto compaction, which means compact() is called after
	// every change to the database. Defaults to false.
	//
	// Used by New(), for local databases only
	AutoCompaction bool

	// One of 'idb', 'leveldb', 'websql', or 'http'. If unspecified, PouchDB
	// will infer this automatically, preferring IndexedDB to WebSQL in
	// browsers that support both.
	//
	// Used by New(), for local databases only.
	Adaptor string

	// See https://github.com/Level/levelup#options
	DB *js.Object

	// Specify how many old revisions we keep track (not a copy) of.
	//
	// Used by New(), for local databases only.
	RevsLimit int

	// Remote databases only.) Ajax requester options. These are passed
	// verbatim to request (in Node.js) or a request shim (in the browser),
	// with the exception of 'cache'.
	//  - cache: Appends a random string to the end of all HTTP GET requests
	//    to avoid them being cached on IE. Set this to true to prevent this
	//    happening.
	//  - headers: Allows you to customise headers that are sent to the remote
	//    HTTP Server.
	//  - username + password: Alternate method to provide auth credentians to
	//    using a database name in the form `http://user:pass@host/name`
	//  - withCredentials: Set to false to disable transferring cookies or HTTP
	//    Auth information. Defaults to true.
	//  - skip_setup: Initially PouchDB checks if the database exists, and
	//    tries to create it, if it does not exist yet. Set this to true to skip
	//    this setup.
	//
	// Used by any method that accesses a remote database.
	Ajax map[string]interface{}

	// Specifies whether you want to use persistent or temporary storage.
	//
	// Used by New(), for IndexedDB only.
	Storage string

	// Amount in MB to request for storage, which you will need if you are
	// storing >5MB in order to avoid storage limit errors on iOS/Safari.
	//
	// Used by New(), for WebSQL only.
	Size int

	// Fetch specific revision of a document. Defaults to winning revision.
	//
	// Used by Get().
	Rev string

	// Include revision history of the document.
	//
	// Used by Get().
	Revs bool

	// Include a list of revisions of the document, and their availability.
	//
	// Used by Get().
	RevsInfo bool

	// Fetch requested leaf revisions in the order specified.
	//
	// Used by Get().
	OpenRevs []string

	// Fetch all leaf revisions.
	//
	// Used by Get().
	AllOpenRevs bool

	// If specified, conflicting leaf revisions will be attached in _conflicts
	// array.
	//
	// Used by Get(), AllDocs() and Query().
	Conflicts bool

	// Include attachment data.
	//
	// Used by Get(), AllDocs() and Query().
	Attachments bool

	// Include the document itself in each row in the doc field. The default
	// is to only return the _id and _rev properties.
	//
	// Used by AllDocs() and Query().
	IncludeDocs bool

	// Get documents with IDs in a certain range (inclusive/inclusive).
	//
	// Used by AllDocs() and Query().
	StartKey string
	EndKey   string

	// Exclude documents having an ID equal to the given EndKey.
	// Note this flag has the reverse sense of the PouchDB inclusive_end flag.
	//
	// Used by AllDocs() and Query().
	ExclusiveEnd bool

	// Maximum number of documents to return.
	//
	// Used by AllDocs() and Query()
	Limit int

	// Number of docs to skip before returning.
	//
	// Used by AllDocs() and Query().
	Skip int

	// Reverse the order of the output documents. Note that the order of
	// StartKey and EndKey is reversed when Descending is true.
	//
	// Used by AllDocs() and Query().
	Descending bool

	// Only return documents with IDs matching this string.
	//
	// Used by AllDocs() and Query().
	Key string

	// Array of string keys to fetch in a single shot.
	//  - Neither StartKey nor EndKey can be specified with this option.
	//  - The rows are returned in the same order as the supplied keys array.
	//  - The row for a deleted document will have the revision ID of the
	//    deletion, and an extra key "deleted":true in the value property.
	//  - The row for a nonexistent document will just contain an "error"
	//    property with the value "not_found".
	//  - For details, see the CouchDB query options documentation:
	//    http://wiki.apache.org/couchdb/HTTP_view_API#Querying_Options
	//
	// Used by AllDocs() and Query().
	Keys []string

	// Reference a filter function from a design document to selectively get
	// updates. To use a view function, pass _view here and provide a reference
	// to the view function in options.view.
	// See also: https://pouchdb.com/api.html#filtered-replication
	//
	// Used by Replicate()
	Filter string

	// Only show changes for docs with these ids (array of strings).
	//
	// Used by Replicate().
	DocIDs []string

	// Object containing properties that are passed to the filter function,
	// e.g. {"foo":"bar"}.
	//
	// Used by Replicate().
	QueryParams map[string]interface{}

	// Specify a view function (e.g. 'design_doc_name/view_name' or 'view_name'
	// as shorthand for 'view_name/view_name') to act as a filter.
	//
	// Used by Replicate().
	View string

	// Replicate changes after the given sequence number.
	//
	// Used by Replicate().
	Since int64

	Heartbeat       int64
	Timeout         int64
	BatchSize       int
	BatchesLimit    int
	BackOffFunction func(int) int

	// The name of a view in an existing design document (e.g.
	// 'mydesigndoc/myview', or 'myview' as a shorthand for 'myview/myview').
	//
	// Used by Query().
	MapFuncName string

	// A JavaScript object representing a map function. It is not possible to
	// define map functions in GopherJS.
	//
	// Used by Query()
	MapFunc *js.Object

	// The name of a built-in reduce function: '_sum', '_count', or '_stats'.
	//
	// Used by Query()
	ReduceFuncName string

	// A JavaScript object representing a reduce function. It is not possible
	// to define reduce functions in GopherJS.
	//
	// Used by Query()
	ReduceFunc *js.Object

	// True if you want the reduce function to group results by keys, rather
	// than returning a single result.
	//
	// Used by Query().
	Group bool

	// Number of elements in a key to group by, assuming the keys are arrays.
	// Defaults to the full length of the array.
	//
	// Used by Query().
	GroupLevel int

	// Only applies to saved views. Can be one of:
	//  - unspecified (default): Returns the latest results, waiting for the
	//    view to build if necessary.
	//  - 'ok': Returns results immediately, even if they’re out-of-date.
	//  - 'update_after': Returns results immediately, but kicks off a build
	//    afterwards.
	//
	// Used by Query().
	Stale string
}

func (o *Options) compile() map[string]interface{} {
	opts := make(map[string]interface{})
	if o.AutoCompaction {
		opts["auto_compaction"] = true
	}
	if o.Adaptor != "" {
		opts["adaptor"] = o.Adaptor
	}
	if o.RevsLimit > 0 {
		opts["revs_limit"] = o.RevsLimit
	}
	if o.Ajax != nil {
		opts["ajax"] = o.Ajax
	}
	if o.Storage != "" {
		opts["storage"] = o.Storage
	}
	if o.Size > 0 {
		opts["size"] = o.Size
	}
	if o.Rev != "" {
		opts["rev"] = o.Rev
	}
	if o.Revs {
		opts["revs"] = true
	}
	if o.RevsInfo {
		opts["revs_info"] = true
	}
	if o.AllOpenRevs {
		opts["open_revs"] = "all"
	} else if len(o.OpenRevs) > 0 {
		opts["open_revs"] = o.OpenRevs
	}
	if o.Conflicts {
		opts["conflicts"] = true
	}
	if o.Attachments {
		opts["attachments"] = true
	}
	if o.IncludeDocs {
		opts["include_docs"] = true
	}
	if o.StartKey != "" {
		opts["startkey"] = o.StartKey
	}
	if o.EndKey != "" {
		opts["endkey"] = o.EndKey
	}
	if o.ExclusiveEnd {
		opts["inclusive_end"] = false
	}
	if o.Limit > 0 {
		opts["limit"] = o.Limit
	}
	if o.Skip > 0 {
		opts["skip"] = o.Skip
	}
	if o.Descending {
		opts["descending"] = true
	}
	if o.Key != "" {
		opts["key"] = o.Key
	}
	if len(o.Keys) > 0 {
		opts["keys"] = o.Keys
	}
	if o.Filter != "" {
		opts["filter"] = o.Filter
	}
	if len(o.DocIDs) > 0 {
		opts["doc_ids"] = o.DocIDs
	}
	if o.QueryParams != nil {
		opts["query_params"] = o.QueryParams
	}
	if o.View != "" {
		opts["view"] = o.View
	}
	if o.Since > 0 {
		opts["since"] = o.Since
	}
	if o.Heartbeat > 0 {
		opts["heartbeat"] = o.Heartbeat
	}
	if o.Timeout > 0 {
		opts["timeout"] = o.Timeout
	}
	if o.BatchSize > 0 {
		opts["batch_size"] = o.BatchSize
	}
	if o.BatchesLimit > 0 {
		opts["batches_limit"] = o.BatchesLimit
	}
	if o.BackOffFunction != nil {
		opts["back_off_function"] = o.BackOffFunction
	}
	if o.MapFuncName != "" {
		opts["fun"] = o.MapFuncName
	} else if o.MapFunc != nil {
		opts["fun"] = o.MapFunc
	}
	if o.ReduceFuncName != "" {
		opts["reduce"] = o.ReduceFuncName
	} else if o.ReduceFunc != nil {
		opts["reduce"] = o.ReduceFunc
	}
	if o.Group {
		opts["group"] = true
	}
	if o.GroupLevel > 0 {
		opts["group_level"] = o.GroupLevel
	}
	if o.Stale != "" {
		opts["stale"] = o.Stale
	}
	return opts
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
	if GlobalPouch != nil && jsbuiltin.TypeOf(GlobalPouch) != "undefined" {
		return GlobalPouch
	}
	GlobalPouch := js.Global.Get("PouchDB")
	if jsbuiltin.TypeOf(GlobalPouch) == "undefined" {
		GlobalPouch = js.Global.Call("require", "pouchdb")
	}
	return GlobalPouch
}

// Plugin registers a loaded plugin with the global PouchDB object
func Plugin(plugin *js.Object) {
	globalPouch().Call("plugin", plugin)
}

// Debug enables debugging for the specified module.
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
