package integration

import (
	"testing"
	"fmt"
)


func TestBasicUploadQuery(t *testing.T) {
	sdURL, err := getStardogUrl()
	if sdURL == "" {
		t.Skip("Integration tests are not configured")
	}

	if err != nil {
		t.Fatalf("Error getting the stardog url %s", err)
	}
	dbName := getRandomDbName("tstdb", 4)
	err = makeDb(sdURL, dbName)
	if err != nil {
		t.Fatalf("Failed to create the the db %s %s", dbName, err)
	}
	txid, err := startTransaction(sdURL, dbName)
	if err != nil {
		t.Fatalf("Failed start the transact %s", err)
	}
	fmt.Printf("\nID %s\n", txid)
	err = postRows(sdURL, dbName, txid, "etc/rows.rdf")
	if err != nil {
		t.Fatalf("Failed to post the rows %s", err)
	}
	err = commitTransaction(sdURL, dbName, txid)
	if err != nil {
		t.Fatalf("Failed start the transaction %s", err)
	}

	results := make([]string, 3)
	for i := range results {
		data, err := queryAll(sdURL, dbName)
		if err != nil {
			t.Fatalf("Query failed %s", err)
		}
		results[i] = data
	}
	for i, _ := range results {
		if results[0] != results[i] {
			fmt.Printf("HERE IS 0 %s", results[0])
			fmt.Printf("HERE IS %d %s", i, results[i])
			t.Fatalf("The servers 0 and %d did not agree on the values", i)
		}
	}
}