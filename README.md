# ezf
Find files with the command line

### Why create ezf?

There are great tools available for finding files. I regularly find myself using a combination of `find` and `grep` to search for text within files matching a given extension:

```sh
time find ~/dev/github.com/codingconcepts \
-type f \
-name "*.md" \
-exec grep -l "demo-locality" {} + 2>/dev/null

0.31s user 1.48s system 87% cpu 2.056 total
```

I've also never committed that command to memory and so need to keep a repository of frequently used commands (which slows me down when I want to find text within a file).

With ezf, that same command looks like this:

```sh
time ezf \
--dir ~/dev/github.com/codingconcepts \
--search demo-locality \
--name "*.md"

0.24s user 0.84s system 142% cpu 0.763 total
```

Not only is ezf easier for me to remember, it's also ~2.5x faster!

Win win!

### Installation

Head over to 