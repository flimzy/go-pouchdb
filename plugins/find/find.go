// +build js

package pouchdb_find

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"

	"github.com/flimzy/go-pouchdb"
)

type PouchPluginFind struct {
	*pouchdb.PouchDB
}

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
