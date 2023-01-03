// Copyright (c) 2022, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// copied and heavily edited from go src/cmd/gofmt/gofmt.go:

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/goki/gosl/slprint"
)

// flags
var (
	outDir  = flag.String("out", "shaders", "output directory for shader code, relative to where gosl is invoked")
	keepTmp = flag.Bool("keep", false, "keep temporary converted versions of the source files, for debugging")
)

// Keep these in sync with go/format/format.go.
const (
	tabWidth    = 8
	printerMode = slprint.UseSpaces | slprint.TabIndent | printerNormalizeNumbers

	// printerNormalizeNumbers means to canonicalize number literal prefixes
	// and exponents while printing. See https://golang.org/doc/go1.13#gosl.
	//
	// This value is defined in go/printer specifically for go/format and cmd/gosl.
	printerNormalizeNumbers = 1 << 30
)

var (
	inFiles    []string            // list of all output files saved
	filesProcd = map[string]bool{} // prevent redundancies
	slFiles    = map[string][][]byte{}
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gosl [flags] [path ...]\n")
	flag.PrintDefaults()
}

func isGoFile(f fs.DirEntry) bool {
	// ignore non-Go files
	name := f.Name()
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !f.IsDir()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	goslMain()
}

func addFile(fn string) bool {
	if _, has := filesProcd[fn]; has {
		return false
	}
	inFiles = append(inFiles, fn)
	filesProcd[fn] = true
	return true
}

func goslMain() {
	if *outDir != "" {
		os.MkdirAll(*outDir, 0755)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Printf("at least one file name must be passed\n")
		return
	}

	for _, arg := range args {
		switch info, err := os.Stat(arg); {
		case err != nil:
			fmt.Println(err)
		case !info.IsDir():
			// Non-directory arguments are always formatted.
			arg := arg
			addFile(arg)
		default:
			// Directories are walked, ignoring non-Go files.
			err := filepath.WalkDir(arg, func(path string, f fs.DirEntry, err error) error {
				if err != nil || !isGoFile(f) {
					return err
				}
				_, err = f.Info()
				if err != nil {
					return nil
				}
				addFile(path)
				return nil
			})
			if err != nil {
				log.Println(err)
			}
		}
	}

	processFiles(inFiles)
}
