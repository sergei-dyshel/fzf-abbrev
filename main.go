package main

import (
	fzf "github.com/sergei-dyshel/fzf-abbrev/src"
	"github.com/sergei-dyshel/fzf-abbrev/src/protector"
)

var version string = "0.30"
var revision string = "devel"

func main() {
	protector.Protect()
	fzf.Run(fzf.ParseOptions(), version, revision)
}
