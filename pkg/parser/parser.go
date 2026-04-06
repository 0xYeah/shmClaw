package parser

import (
	"errors"
	"io"
	"os"
)

var (
	ErrNotImplemented = errors.New("parser not implemented yet")
)

// Parser defines a unified interface for extracting text from multimodal documents.
type Parser interface {
	Parse(filePath string) (string, error)
}

// MarkdownParser implements the Parser interface for markdown files.
type MarkdownParser struct{}

func (m *MarkdownParser) Parse(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// PDFParser is a placeholder for future PDF parsing implementation.
type PDFParser struct{}

func (p *PDFParser) Parse(filePath string) (string, error) {
	return "", ErrNotImplemented
}

// DocxParser is a placeholder for future DOCX parsing implementation.
type DocxParser struct{}

func (d *DocxParser) Parse(filePath string) (string, error) {
	return "", ErrNotImplemented
}
