// +build js

package pouchdb_find

import (
	"encoding/json"
	"errors"
	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/go-pouchdb"
)

type PouchPluginFind struct {
	*pouchdb.PouchDB
}

var loaded bool
func New(db *pouchdb.PouchDB) *PouchPluginFind {
	if ! loaded {
		// Load the JS plugin
		js.Global.Call("require", "/home/jonhall/go/src/github.com/flimzy/go-pouchdb/node_modules/pouchdb-find")
		loaded = true
	}
	return &PouchPluginFind{ db }
}

type Index struct {
	Fields []string `json:'fields'`
	Name   string   `json:'name'`
	Ddoc   string   `json:'ddoc'`
	Type   string   `json:'type'`
}

type indexWrapper struct {
	index  Index
}

type findError struct {
	error
	exists bool
}

func (e findError) IndexExists() bool {
	return e.exists
}

func (db *PouchPluginFind) CreateIndex(index Index) error {
	i := indexWrapper{ index }
	rw := pouchdb.NewResultWaiter()
	db.Call("createIndex", i, rw.Done)
	result, err := rw.ReadResult()
	if err != nil {
		return findError{ err, false }
	}
	if result["result"] == "exists" {
		return findError{
			errors.New("Index exists"),
			true,
		}
	}
	return nil
}

type IndexDef struct {
	Ddoc string `json:'ddoc'`
	Name string `json:'name'`
	Type string `json:'type'`
	Def  struct {
		Fields []map[string]string `json:'fields'`
	} `json:'def'`
}

type indexDefsWrapper struct {
	indexes []IndexDef
}

func (db *PouchPluginFind) GetIndexes() ([]IndexDef, error) {
	rw := pouchdb.NewResultWaiter()
	db.Call("getIndexes", rw.Done)
	result, err := rw.Read()
	if err != nil {
		return nil, err
	}
	var i indexDefsWrapper
	err = pouchdb.ConvertJSONObject(result,&i)
	if err != nil {
		return nil, err
	}
	return i.indexes, nil
}


func (db *PouchPluginFind) DeleteIndex(index IndexDef) error {
	i := indexDefsWrapper{ []IndexDef{ index } }
	rw := pouchdb.NewResultWaiter()
	db.Call("deleteIndex", i, rw.Done)
	_, err := rw.Read()
	return err
}

func (db *PouchPluginFind) JSONFind(request string, doc interface{}) error {
	var jsonRequest interface{}
	err := json.Unmarshal([]byte(request), &jsonRequest)
	if err != nil {
		return err
	}
	rw := pouchdb.NewResultWaiter()
	db.Call("find", jsonRequest, rw.Done)
	result, err := rw.Read()
	if err != nil {
		return err
	}
	err = pouchdb.ConvertJSONObject(result, doc)
	return err
}
