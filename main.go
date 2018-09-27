package main

import (
	"github.com/sergei-dyshel/fzf-abbrev/src"
	"github.com/sergei-dyshel/fzf-abbrev/src/protector"
)

var version string = "0.27"
var revision string = "devel"

func main() {
	protector.Protect()
	fzf.Run(fzf.ParseOptions(), version, revision)
}
