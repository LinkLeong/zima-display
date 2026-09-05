package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var version = "dev"

type configuration struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

type client struct {
	baseURL string
	token   string
	http    *http.Client
}

type manifest struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Kind  string   `json:"kind"`
	Pages []string `json:"pages"`
}

func main() {
	configPath := flag.String("config", os.Getenv("ZIMA_DISPLAY_CONFIG"), "path to connection config")
	baseURL := flag.String("url", os.Getenv("ZIMA_DISPLAY_URL"), "Zima Display base URL")
	token := flag.String("token", os.Getenv("ZIMA_DISPLAY_TOKEN"), "automation API token")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}
	if *configPath != "" {
		cfg, err := loadConfig(*configPath)
		fatalIf(err)
		if *baseURL == "" {
			*baseURL = cfg.BaseURL
		}
		if *token == "" {
			*token = cfg.Token
		}
	}
	if *baseURL == "" || *token == "" {
		fatalIf(errors.New("configure --url and --token, or pass --config"))
	}
	api := &client{baseURL: strings.TrimRight(*baseURL, "/") + "/api/v1", token: *token, http: &http.Client{Timeout: 10 * time.Minute}}
	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	var err error
	switch args[0] {
	case "status":
		err = api.status()
	case "present":
		if len(args) != 2 {
			err = errors.New("usage: zima-displayctl present <file-or-image-directory>")
		} else {
			err = api.present(args[1])
		}
	case "play":
		if len(args) != 2 {
			err = errors.New("usage: zima-displayctl play <file-or-url>")
		} else {
			err = api.play(args[1])
		}
	case "next", "previous", "stop":
		err = api.emptyPost("/presentation/" + args[0])
	case "mode":
		if len(args) != 2 {
			err = errors.New("usage: zima-displayctl mode <dashboard|canvas|clock|black|terminal>")
		} else {
			err = api.action(map[string]any{"action": "mode", "mode": args[1]})
		}
	default:
		err = fmt.Errorf("unknown command %q", args[0])
	}
	fatalIf(err)
}

func (c *client) present(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		files, err := imageFiles(path)
		if err != nil {
			return err
		}
		return c.presentImages(filepath.Base(filepath.Clean(path)), path, files)
	}
	extension := strings.ToLower(filepath.Ext(path))
	switch extension {
	case ".md", ".markdown", ".txt":
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return c.presentText(strings.TrimSuffix(filepath.Base(path), extension), path, paginateMarkdown(string(data)))
	case ".pptx":
		if pages, convertErr := convertOffice(path); convertErr == nil {
			return c.presentImages(strings.TrimSuffix(filepath.Base(path), extension), path, pages)
		}
		pages, err := extractPPTXText(path)
		if err != nil {
			return fmt.Errorf("convert PPTX: %w", err)
		}
		fmt.Fprintln(os.Stderr, "warning: LibreOffice/pdftoppm unavailable; presenting extracted slide text")
		return c.presentText(strings.TrimSuffix(filepath.Base(path), extension), path, pages)
	case ".pdf":
		pages, err := convertPDF(path)
		if err != nil {
			return err
		}
		return c.presentImages(strings.TrimSuffix(filepath.Base(path), extension), path, pages)
	case ".png", ".jpg", ".jpeg", ".webp", ".bmp":
		return c.presentImages(strings.TrimSuffix(filepath.Base(path), extension), path, []string{path})
	default:
		return fmt.Errorf("unsupported presentation format %q", extension)
	}
}

func (c *client) presentText(title, source string, pages []string) error {
	payload := map[string]any{"title": title, "source": source, "pages": pages}
	var created manifest
	if err := c.jsonRequest(http.MethodPost, "/presentations/text", payload, &created); err != nil {
		return err
	}
	if err := c.emptyPost("/presentations/" + url.PathEscape(created.ID) + "/activate"); err != nil {
		return err
	}
	fmt.Printf("Presenting %s (%d pages)\n", created.Title, len(created.Pages))
	return nil
}

func (c *client) presentImages(title, source string, pages []string) error {
	if root := generatedPageRoot(pages); root != "" {
		defer os.RemoveAll(root)
	}
	archivePath, err := createImageArchive(pages)
	if err != nil {
		return err
	}
	defer os.Remove(archivePath)
	var created manifest
	if err := c.multipartRequest("/presentations", "archive", archivePath, map[string]string{"title": title, "source": source}, &created); err != nil {
		return err
	}
	if err := c.emptyPost("/presentations/" + url.PathEscape(created.ID) + "/activate"); err != nil {
		return err
	}
	fmt.Printf("Presenting %s (%d pages)\n", created.Title, len(created.Pages))
	return nil
}

