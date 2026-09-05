package main

import "testing"

func TestPaginateMarkdownSplitsHeadingsAndLongPages(t *testing.T) {
	content := "# First\nA\nB\nC\nD\n## Second\nE\nF"
	pages := paginateMarkdown(content)
	if len(pages) != 2 {
		t.Fatalf("got %d pages, want 2: %#v", len(pages), pages)
	}
	if pages[0] != "# First\nA\nB\nC\nD" || pages[1] != "## Second\nE\nF" {
		t.Fatalf("unexpected pages %#v", pages)
	}
}
