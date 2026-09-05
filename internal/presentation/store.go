package presentation

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxPages = 500

type Manifest struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Kind      string    `json:"kind"`
	Pages     []string  `json:"pages"`
	Source    string    `json:"source,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	root string
}

func New(root string) *Store {
	return &Store{root: root}
}

func (s *Store) CreateText(title, source string, pages []string) (Manifest, error) {
	if len(pages) == 0 || len(pages) > maxPages {
		return Manifest{}, fmt.Errorf("document must contain between 1 and %d pages", maxPages)
	}
	id, err := newID()
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{ID: id, Title: cleanTitle(title), Kind: "text", Source: source, CreatedAt: time.Now().UTC()}
	directory := filepath.Join(s.root, id)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return Manifest{}, err
	}
	for index, content := range pages {
		name := fmt.Sprintf("%03d.txt", index+1)
		if err := os.WriteFile(filepath.Join(directory, name), []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
			_ = os.RemoveAll(directory)
			return Manifest{}, err
		}
		manifest.Pages = append(manifest.Pages, name)
	}
	if err := writeManifest(directory, manifest); err != nil {
		_ = os.RemoveAll(directory)
		return Manifest{}, err
	}
	return manifest, nil
}

func (s *Store) ImportZip(path, fallbackTitle, source string) (Manifest, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("open presentation archive: %w", err)
	}
	defer reader.Close()
	id, err := newID()
	if err != nil {
		return Manifest{}, err
	}
	directory := filepath.Join(s.root, id)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{ID: id, Title: cleanTitle(fallbackTitle), Kind: "images", Source: source, CreatedAt: time.Now().UTC()}
	var totalWritten int64
	for _, entry := range reader.File {
		name := filepath.ToSlash(filepath.Clean(entry.Name))
		if entry.FileInfo().IsDir() || name == "manifest.json" {
			continue
		}
		if strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") || strings.Contains(name, "/") || !isImage(name) {
			_ = os.RemoveAll(directory)
			return Manifest{}, fmt.Errorf("invalid presentation entry %q", entry.Name)
		}
		if len(manifest.Pages) >= maxPages {
			_ = os.RemoveAll(directory)
			return Manifest{}, fmt.Errorf("presentation exceeds %d pages", maxPages)
		}
		sourceFile, openErr := entry.Open()
		if openErr != nil {
			_ = os.RemoveAll(directory)
			return Manifest{}, openErr
		}
		destination, createErr := os.OpenFile(filepath.Join(directory, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if createErr != nil {
			sourceFile.Close()
			_ = os.RemoveAll(directory)
			return Manifest{}, createErr
		}
		limited := &io.LimitedReader{R: sourceFile, N: (100 << 20) + 1}
		written, copyErr := io.Copy(destination, limited)
		closeErr := destination.Close()
		sourceFile.Close()
		if written > 100<<20 {
			copyErr = fmt.Errorf("presentation page %q exceeds 100 MB", name)
		}
		totalWritten += written
		if totalWritten > 2<<30 {
			copyErr = errors.New("expanded presentation exceeds 2 GB")
		}
		if copyErr != nil || closeErr != nil {
			_ = os.RemoveAll(directory)
			return Manifest{}, errors.Join(copyErr, closeErr)
		}
		manifest.Pages = append(manifest.Pages, name)
	}
	if len(manifest.Pages) == 0 {
		_ = os.RemoveAll(directory)
		return Manifest{}, errors.New("presentation archive contains no supported images")
	}
	sort.Slice(manifest.Pages, func(i, j int) bool { return naturalLess(manifest.Pages[i], manifest.Pages[j]) })
	if err := writeManifest(directory, manifest); err != nil {
		_ = os.RemoveAll(directory)
		return Manifest{}, err
	}
	return manifest, nil
}

func (s *Store) Get(id string) (Manifest, error) {
	if !validID(id) {
		return Manifest{}, errors.New("invalid presentation id")
	}
	data, err := os.ReadFile(filepath.Join(s.root, id, "manifest.json"))
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	if manifest.ID != id || len(manifest.Pages) == 0 {
		return Manifest{}, errors.New("invalid presentation manifest")
	}
	return manifest, nil
}

func (s *Store) PagePaths(manifest Manifest) ([]string, error) {
	paths := make([]string, 0, len(manifest.Pages))
	for _, name := range manifest.Pages {
		if filepath.Base(name) != name {
			return nil, errors.New("invalid presentation page path")
		}
		paths = append(paths, filepath.Join(s.root, manifest.ID, name))
	}
	return paths, nil
}

func (s *Store) TextPages(manifest Manifest) ([]string, error) {
	paths, err := s.PagePaths(manifest)
	if err != nil {
		return nil, err
	}
	pages := make([]string, 0, len(paths))
	for _, path := range paths {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		pages = append(pages, string(data))
	}
	return pages, nil
}

func (s *Store) Delete(id string) error {
	if !validID(id) {
		return errors.New("invalid presentation id")
	}
	return os.RemoveAll(filepath.Join(s.root, id))
}

func writeManifest(directory string, manifest Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "manifest.json"), append(data, '\n'), 0o644)
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Untitled presentation"
	}
	if len([]rune(title)) > 120 {
		return string([]rune(title)[:120])
	}
	return title
}

func newID() (string, error) {
	data := make([]byte, 8)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return time.Now().UTC().Format("20060102-150405-") + hex.EncodeToString(data), nil
}

func validID(id string) bool {
	return id != "" && filepath.Base(id) == id && !strings.Contains(id, "..")
}

func isImage(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func naturalLess(left, right string) bool {
	return strings.ToLower(left) < strings.ToLower(right)
}
