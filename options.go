// +build js

package pouchdb

import "github.com/gopherjs/gopherjs/js"

// Options represents the optional configuration options for a PouchDB operation.
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
	Adapter string

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
	if o.Adapter != "" {
		opts["adapter"] = o.Adapter
	}
	if o.DB != nil {
		opts["db"] = o.DB
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
