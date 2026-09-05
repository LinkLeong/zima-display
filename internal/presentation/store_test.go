package presentation

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateTextPersistsPages(t *testing.T) {
	store := New(t.TempDir())
	manifest, err := store.CreateText("Demo", "demo.md", []string{"First", "Second"})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Get(manifest.ID)
	if err != nil {
		t.Fatal(err)
	}
	pages, err := store.TextPages(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || pages[0] != "First\n" || pages[1] != "Second\n" {
		t.Fatalf("unexpected pages %#v", pages)
	}
}

func TestImportZipRejectsPathTraversal(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "bad.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("../escape.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte("not-an-image"))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	store := New(t.TempDir())
	if _, err := store.ImportZip(archivePath, "Bad", ""); err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}
