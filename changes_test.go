// +build js

package pouchdb

import (
	"fmt"
	"strings"
	"testing"
)

func TestNoChanges(t *testing.T) {
	db := newPouch("changes")
	cf, err := db.Changes(Options{})
	if err != nil {
		t.Fatalf("Error opening changes feed: %s", err)
	}
	var count int
	expectedCount := 0
	for cf.Next() {
		count++
	}
	if count != expectedCount {
		t.Errorf("Got %d results, expected %d\n", count, expectedCount)
	}
	if err = cf.Err(); err != nil {
		t.Fatalf("Error reading changes: %s", err)
	}
}

func feedsEqual(f1, f2 *ChangesFeed) bool {
	if f1.ID != f2.ID {
		return false
	}
	if f1.Deleted != f2.Deleted {
		return false
	}
	if f1.Seq != f2.Seq {
		return false
	}
	if len(f1.Changes) != len(f2.Changes) {
		return false
	}
	for i := range f1.Changes {
		parts1 := strings.Split(f1.Changes[i].Rev, "-")
		parts2 := strings.Split(f2.Changes[i].Rev, "-")
		if parts1[0] != parts2[0] {
			return false
		}
	}
	return true
}

func TestChanges(t *testing.T) {
	db := newPouch("changes")
	cf, err := db.Changes(Options{})
	if err != nil {
		t.Fatalf("Error opening changes feed: %s", err)
	}
	var count int
	expectedCount := 1
	expectedResults := []*ChangesFeed{
		&ChangesFeed{
			DB:      db,
			ID:      "foobar",
			Deleted: true,
			Changes: []struct {
				Rev string
			}{
				struct {
					Rev string
				}{
					Rev: "2-xxx",
				},
			},
			Seq: "2",
		},
	}

	doc := map[string]interface{}{
		"_id": "foobar",
		"foo": "bar",
	}
	rev, err := db.Put(doc)
	if err != nil {
		t.Fatalf("Error storing document: %s", err)
	}
	doc["_rev"] = rev
	if _, err = db.Remove(doc, Options{}); err != nil {
		t.Fatalf("Error deleting document: %s", err)
	}

	cf, err = db.Changes(Options{
		Style: AllDocs,
		Limit: expectedCount,
	})
	if err != nil {
		t.Fatalf("Error opening changes feed: %s", err)
	}

	for cf.Next() {
		expected := expectedResults[count]
		if !feedsEqual(expected, cf) {
			fmt.Printf("Result %d not as expected.\n\tExpected: %+v\n\t  Actual: %+v\n", count+1, expected, cf)
		}
		count++
	}
	if count != expectedCount {
		t.Errorf("Got %d results, expected %d\n", count, expectedCount)
	}
	if err = cf.Err(); err != nil {
		t.Fatalf("Error reading changes: %s", err)
	}

}

// func xTestLiveChanges(t *testing.T) {
// 	fmt.Printf("x\n")
// 	db := newPouch("changes")
// 	cf, err := db.Changes(Options{
// 		Live: true,
// 	})
// 	defer cf.Close()
//
// 	fmt.Printf("y\n")
// 	if err != nil {
// 		t.Fatalf("Error opening changes feed: %s", err)
// 	}
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		fmt.Printf("z\n")
// 		for cf.Next() {
// 			fmt.Printf("zz\n")
// 			spew.Dump(cf)
// 		}
// 		fmt.Printf("a\n")
// 		if err := cf.Err(); err != nil {
// 			t.Fatalf("Error reading changes: %s", err)
// 		}
//
// 	}()
// 	fmt.Printf("sleeping\n")
// 	time.Sleep(1 * time.Second)
// 	fmt.Printf("waiting\n")
// 	wg.Wait()
// 	fmt.Printf("Done waiting\n")
// }
