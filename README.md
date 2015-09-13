[![Build Status](https://travis-ci.org/flimzy/go-pouchdb.svg?branch=master)](https://travis-ci.org/flimzy/go-pouchdb)

# go-pouchdb

[GopherJS](http://www.gopherjs.org/) bindings for [PouchDB](http://pouchdb.com/).

## Requirements

This package requires PouchDB 4.0.2 or newer. This is due to a bug in earlier versions of PouchDB, which crashed when destroy() was passed with options. See [issue #4323](https://github.com/pouchdb/pouchdb/issues/4323) for details.  While I would normally try to support older versions simultaneously, I suspect anyone using this package is likely to be using the latest version of PouchDB, so I expect this not to be an issue.  If you have a requirement to use this with an older version of PouchDB, and this bug is a problem for you, please open an issue, and I will try to find a work-around.

## General philosophy

There are a number of design decisions that must be made when writing wrappers and/or bindings between langauges.  One common approach is to write the smallest, most minimal wrapper necessary around. The [JavaScript bindings in GopherJS](https://github.com/gopherjs/gopherjs/blob/master/js/js.go) take this approach, and I believe that is an appropriate decision in that case, where the goal is to provide access to the underlying JavaScript primatives, with the least amount of abstraction so that users of the bindings can accomplish anything they need.

I have taken a different approach with this package.  Rather than attempting to provide a minimal wrapper around the existing PouchDB, I have decided to take the approach that my goal is not to wrap PouchDB, but rather to provide client-side database access for GopherJS programs. In other words, I want to accomplish the same goal as PouchDB, but in client-side Go, rather than in client-side JavaScript.

What this means in concrete terms is that much of this API doesn't look very PouchDB-ish. And it doens't look very JavaScript-ish.  You won't find any callbacks here! Under the hood, PouchDB still uses callbacks, but go-pouchdb abstracts that away, to give you a much more Go-idiomatic interface. Many functions have been renamed from their PouchDB versions, to be more Go-idiomatic, as well. I have followed the lead of Felix Lange, author of the popular Go CouchDB library [http://godoc.org/github.com/fjl/go-couchdb](http://godoc.org/github.com/fjl/go-couchdb), and whenever there has been a choice between the PouchDB name and the go-couchdb name, I have chosen the latter.

This decision may mean that anyone familiar with PouchDB in JavaScript may have a slightly steeper learning curve when using this library. But I think that's a small price to pay when considering that anyone already familiar with Go will have a much easier time. And it is primarily the latter group of people (which includes myself!) to which I a hope to cater with this library.

## Status

This is a work in progress. Use at your own risk. Please contribute pull requests!

The following table shows the status of each API method.

| PouchDB name       | go-pouchdb signature(s)                                                                  | Comments
|--------------------|------------------------------------------------------------------------------------------|-----------
| new()              | New(db_name string) *PouchDB                                                             |
|                    | NewFromOpts(Options) *PouchDB                                                            |
| destroy()          | (db \*PouchDB) Destroy(Options) error                                                    |
| put()              | (db \*PouchDB) Put(doc interface{}) (newrev string, err error)                           |
| get()              | (db \*PouchDB) Get(id string, doc interface{}, opts Options) error                       |
| remove()           | (db \*PouchDB) Remove(doc interface{}, opts Options) (newrev string, err error)          |
| bulkDocs()         | (db \*PouchDB) BulkDocs(docs interface{}, opts Options) ([]Result, error)                |
| allDocs()          | (db \*PouchDB) AllDocs(result interface{}, opts Options) error                           |
| viewCleanup()      | (db \*PouchDB) ViewCleanup() error                                                       |
| info()             | (db \*PouchDB) Info() (\*js.Object, error)                                               |
| compact()          | (db \*PouchDB) Compact(opts Options) error                                               |
| revsDiff()         | --                                                                                       |
| defaults()         | n/a                                                                                      | Pass options to New() instead
| debug.enable()     | Debug(module string)                                                                     |
| debug.disable()    | DebugDisable()                                                                           |
| changes()          | --                                                                                       |
| replicate()        | --                                                                                       |
| sync()             | --                                                                                       |
| putAttachment()    | (db \*PouchDB) PutAttachment(docid string, att \*Attachment, rev string) (string, error) | Only tested in Node
| getAttachment()    | (db \*PouchDB) Attachment(docid, name, rev string) (\*Attachment, error)                 | Only tested in Node
| removeAttachment() | (db \*PouchDB) DeleteAttachment(docid, name, rev string) (string, error)                 |
| query()            | n/a                                                                                      | To be [deprecated](http://pouchdb.com/api.html#query_database).
| on()               | --                                                                                       |
| plugin()           | --                                                                                       |

## Implementation notes

### On the handling of JSON

Go has some spiffy JSON capabilities that don't exist in JavaScript. Of particular note, the `[encoding/json](http://golang.org/pkg/encoding/json/)` package understands special [struct tags](http://stackoverflow.com/q/10858787/13860), and does some handy key-name manipulation for us. However, PouchDB gives us already-parsed JSON objects, which means we can't take advantage of Go's enhanced JSON handling.  To get around this, every document read from PouchDB is first converted back into JSON with the `json.Marshal()` method, then converted back into an object, this time as a native Go object. And when putting documents into PouchDB, the reverse is done. This allows you to take advantage of Go's "superior" (or at least more idiomatic) JSON handling.

Related to this, rather than methods such as `Get()` simply returning an object, as they do in PouchDB, they take a document reference as an argument, and assign the document to this reference. This allows you to provide a prescribed data type into which the document is expected to fit. This is how the popular [fjl/go-couchdb](https://github.com/fjl/go-couchdb) library works, too. In fact, I have copied a lot from fjl in this package.

## License

This software is released under the terms of the MIT license. See LICENCE.md for details.

