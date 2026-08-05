package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDBEngine(t *testing.T) {
	cases := map[string]DBEngine{
		"pg":      DBPostgres,
		"mariadb": DBMySQL,
		"sqlite3": DBSQLite,
	}
	for input, expected := range cases {
		actual, err := ParseDBEngine(input)
		if err != nil || actual != expected {
			t.Fatalf("%s: %v %v", input, actual, err)
		}
	}
}

func TestFileSHA256(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(path, []byte("PulseNet"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := FileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	if hash != "202f755b403cbf7376ac93421c9e964d78441fbccce58744752ba2f3a134d9f2" {
		t.Fatalf("hash = %s", hash)
	}
}
