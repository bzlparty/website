package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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

var (
	assets       []string
	excludePath  string
	files        FilePair
	packageName  string
	templateFile FileContent
	config       PackageConfig
)

type Link struct {
	Name string
	Href string
}

type TemplateData struct {
	Title    string
	Heading  string
	Content  string
	NavLinks []Link
	Package  string
	Css      []string
}

type FileConfig struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Target     string `json:"target"`
}
type PackageConfig struct {
	Name     string       `json:"name"`
	Files    []FileConfig `json:"files"`
	Template string       `json:"template"`
}

type RenderedFile struct {
	Content    string
	Heading    string
	Target     string
	TargetPath string
}

func renderFiles(files []FileConfig) (result []RenderedFile) {
	for _, file := range config.Files {
		content, err := ioutil.ReadFile(file.SourcePath)

		if err != nil {
			log.Fatal(fmt.Sprintf("Error processing file %s", file.SourcePath), err)
		}

		doc := parseMarkdown(content)
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

func generatePackageLinks(files []RenderedFile, _package string) (links []Link) {
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

func main() {
	parseFlags()

	var renderedFiles = renderFiles(config.Files)
	var links = generatePackageLinks(renderedFiles, config.Name)

	templateFile, err := ioutil.ReadFile(config.Template)

	// sort.Slice(nav, func(i, j int) bool {
	// 	return nav[i].Name < nav[j].Name
	// })

	t, err := template.New("page").Parse(string(templateFile))

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range renderedFiles {
		data := TemplateData{
			Title:    getTitle(&file, config.Name),
			Package:  config.Name,
			NavLinks: links,
			Content:  file.Content,
		}

		output, err := os.Create(file.TargetPath)
		if err != nil {
			os.Exit(128)
		}

		if err = t.Execute(output, data); err != nil {
			log.Fatal(fmt.Sprintf("Could not write content to file %s", file.TargetPath))
			os.Exit(128)
		}
	}
}

func parseFlags() {
	flag.Func("jsonConfig", "JSON Config string", func(value string) error {
		json.Unmarshal([]byte(value), &config)
		return nil
	})
	flag.Parse()
}

type FilePair map[string]string

func (v *FilePair) String() string {
	var r []string

	for source, target := range *v {
		r = append(r, fmt.Sprintf("%s:%s", source, target))
	}

	return strings.Join(r, ",")
}

func (v *FilePair) Set(s string) error {
	r := make(map[string]string)

	for _, f := range strings.Split(s, ",") {
		source, target, err := splitFiles(f)

		if err != nil {
			return err
		}

		r[source] = target
	}

	if len(r) == 0 {
		return errors.New("No files given")
	}

	*v = r

	return nil
}

type FileContent []byte

func (v *FileContent) Set(path string) (err error) {
	*v, err = ioutil.ReadFile(path)

	return
}

func (v *FileContent) String() string {
	return string(*v)
}

func splitFiles(s string) (string, string, error) {
	x := strings.Split(s, ":")

	if len(x) < 2 {
		return "", "", errors.New("File pair uncomplete")
	}

	return x[0], x[1], nil
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
