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
		plugin := js.Global.Call("require", "/home/jonhall/go/src/github.com/flimzy/go-pouchdb/node_modules/pouchdb-find")
		pouchdb.RegisterPlugin(plugin)
		loaded = true
	}
	return &PouchPluginFind{ db }
}

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

func (e *findError) IndexExists() bool {
	return e.exists
}

func (db *PouchPluginFind) CreateIndex(index Index) *findError {
	i := indexWrapper{ index }
	var jsonIndex map[string]interface{}
	pouchdb.ConvertJSONObject(i, &jsonIndex)
	rw := pouchdb.NewResultWaiter()
	db.Call("createIndex", jsonIndex, rw.Done)
	result, err := rw.ReadResult()
	if err != nil {
		return &findError{ err, false }
	}
	if result["result"] == "exists" {
		return &findError{
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