func (c *client) play(target string) error {
	parsed, _ := url.Parse(target)
	if parsed != nil && (parsed.Scheme == "http" || parsed.Scheme == "https" || parsed.Scheme == "rtsp") {
		return c.action(map[string]any{"action": "play", "path": target})
	}
	if _, err := os.Stat(target); err != nil {
		return err
	}
	var uploaded struct {
		Path string `json:"path"`
	}
	if err := c.multipartRequest("/upload", "file", target, nil, &uploaded); err != nil {
		return err
	}
	return c.action(map[string]any{"action": "play", "path": uploaded.Path})
}

func (c *client) status() error {
	var value any
	if err := c.jsonRequest(http.MethodGet, "/presentation/status", nil, &value); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(value, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *client) action(payload any) error {
	return c.jsonRequest(http.MethodPost, "/action", payload, nil)
}

func (c *client) emptyPost(path string) error {
	return c.jsonRequest(http.MethodPost, path, map[string]any{}, nil)
}

func (c *client) jsonRequest(method, path string, payload, destination any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	request, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	return c.do(request, destination)
}

func (c *client) multipartRequest(path, field, filename string, values map[string]string, destination any) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)
	go func() {
		part, writeErr := writer.CreateFormFile(field, filepath.Base(filename))
		if writeErr == nil {
			_, writeErr = io.Copy(part, file)
		}
		for key, value := range values {
			if writeErr == nil {
				writeErr = writer.WriteField(key, value)
			}
		}
		if closeErr := writer.Close(); writeErr == nil {
			writeErr = closeErr
		}
		_ = pipeWriter.CloseWithError(writeErr)
	}()
	request, err := http.NewRequest(http.MethodPost, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return c.do(request, destination)
}

func generatedPageRoot(pages []string) string {
	if len(pages) == 0 {
		return ""
	}
	directory := filepath.Dir(pages[0])
	base := filepath.Base(directory)
	if strings.HasPrefix(base, "zima-display-office-") || strings.HasPrefix(base, "zima-display-pdf-") {
		return directory
	}
	return ""
}

func (c *client) do(request *http.Request, destination any) error {
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var apiError struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(response.Body).Decode(&apiError)
		if apiError.Error == "" {
			apiError.Error = response.Status
		}
		return errors.New(apiError.Error)
	}
	if destination != nil {
		return json.NewDecoder(response.Body).Decode(destination)
	}
	return nil
}

func loadConfig(path string) (configuration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return configuration{}, err
	}
	var cfg configuration
	if err := json.Unmarshal(data, &cfg); err != nil {
		return configuration{}, err
	}
	return cfg, nil
}

func paginateMarkdown(content string) []string {
	var pages []string
	var current []string
	flush := func() {
		if len(current) == 0 {
			return
		}
		pages = append(pages, strings.Join(current, "\n"))
		current = nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		newSection := strings.HasPrefix(trimmed, "# ") && len(current) > 0
		newSection = newSection || (strings.HasPrefix(trimmed, "## ") && len(current) >= 4)
		if trimmed == "---" || newSection {
			flush()
			if trimmed == "---" {
				continue
			}
		}
		if trimmed != "" {
			current = append(current, trimmed)
		}
		if len(current) >= 16 {
			flush()
		}
	}
	flush()
	if len(pages) == 0 {
		return []string{"Empty document"}
	}
	return pages
}

func extractPPTXText(path string) ([]string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	type slide struct {
		number int
		file   *zip.File
	}
	var slides []slide
	pattern := regexp.MustCompile(`^ppt/slides/slide([0-9]+)\.xml$`)
	for _, file := range reader.File {
		match := pattern.FindStringSubmatch(file.Name)
		if len(match) != 2 {
			continue
		}
		number, _ := strconv.Atoi(match[1])
		slides = append(slides, slide{number: number, file: file})
	}
	sort.Slice(slides, func(i, j int) bool { return slides[i].number < slides[j].number })
	var pages []string
	for _, item := range slides {
		stream, err := item.file.Open()
		if err != nil {
			return nil, err
		}
		decoder := xml.NewDecoder(stream)
		var texts []string
		for {
			token, decodeErr := decoder.Token()
			if decodeErr == io.EOF {
				break
			}
			if decodeErr != nil {
				stream.Close()
				return nil, decodeErr
			}
			start, ok := token.(xml.StartElement)
			if !ok || start.Name.Local != "t" {
				continue
			}
			var text string
			if decoder.DecodeElement(&text, &start) == nil && strings.TrimSpace(text) != "" {
				texts = append(texts, strings.TrimSpace(text))
			}
		}
		stream.Close()
		pages = append(pages, strings.Join(texts, "\n"))
	}
	if len(pages) == 0 {
		return nil, errors.New("PPTX contains no readable slides")
	}
	return pages, nil
}

