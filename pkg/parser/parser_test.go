package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownParser(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.md")
	content := "# Hello World\nThis is a test markdown file."

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	p := &MarkdownParser{}
	parsed, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse markdown: %v", err)
	}

	if parsed != content {
		t.Fatalf("Expected %q, got %q", content, parsed)
	}
}

func TestUnimplementedParsers(t *testing.T) {
	pdf := &PDFParser{}
	_, err := pdf.Parse("dummy.pdf")
	if err != ErrNotImplemented {
		t.Fatalf("Expected ErrNotImplemented for PDF, got %v", err)
	}

	docx := &DocxParser{}
	_, err = docx.Parse("dummy.docx")
	if err != ErrNotImplemented {
		t.Fatalf("Expected ErrNotImplemented for DOCX, got %v", err)
	}
}
