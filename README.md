[![Build Status](https://travis-ci.org/flimzy/go-pouchdb.svg?branch=master)](https://travis-ci.org/flimzy/go-pouchdb)

# go-pouchdb

[GopherJS](http://www.gopherjs.org/) bindings for [PouchDB](http://pouchdb.com/).

## Status

This is a work in progress. Use at your own risk. Please contribute pull requests!

The following table shows the status of each API method.

| PouchDB name     | go-pouchdb signature(s)                      | Comments
|------------------|----------------------------------------------|-----------
| new()            | New(db_name string) *PouchDB                 |
|                  | NewFromOpts(Options) *PouchDB                |
| destroy()        | (db *PouchDB) Destroy() error                | No support for 'options' argument. See [issue #4323](https://github.com/pouchdb/pouchdb/issues/4323).
| put()            | (db *PouchDB) Put(doc interface{}) error     |
| get()            | (db *PouchDB) Get(id string, doc interface{}, opts Options) error |
| bulkDocs()       | --                                           |
| allDocs()        | --                                           |
| viewCleanup()    | --                                           |
| info()           | (db *PouchDB) Info() (*js.Object, error)     |
| compact()        | --                                           |
| revsDiff()       | --                                           |
| defaults()       | --                                           |
| debug.enable()   | Debug(module string)                         |
| debug.disable()  | DebugDisable()                               |
| changes()        | --                                           |
| replicate()      | --                                           |
| sync()           | --                                           |
| getAttachment()  | --                                           |
| query()          | --                                           |
| on()             | --                                           |
| plugin()         | --                                           |

## Implementation notes

### On the handling of JSON

Go has some spiffy JSON capabilities that don't exist in JavaScript. Of particular note, the `[encoding/json](http://golang.org/pkg/encoding/json/)` package understands special [struct tags](http://stackoverflow.com/q/10858787/13860), and does some handy key-name manipulation for us. However, PouchDB gives us already-parsed JSON objects, which means we can't take advantage of Go's enhanced JSON handling.  To get around this, every document read from PouchDB is first converted back into JSON with the `json.Marshal()` method, then converted back into an object, this time as a native Go object. And when putting documents into PouchDB, the reverse is done. This allows you to take advantage of Go's "superior" (or at least more idiomatic) JSON handling.

Related to this, rather than methods such as `Get()` simply returning an object, as they do in PouchDB, they take a document reference as an argument, and assign the document to this reference. This allows you to provide a prescribed data type into which the document is expected to fit. This is how the popular [fjl/go-couchdb](https://github.com/fjl/go-couchdb) library works, too. In fact, I have copied a lot from fjl in this package.

## License

This software is released under the terms of the MIT license. See LICENCE.md for details.

