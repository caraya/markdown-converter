package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
	"go.abhg.dev/goldmark/toc"
	"gopkg.in/yaml.v2"
)

type FrontMatterData struct {
	Title string   `yaml:"title"`
	Date  string   `yaml:"date"`
	Tags  []string `yaml:"tags"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: <program> <destination-directory> <markdown-file-or-directory>...")
		os.Exit(1)
	}

	destDir := os.Args[1]
	md := createMarkdownConverter()

	for _, arg := range os.Args[2:] {
		fi, err := os.Stat(arg)
		if err != nil {
			fmt.Printf("Error accessing %s: %v\n", arg, err)
			continue
		}

		if fi.IsDir() {
			processDirectory(arg, destDir, md)
		} else {
			processFile(arg, destDir, md)
		}
	}
}

func createMarkdownConverter() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
			&toc.Extender{
				TitleDepth: 2,
			},
			&mermaid.Extender{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

func processFile(sourceFileName, destDir string, md goldmark.Markdown) {
	fileContent, err := os.ReadFile(sourceFileName)
	if err != nil {
		fmt.Printf("Cannot read file %s: %v\n", sourceFileName, err)
		return
	}

	if !utf8.Valid(fileContent) {
		fmt.Printf("File %s is not a valid UTF-8 text file\n", sourceFileName)
		return
	}

	frontmatter, content, err := extractFrontmatterAndContent(fileContent)
	if err != nil {
		fmt.Printf("Failed to extract frontmatter: %v", err)
	}

	var data FrontMatterData
	if err := yaml.Unmarshal([]byte(frontmatter), &data); err != nil {
		fmt.Printf("Failed to unmarshal frontmatter: %v", err)
	}

	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		fmt.Printf("Failed to convert Markdown content: %v\n", err)
		return
	}

	writeHTMLToFile(buf, sourceFileName, destDir, data)
}

func processDirectory(dirPath, destDir string, md goldmark.Markdown) {
	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(d.Name(), ".md") || strings.HasSuffix(d.Name(), ".markdown")) {
			processFile(path, destDir, md)
		}
		return nil
	})
}

func extractFrontmatterAndContent(fileContent []byte) (frontmatter, content string, err error) {
	const delimiter = "---"
	contentStr := string(fileContent)

	if !strings.HasPrefix(contentStr, delimiter) {
		return "", contentStr, nil
	}

	parts := strings.SplitN(contentStr, delimiter, 3)

	if len(parts) < 3 {
		return "", "", fmt.Errorf("malformed frontmatter: missing end delimiter")
	}

	return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), nil
}

func writeHTMLToFile(buf bytes.Buffer, sourceFileName, destDir string, data interface{}) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Failed to create the destination directory %s: %v\n", destDir, err)
		return
	}

	filename := filepath.Base(sourceFileName)
	targetLocation := filepath.Join(destDir, strings.TrimSuffix(filename, filepath.Ext(filename))+".html")

	tmpl, err := template.ParseFiles("templates/template.html")
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		return
	}

	var wrappedContent bytes.Buffer
	if err := tmpl.Execute(&wrappedContent, map[string]interface{}{
		"Content":  template.HTML(buf.String()),
		"Metadata": data, // Pass the frontmatter data here
	}); err != nil {
		fmt.Printf("Failed to execute template: %v\n", err)
		return
	}

	err = os.WriteFile(targetLocation, wrappedContent.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Failed to write to file %s: %v\n", targetLocation, err)
		return
	}

	fmt.Printf("File successfully written to %s\n", targetLocation)
}
