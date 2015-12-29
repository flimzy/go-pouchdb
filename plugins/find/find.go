// +build js

package pouchdb_find

import (
	"errors"
	"log"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"

	"github.com/flimzy/go-pouchdb"
)

type PouchPluginFind struct {
	*pouchdb.PouchDB
}

// New loads the pouchdb-find plugin (if not already loaded) and returns
// a plugin instance.
func New(db *pouchdb.PouchDB) *PouchPluginFind {
	fnType := jsbuiltin.TypeOf(db.GetJS("createIndex"))
	if fnType == "undefined" {
		// Load the JS plugin
		plugin := js.Global.Call("require", "pouchdb-find")
		pouchdb.Plugin(plugin)
	} else if fnType != "function" {
		log.Fatal("Cannot load pouchdb-find plugin; .createIndex method already exists as a non-function")
	}
	return &PouchPluginFind{db}
}

// Index defines an index to be created
type Index struct {
	Fields []string `json:"fields"`
	Name   string   `json:"name,omitifempty"`
	Ddoc   string   `json:"ddoc,omitifempty"`
	Type   string   `json:"type,omitifempty"`
}

type indexWrapper struct {
	Index Index `json:"index"`
}

type findError struct {
	error
	exists bool
}

// IndexExists returns true if the error indicates that the index to be created
// already exists.
func (e *findError) IndexExists() bool {
	return e.exists
}

// Creates the requested index.
//
// See https://github.com/nolanlawson/pouchdb-find#dbcreateindexindex--callback
func (db *PouchPluginFind) CreateIndex(index Index) *findError {
	i := indexWrapper{index}
	var jsonIndex map[string]interface{}
	pouchdb.ConvertJSONObject(i, &jsonIndex)
	rw := pouchdb.NewResultWaiter()
	db.Call("createIndex", jsonIndex, rw.Done)
	result, err := rw.ReadResult()
	if err != nil {
		return &findError{err, false}
	}
	if result["result"] == "exists" {
		return &findError{
			errors.New("Index exists"),
			true,
		}
	}
	return nil
}

// IndexDef describes an index as fetched from the database
type IndexDef struct {
	Ddoc string `json:"ddoc"`
	Name string `json:"name"`
	Type string `json:"type"`
	Def  struct {
		Fields []map[string]string `json:"fields"`
	} `json:"def"`
}

type indexDefsWrapper struct {
	Indexes []*IndexDef `json:"indexes"`
}

// GetIndex returns a list of existing indexes.
//
// See https://github.com/nolanlawson/pouchdb-find#dbgetindexescallback
func (db *PouchPluginFind) GetIndexes() ([]*IndexDef, error) {
	rw := pouchdb.NewResultWaiter()
	db.Call("getIndexes", rw.Done)
	result, err := rw.Read()
	if err != nil {
		return nil, err
	}
	var i indexDefsWrapper
	err = pouchdb.ConvertJSObject(result, &i)
	if err != nil {
		return nil, err
	}
	return i.Indexes, nil
}

// DeleteIndex deletes the requested index.
//
// See https://github.com/nolanlawson/pouchdb-find#dbdeleteindexindex--callback
func (db *PouchPluginFind) DeleteIndex(index *IndexDef) error {
	var i map[string]interface{}
	err := pouchdb.ConvertJSONObject(index, &i)
	if err != nil {
		return err
	}
	rw := pouchdb.NewResultWaiter()
	db.Call("deleteIndex", i, rw.Done)
	_, err = rw.Read()
	return err
}

// Find performs the requested search query
//
// See https://github.com/nolanlawson/pouchdb-find#dbfindrequest--callback
func (db *PouchPluginFind) Find(request, doc interface{}) error {
	rw := pouchdb.NewResultWaiter()
	db.Call("find", request, rw.Done)
	result, err := rw.Read()
	if err != nil {
		return err
	}
	return pouchdb.ConvertJSObject(result, doc)
}
