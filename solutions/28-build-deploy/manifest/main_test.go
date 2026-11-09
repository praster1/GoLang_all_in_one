package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestManifest(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string]string{"b.bin": "abc", "a.bin": "", "manifest.json": "old"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"-dir", dir}, &out, &errOut); code != 0 {
		t.Fatalf("code=%d: %s", code, errOut.String())
	}
	var got []entry
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "a.bin" || got[1].Name != "b.bin" || got[1].Size != 3 {
		t.Fatalf("got=%v", got)
	}
	if got[1].SHA256 != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal("hash mismatch")
	}
}

func TestManifestErrors(t *testing.T) {
	var out, errOut bytes.Buffer
	if run(nil, &out, &errOut) != 2 {
		t.Fatal("missing argument")
	}
	if run([]string{"-dir", filepath.Join(t.TempDir(), "missing")}, &out, &errOut) != 1 {
		t.Fatal("missing directory")
	}
	empty, err := manifest(t.TempDir())
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatal("empty directory contract")
	}
}
