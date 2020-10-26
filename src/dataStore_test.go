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
	if err := saveMapConfToDB(); err != nil {
		t.Fatal(err)
	}
	dbStats.BackupFile = "testdata/out.twdb"
	dbStats.BackupConfigOnly = true
	err = backupDB()
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
