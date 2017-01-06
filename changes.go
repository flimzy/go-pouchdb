// +build js

package pouchdb

import (
	"encoding/json"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
)

// ChangesFeed is an iterator for database changes. See https://pouchdb.com/api.html#changes
// On each call to the Next method, the event fields are updated for the current
// event. Next is designed to be used in a for loop:
//
//     feed, err := client.Changes("db", nil)
//         ...
//         for feed.Next() {
//             fmt.Printf("changed: %s", feed.ID)
//         }
//         err = feed.Err()
//         ...
type ChangesFeed struct {
	// DB is the database. Since all events in a _changes feed
	// belong to the same database, this field is always equivalent to the
	// database from the DB.Changes call that created the feed object
	DB *PouchDB

	// ID is the document ID of the current event.
	ID string

	// Deleted is true when the event represents a deleted document.
	Deleted bool

	// Seq is the database update sequence number of the current event.
	// After all items have been processed, set to the last_seq value sent
	// by CouchDB.
	Seq string

	// Changes is the list of the document's leaf revisions.
	Changes []struct {
		Rev string
	}

	// The document. This is populated only if the feed option
	// "include_docs" is true.
	Doc json.RawMessage

	closed    bool
	err       error
	changes   *js.Object
	sendEvent chan<- event
	readEvent <-chan event
	results   []*js.Object
}

// Close closes the changesfeed and frees up any allocated resources. Note that
// this should be called even for one-shot feed operations.
func (cf *ChangesFeed) Close() error {
	fmt.Printf("Cancelling\n")
	cf.changes.Call("cancel")
	close(cf.sendEvent)
	cf.closed = true
	return nil
}

func (cf *ChangesFeed) Err() error {
	return cf.err
}

// Next populates the feed with the next value. Note that order is not
// guaranteed. If you need events in a specific order, please refer to the
// sequence ID.
func (cf *ChangesFeed) Next() bool {
	if cf.closed {
		return false
	}
	if cf.results != nil {
		if cf.nextResult() {
			// If false, continue, in case we received events out of order
			return true
		}
	}
	e := <-cf.readEvent

	switch e.EventType {
	case "change":
		return cf.next(e.Object)
	case "complete":
		r := e.Object.Get("results")
		cf.results = make([]*js.Object, r.Length())
		for i := 0; i < r.Length(); i++ {
			cf.results[i] = r.Index(i)
		}
		// If false, it means there were no results, so there's no chance of
		// having received events out of order.
		return cf.nextResult()
	case "error":
		cf.err = &js.Error{Object: e.Object}
		return false
	}

	return false
}

func (cf *ChangesFeed) next(o *js.Object) bool {
	cf.ID = o.Get("id").String()
	cf.Deleted = o.Get("deleted").Bool()
	cf.Seq = o.Get("seq").String()
	if changes := o.Get("changes"); jsbuiltin.TypeOf(changes) != "undefined" {
		cf.Changes = make([]struct {
			Rev string
		}, changes.Length())
		for i := 0; i < changes.Length(); i++ {
			cf.Changes[i].Rev = changes.Index(i).Get("rev").String()
		}
	}

	return true
}

func (cf *ChangesFeed) nextResult() bool {
	if len(cf.results) == 0 {
		cf.results = nil
		return false
	}
	var o *js.Object
	o, cf.results = cf.results[0], cf.results[1:]
	return cf.next(o)
}

type event struct {
	EventType string
	Object    *js.Object
}

func (db *PouchDB) Changes(opts Options) (*ChangesFeed, error) {
	eChan := make(chan event)
	cf := &ChangesFeed{
		DB:        db,
		changes:   db.Call("changes", opts.compile()),
		sendEvent: eChan,
		readEvent: eChan,
	}
	cf.changes.Call("on", "change", func(o *js.Object) {
		go cf.handleEvent("change", o)
	})
	// cf.changes.Call("on", "complete", func(o *js.Object) {
	// 	go cf.handleEvent("complete", o)
	// })
	cf.changes.Call("on", "error", func(o *js.Object) {
		go cf.handleEvent("error", o)
	})
	return cf, nil
}

func (cf *ChangesFeed) handleEvent(eventType string, o *js.Object) {
	cf.sendEvent <- event{
		EventType: eventType,
		Object:    o,
	}
}
