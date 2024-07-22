# ezf
Find files with the command line

### Why create ezf?

There are great tools available for finding files. I regularly find myself using a combination of `find` and `grep` to search for text within files matching a given extension:

```sh
time find ~/dev/github.com/codingconcepts \
-type f \
-name "*.md" \
-exec grep -l "demo-locality" {} + 2>/dev/null

0.32s user 1.80s system 53% cpu 3.977 total
0.31s user 1.48s system 87% cpu 2.056 total
0.31s user 1.49s system 92% cpu 1.959 total
```

I've also never committed that command to memory and so need to keep a repository of frequently used commands (which slows me down when I want to find text within a file).

With ezf, that same command looks like this:

```sh
time ezf \
-d ~/dev/github.com/codingconcepts \
-s demo-locality \
-n "*.md"

0.25s user 0.81s system 143% cpu 0.737 total
0.25s user 0.81s system 143% cpu 0.741 total
0.25s user 0.82s system 142% cpu 0.754 total
```

Not only is ezf easier for me to remember, it's also ~3x faster!

Win win!

### Installation

Head over to 

### Usage

Generate usage text

```sh
Find files easily from the command line.

Usage:
  ezf [flags]
  ezf [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Show the version of ezf.

Flags:
  -c, --concurrency int   maximum concurrency to use for file searching (default 4)
  -d, --dir string        directory to search (default ".")
  -h, --help              help for ezf
  -n, --name string       file name pattern to match
  -s, --search string     string to search for

Use "ezf [command] --help" for more information about a command.
```

Display version of ezf

```sh
ezf version
v0.0.1
```

Find text within a file

```sh
ezf -d github.com/codingconcepts/ezf -s search -n "*.*"

github.com/codingconcepts/ezf/README.md
github.com/codingconcepts/ezf/ezf.go
```