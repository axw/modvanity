## metaimport

Metaimport generates HTML files containing `<meta name="go-import">`
tags for remote Git repositories. 

These tags are used by commands such as `go get` to determine how to fetch 
source code. See `go help importpath` for details.

## Example

```
$ metaimport -o html example.org/myrepo github.com/user/myrepo
```

Once the HTML files are generated, you can serve them at the root of your domain 
(`example.org` in this example) with something like:

```
$ cd html/example.org
$ python -m SimpleHTTPServer 443
$ go get example.org/myrepo # should now work
```

## Usage

```
usage: metaimport [-godoc] [-o dir] [-branch branch] <import> <repo>

metaimport generates HTML files with <meta name="go-import"> tags as expected
by go get. 'repo' specifies the Git repository containing Go source code to
generate meta tags for. 'import' specifies the import path of the root of
the repository.

The program automatically handles generating HTML files for subpackages in the
repository.

Flags
   -branch   Branch to use (default: repository's default branch).
   -godoc    Include <meta name="go-source"> tag as expected by godoc.org.
             Only partial support for repositories not hosted on github.com.
   -o        Directory to write generated HTML (default: html).
             It creates the directory with 0744 permissions if it doesn't exist.
```
