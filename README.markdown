This is a fork of popular command-line fuzzy finder
[fzf](https://github.com/junegunn/fzf) which implements so-called
*abbreviation* (or *acronym*) fuzzy matching.

Unlike "normal" fuzzy matching, this method allows pattern letters to match not
at any position of searched text but only at word beginnings. In other words
search query must constitute a valid abbreviation (or acronym) for the some
words from searched text.

For example, `foo bar` will be matched by `fb` or `fob` but not by `fr` or
`far`.

Please see README of my other project
[vim-abbrev-matcher](https://github.com/vim-scripts/vim-abbrev-matcher) for
expanded reasoning and benefits of such approach. In short, this type of search
filters candidates much better and faster than normal fuzzy search which is
beneficial in some use cases, such as:
- When searching command-line histories you might prefer to turn sorting off in
  order to see filtered candidates in original (chronological) order. In this
  case abbreviation matching narrows candidates much faster.
- Fuzzy searchers often provide some sorting heuristic which scores each match
  and tries to give you "best" matches first. This heuristic may fail for your
  specific use case and give you a large number of unrelated results.
  Abbreviation matching also uses such heuristic but it also produces much less
  results so there is more chance the needed element will be contained in
  displayed results even when having a bad score.

# Installation

```
go get github.com/sergei-dyshel/fzf-abbrev
```

# Usage

`fzf-abbrev` may be used as drop-in replacement for `fzf`. In that case just
copy compiled binary over `~/.fzf/bin/fzf`). In order to get "normal" fuzzy
matching, start your query with `#`.

There is one new command-line option `--abbrev` which is comma-separated list
of options that control new matcher's behavior:
- `no-default` brings back normal fuzzy matching as default behavior, so
  `fzf-abbrev --abbrev=no-default` works the same way as original `fzf`.
- `fast` makes matching faster but resulted scoring may not be optimal. Use
  this option when candidate sorting is off (i.e. for filtering histories).
- `file-paths` optimizes scoring heuristic to match file paths. In this case it
  will prefer matching file names over matching file directories.

# Scoring

Basically similar to
[vim-abbrev-matcher](https://github.com/vim-scripts/vim-abbrev-matcher#ranking).
The weights are slightly different though.

# Development status

This is work in progress but current version is already pretty stable and
usable.