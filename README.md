## modvanity

`go get github.com/axw/modvanity`

Specify a Git repository and `modvanity` will generate a directory of
HTML files containing `<meta name="go-import">` tags for the Go modules
in the repository, suited for your vanity URL.

These tags are used by commands such as `go get` to determine how to fetch 
source code. See `go help importpath` for details.

## Example

```
$ modvanity -o html example.org/myrepo https://github.com/user/myrepo
```

will generate HTML files like:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="go-import" content="example.org/myrepo git https://github.com/user/myrepo">
    <meta http-equiv="refresh" content="0; url='https://godoc.org/example.org/myrepo'">
  </head>
  <body>
    Redirecting to <a href="https://pkg.go.dev/example.org/myrepo">https://pkg.go.dev/example.org/myrepo</a>
  </body>
</html>
```

Once the HTML files are generated, you can serve them at the root of your domain 
(`example.org` in this example) with something like:

```
$ cd html/example.org
$ file-server . # serve files in the directory over https
$ go get example.org/myrepo # should now work
```

## Usage

See `modvanity -h`.

```
usage: modvanity [-branch branch] [-o dir] [-redirect] <import-prefix> <repo>

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
```
