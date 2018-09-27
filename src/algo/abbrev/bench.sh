#!/bin/bash

set -x
set -e

exe=abbrev.test
prof=abbrev.prof
pdf=abbrev.pdf

go test github.com/sergei-dyshel/fzf-abbrev/src/algo/abbrev -bench 'BenchmarkVeryLongLine' \
         "$@" -cpuprofile=$prof

go tool pprof --pdf -compact_labels $exe $prof > $pdf
xdg-open 1>/dev/null 2>&1 $pdf &