package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Link struct {
	Name string
	Href string
}

type TemplateData struct {
	Title    string
	Content  string
	NavLinks []Link
	Package  string
}

type FileInfo struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Target     string `json:"target"`
}

type JsonConfig struct {
	Name     string     `json:"name"`
	Files    []FileInfo `json:"files"`
	Template string     `json:"template"`
}

type RenderedFile struct {
	Content    string
	Heading    string
	Target     string
	TargetPath string
}

func renderFiles(c JsonConfig) (result []RenderedFile, err error) {
	var content []byte
	for _, file := range c.Files {
		content, err = ioutil.ReadFile(file.SourcePath)

		if err != nil {
			return
		}

		doc := parseMarkdown(content)
		rewriteLocalLinks(doc, c.Name)
		html := renderHTML(doc)
		heading := getFirstHeading(doc)
		result = append(result, RenderedFile{
			Content:    string(html),
			Heading:    heading,
			Target:     file.Target,
			TargetPath: file.TargetPath,
		})
	}

	return
}

func generatePackageLinks(files []RenderedFile) (links []Link) {
	for _, file := range files {
		if filepath.Base(file.Target) == "index.html" {
			continue
		}
		links = append(links, Link{Name: file.Heading, Href: "/" + file.Target})
	}

	return
}

func getTitle(file *RenderedFile, packageName string) string {
	var title = "| bzlparty"
	var heading = file.Heading
	if packageName != "" {
		if filepath.Base(file.Target) == "index.html" {
			heading = "README"
		}

		return fmt.Sprintf("%s - %s %s", heading, packageName, title)
	}

	return fmt.Sprintf("%s %s", heading, title)
}

func processFiles(c JsonConfig) (err error) {
	renderedFiles, err := renderFiles(c)
	if err != nil {
		return
	}
	links := generatePackageLinks(renderedFiles)
	templateFile, err := ioutil.ReadFile(c.Template)

	if err != nil {
		return
	}

	t, err := template.New("page").Parse(string(templateFile))

	if err != nil {
		return
	}

	for _, file := range renderedFiles {
		data := TemplateData{
			Title:    getTitle(&file, c.Name),
			Package:  c.Name,
			NavLinks: links,
			Content:  file.Content,
		}

		var output io.Writer
		output, err = os.Create(file.TargetPath)

		if err != nil {
			return
		}

		if err = t.Execute(output, data); err != nil {
			return
		}
	}

	return
}

func main() {
	var config JsonConfig
	parseConfig(&config)

	if err := processFiles(config); err != nil {
		log.Fatal(fmt.Sprintf("There was an Error, %s", err.Error()))
		os.Exit(127)
	}
}

func parseConfig(c *JsonConfig) {
	flag.Func("jsonConfig", "JSON Config string", func(value string) error {
		json.Unmarshal([]byte(value), c)
		return nil
	})
	flag.Parse()
}

func rewriteLocalLinks(doc ast.Node, packageName string) {
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if l, ok := node.(*ast.Link); ok && entering {
			href := string(l.Destination)
			if strings.HasPrefix(href, "/") {
				link := strings.Replace(href, ".md", ".html", 1)
				l.Destination = []byte(fmt.Sprintf("/%s%s", packageName, link))
				return ast.Terminate
			}
		}

		return ast.GoToNext
	})
}

func getFirstHeading(doc ast.Node) (heading string) {
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if h, ok := node.(*ast.Heading); ok && entering {
			if h.Level == 1 {
				heading = string(h.Children[0].AsLeaf().Literal)
				return ast.Terminate
			}
		}

		return ast.GoToNext
	})

	return
}

func parseMarkdown(md []byte) ast.Node {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	return p.Parse(md)
}

func renderHTML(doc ast.Node) []byte {
	htmlFlags := html.CommonFlags
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
