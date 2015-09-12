# go-pouchdb

[GopherJS](http://www.gopherjs.org/) bindings for [PouchDB](http://pouchdb.com/).

## Status

This is a work in progress. Use at your own risk. Please contribute pull requests!

## License

This software is released under the terms of the MIT license. See LICENCE.md for details.

## Status

The following table shows the status of each API method.

| PouchDB name     | go-pouchdb signature(s)                      | Tested  | Comments
|------------------|----------------------------------------------|---------|-----------
| new()            | New(db_name string) *PouchDB                 | Yes     |
|                  | NewFromOpts(Options) *PouchDB                | Yes     |
| destroy()        | (db *PouchDB) Destroy() error                | Yes     | No support for 'options' argument. See [issue #4323](https://github.com/pouchdb/pouchdb/issues/4323).
| put()            | --                                           | --      |
| get()            | --                                           | --      |
| bulkDocs()       | --                                           | --      |
| allDocs()        | --                                           | --      |
| viewCleanup()    | --                                           | --      |
| info()           | (db *PouchDB) Info() (*js.Object, error)     | Yes     |
| compact()        | --                                           | --      |
| revsDiff()       | --                                           | --      |
| defaults()       | --                                           | --      |
| debug.enable()   | Debug(module string)                         | Yes     |
| debug.disable()  | DebugDisable()                               | Yes     |
| changes()        | --                                           | --      |
| replicate()      | --                                           | --      |
| sync()           | --                                           | --      |
| getAttachment()  | --                                           | --      |
| query()          | --                                           | --      |
| on()             | --                                           | --      |
| plugin()         | --                                           | --      |
