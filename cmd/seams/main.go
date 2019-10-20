package main

import (
	"github.com/ichiban/seams"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(seams.Analyzer)
}
