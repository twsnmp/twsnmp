package main

import (
	"os"
	"testing"
)

func TestDataStore(t *testing.T) {
	err := openDB("testdata/in.twdb")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("testdata/in.twdb")
	mapConf.MapName = "Test123"
	saveMapConfToDB()
	err = backupDB("testdata/out.twdb", true)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("testdata/out.twdb")
	closeDB()
	mapConf.MapName = ""
	err = openDB("testdata/out.twdb")
	if err != nil {
		t.Fatal(err)
	}
	if mapConf.MapName != "Test123" {
		t.Errorf("Backup MapName = '%s'", mapConf.MapName)
	}
	closeDB()
}
