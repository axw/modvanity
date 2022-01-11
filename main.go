// Command modvanity generates HTML files containing <meta name="go-import">
// tags for Go modules in a GitHub repository.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"golang.org/x/mod/modfile"
)

const help = `usage: modvanity [-branch branch] [-o dir] [-redirect] <import-prefix> <repo>

modvanity generates HTML files with <meta name="go-import"> tags as expected
by go get. 'repo' specifies the GitHub repository containing Go modules.
'import-prefix' is the import path corresponding to the repository root.

Flags
   -branch    Branch to use (default: remote's default branch).
   -o         Output directory for generated HTML files (default: html).
              The directory is created with 0755 permissions if it doesn't exist.
   -redirect  Redirect to pkg.go.dev documentation when visited in a browser (default: true).
   -v         Log verbosely (default: false).

Examples
   modvanity go.elastic.co/apm https://github.com/elastic/apm-agent-go
`

func usage() {
	fmt.Fprintf(os.Stderr, help)
	os.Exit(2)
}

const (
	permDir  = 0755
	permFile = 0644
)

var (
	branch        = flag.String("branch", "", "")
	outputDir     = flag.String("o", "html", "")
	verbose       = flag.Bool("v", false, "")
	godocRedirect = flag.Bool("redirect", true, "")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("modvanity: ")

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		usage()
	}

	importPrefix := args[0]
	repoURL := args[1]
	htmlTmpl := template.Must(template.New("").Parse(tmpl))

	var referenceName plumbing.ReferenceName
	if *branch != "" {
		referenceName = plumbing.NewBranchReferenceName(*branch)
	}

	fs := memfs.New()
	storer := memory.NewStorage()
	repo, err := git.Clone(storer, fs, &git.CloneOptions{
		URL:           repoURL,
		ReferenceName: referenceName,
		SingleBranch:  referenceName != "",
		Depth:         1,
	})
	if err != nil {
		log.Fatalf("making repository: %s", err)
	}

	// Get the tree for the HEAD of the branch.
	head, err := repo.Head()
	if err != nil {
		log.Fatalf("getting HEAD: %s", err)
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		log.Fatalf("getting HEAD commit: %s", err)
	}
	tree, err := commit.Tree()
	if err != nil {
		log.Fatalf("getting HEAD commit tree: %s", err)
	}

	// Find Go modules defined within the repo.
	if *verbose {
		log.Printf("finding modules...")
	}
	modules, err := findModules(tree)
	if err != nil {
		log.Fatalf("finding go modules: %s", err)
	}

	// Write an index.html for each module found.
	for _, module := range modules {
		if !strings.HasPrefix(module, importPrefix) {
			if *verbose {
				log.Printf("ignoring module %q, does not match prefix %q", module, importPrefix)
			}
			continue
		}

		dir := filepath.Join(*outputDir, filepath.FromSlash(module))
		if err := os.MkdirAll(dir, permDir); err != nil {
			log.Fatalf("making directory %s: %s", dir, err)
		}
		indexPath := filepath.Join(dir, "index.html")
		if *verbose {
			log.Printf("writing file %s", indexPath)
		}

		f, err := os.Create(indexPath)
		if err != nil {
			log.Fatalf("creating file: %s", err)
		}
		args := TemplateArgs{
			GoImport: GoImport{
				ImportPrefix: importPrefix,
				VCS:          "git",
				RepoRoot:     repoURL,
			},
			GodocURL:      fmt.Sprintf("https://pkg.go.dev/%s", module),
			GodocRedirect: *godocRedirect,
		}
		if err := htmlTmpl.Execute(f, args); err != nil {
			log.Fatalf("executing template: %s", err)
		}

		if err := f.Close(); err != nil {
			log.Fatalf("closing file: %s", err)
		}
	}
}

const tmpl = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		{{ with .GoImport }}<meta name="go-import" content="{{ .ImportPrefix }} {{ .VCS }} {{ .RepoRoot }}">{{ end }}
		{{ if .GodocRedirect }}<meta http-equiv="refresh" content="0; url='{{ .GodocURL }}'">{{ end }}
	</head>
	<body>
		{{ if .GodocRedirect -}}
		Redirecting to <a href="{{ .GodocURL }}">{{ .GodocURL }}</a>
		{{- else -}}
		Repository: <a href="{{ .GoImport.RepoRoot }}">{{ .GoImport.RepoRoot }}</a>
		<br>
		Godoc: <a href="{{ .GodocURL }}">{{ .GodocURL }}</a>
		{{- end }}
	</body>
</html>
`

type TemplateArgs struct {
	GoImport      GoImport
	GodocRedirect bool
	GodocURL      string
}

type GoImport struct {
	ImportPrefix, VCS, RepoRoot string
}

func findModules(tree *object.Tree) ([]string, error) {
	var modules []string
	if err := tree.Files().ForEach(func(f *object.File) error {
		if strings.HasPrefix(f.Name, ".") || strings.HasPrefix(f.Name, "_") || filepath.Base(f.Name) != "go.mod" {
			return nil
		}

		// Ignore modules inside testdata and internal directories.
		dir, _ := filepath.Split(f.Name)
		for {
			var file string
			dir = strings.TrimSuffix(dir, "/")
			dir, file = filepath.Split(dir)
			if dir == "" {
				break
			}
			if file == "internal" || file == "testdata" {
				return nil
			}
		}

		if *verbose {
			log.Printf("parsing %q", f.Name)
		}
		content, err := f.Contents()
		if err != nil {
			return err
		}
		m, err := modfile.ParseLax(f.Name, []byte(content), nil)
		if err != nil {
			return err
		}
		modules = append(modules, m.Module.Mod.Path)
		return nil
	}); err != nil {
		return nil, err
	}
	return modules, nil
}
