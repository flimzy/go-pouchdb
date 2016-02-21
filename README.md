[![Build Status](https://travis-ci.org/flimzy/go-pouchdb.svg?branch=master)](https://travis-ci.org/flimzy/go-pouchdb) [![GoDoc](https://godoc.org/github.com/flimzy/go-pouchdb?status.png)](http://godoc.org/github.com/flimzy/go-pouchdb)

# go-pouchdb

[GopherJS](http://www.gopherjs.org/) bindings for [PouchDB](http://pouchdb.com/).

## Requirements

This package requires PouchDB 4.0.2 or newer.

## Installation

Installation is a two-step process, because there is the Go code, and the node.js code:

### Install Go code

Usually the simplest way to do this is simply to run `gopherjs get` in your Go project's directory, to fetch all of the build requirements simultaneously.  If you want to explicitly install **go-pouchdb**, you can do so with the following command:

    gopherjs get github.com/flimzy/go-pouchdb

### Install Node.js code

For larger projects with multiple Node.js requirements, you will likely want to maintain a **package.json** file, as you would for most Node or JavaScript packages.  A minimal **package.json** file for use with **go-pouchdb** could look like this:

    {
        "name": "app",
        "dependencies": {
            "pouchdb": ">=4.0.2"
        }
    }

With this in place, you can run `npm install` from your package directory to install **pouchdb** and its requirements, with npm.  To directly and explicitly install just pouchdb, you can also use the command `npm install pouchdb`.

### Deployment

When deploying your code for use in a browser, you have a couple of options.  My recommended method is simply to use [Browserify](http://browserify.org/) to bundle your GopherJS app with the nodejs requirements into a single, shippable .js file.  If this does not work for your purposes, the next simplest option is simply to load PouchDB in the browser before loading your GopherJS app.  **go-pouchdb** will look for a global **pouchdb** object when it starts.  If it cannot find one, and it cannot load pouchdb, it will generate a runtime error.

### Alternatives

If you have different requirements, and cannot load pouchdb globally, there is one alternative.  You can load the PouchDB module into a custom variable, and tell **go-pouchdb** by setting the `pouchdb.GlobalPouch` variable in your `init()` function:

    package main
    
    import (
        "github.com/flimzy/go-pouchdb"
        "github.com/gopherjs/gopherjs/js"
    )
    
    func init() {
        pouchdb.GlobalPouch = js.Global.Get("someOtherPouchDBObject")
    }


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
| replicate()        | Replicate(source, target *PouchDB, opts Options) (Result, error)                         | "One-shot" replication only
| replicate.to()     | n/a                                                                                      | Use Replicate()
| replicate.from()   | n/a                                                                                      | Use Replicate()
| sync()             | Sync(source, target *PouchDB, opts Options) ([]Results, error)                           |
| putAttachment()    | (db \*PouchDB) PutAttachment(docid string, att \*Attachment, rev string) (string, error) |
| getAttachment()    | (db \*PouchDB) Attachment(docid, name, rev string) (\*Attachment, error)                 |
| removeAttachment() | (db \*PouchDB) DeleteAttachment(docid, name, rev string) (string, error)                 |
| query()            | (db \*PouchDB) Query(view string, result interface{}, opts Options) error                |
| query()            | (db \*PouchDB) QueryFunc(view MapFunc, result interface{}, opts Options) error           |
| on()               | --                                                                                       |
| plugin()           | Plugin(\*js.Object)                                                                      | *Primarily for internal use
| --                 | (db \*PouchDB) Call(name string, interface{} ...) (\*js.Object, error)                   |

### TODO

- Add support for Changes Feeds (for use by `changes()` and live replication)
- Add support for plugins
- Add support for 'created' and 'destroyed' event handlers (??)

## Implementation notes

### On the handling of JSON

Go has some spiffy JSON capabilities that don't exist in JavaScript. Of particular note, the [encoding/json](http://golang.org/pkg/encoding/json/) package understands special [struct tags](http://stackoverflow.com/q/10858787/13860), and does some handy key-name manipulation for us. However, PouchDB gives us already-parsed JSON objects, which means we can't take advantage of Go's enhanced JSON handling.  To get around this, every document read from PouchDB is first converted back into JSON with the `json.Marshal()` method, then converted back into an object, this time as a native Go object. And when putting documents into PouchDB, the reverse is done. This allows you to take advantage of Go's "superior" (or at least more idiomatic) JSON handling.

Related to this, rather than methods such as `Get()` simply returning an object, as they do in PouchDB, they take a document reference as an argument, and assign the document to this reference. This allows you to provide a prescribed data type into which the document is expected to fit. This is how the popular [fjl/go-couchdb](https://github.com/fjl/go-couchdb) library works, too. In fact, I have copied a lot from fjl in this package.

## License

This software is released under the terms of the Apache 2.0 license. See LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