func convertOffice(path string) ([]string, error) {
	office, err := firstCommand("libreoffice", "soffice")
	if err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, err
	}
	directory, err := os.MkdirTemp("", "zima-display-office-*")
	if err != nil {
		return nil, err
	}
	// The generated images must survive until the upload completes.
	if err := exec.Command(office, "--headless", "--convert-to", "pdf", "--outdir", directory, path).Run(); err != nil {
		os.RemoveAll(directory)
		return nil, err
	}
	pdf := filepath.Join(directory, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))+".pdf")
	return convertPDFInto(pdf, directory)
}

func convertPDF(path string) ([]string, error) {
	directory, err := os.MkdirTemp("", "zima-display-pdf-*")
	if err != nil {
		return nil, err
	}
	return convertPDFInto(path, directory)
}

func convertPDFInto(path, directory string) ([]string, error) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		os.RemoveAll(directory)
		return nil, errors.New("PDF presentation requires pdftoppm")
	}
	prefix := filepath.Join(directory, "page")
	if output, err := exec.Command("pdftoppm", "-png", "-r", "144", path, prefix).CombinedOutput(); err != nil {
		os.RemoveAll(directory)
		return nil, fmt.Errorf("convert PDF: %v: %s", err, strings.TrimSpace(string(output)))
	}
	files, err := filepath.Glob(prefix + "-*.png")
	if err != nil || len(files) == 0 {
		os.RemoveAll(directory)
		return nil, errors.New("PDF conversion produced no pages")
	}
	sort.Slice(files, func(i, j int) bool { return pageFileLess(files[i], files[j]) })
	return files, nil
}

func imageFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !isImage(entry.Name()) {
			continue
		}
		files = append(files, filepath.Join(directory, entry.Name()))
	}
	sort.Slice(files, func(i, j int) bool { return pageFileLess(files[i], files[j]) })
	if len(files) == 0 {
		return nil, errors.New("image directory contains no supported pages")
	}
	return files, nil
}

func createImageArchive(files []string) (string, error) {
	tmp, err := os.CreateTemp("", "zima-display-pages-*.zip")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	writer := zip.NewWriter(tmp)
	for index, filename := range files {
		input, err := os.Open(filename)
		if err != nil {
			writer.Close()
			tmp.Close()
			os.Remove(path)
			return "", err
		}
		name := fmt.Sprintf("%03d%s", index+1, strings.ToLower(filepath.Ext(filename)))
		entry, err := writer.Create(name)
		if err == nil {
			_, err = io.Copy(entry, input)
		}
		input.Close()
		if err != nil {
			writer.Close()
			tmp.Close()
			os.Remove(path)
			return "", err
		}
	}
	if err := writer.Close(); err != nil {
		tmp.Close()
		os.Remove(path)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

func isImage(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func pageFileLess(left, right string) bool {
	pattern := regexp.MustCompile(`([0-9]+)(?:\.[^.]+)?$`)
	leftMatch := pattern.FindStringSubmatch(filepath.Base(left))
	rightMatch := pattern.FindStringSubmatch(filepath.Base(right))
	if len(leftMatch) == 2 && len(rightMatch) == 2 {
		leftNumber, _ := strconv.Atoi(leftMatch[1])
		rightNumber, _ := strconv.Atoi(rightMatch[1])
		if leftNumber != rightNumber {
			return leftNumber < rightNumber
		}
	}
	return strings.ToLower(filepath.Base(left)) < strings.ToLower(filepath.Base(right))
}

func firstCommand(names ...string) (string, error) {
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", errors.New("LibreOffice is not installed")
}

func fatalIf(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: zima-displayctl [--config FILE] <status|present|play|next|previous|stop|mode>")
}
