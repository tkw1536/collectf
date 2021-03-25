// Command collectf collects files passed via standard input into the destination provided by the first command line argument.
// It ensures that no two files are named identically by appending appropriate suffixes.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	mp := RenameMap(make(map[string]int))

	for r := range readInput() {
		fn := filepath.Join(dest, mp.Get(r))

		// echo out what is done!
		if !move {
			fmt.Printf("cp %s %s\n", r, fn)
			if !simulate {
				must(CopyFile(fn, r))
			}
		} else {
			fmt.Printf("mv %s %s\n", r, fn)
			if !simulate {
				must(MoveFile(fn, r))
			}
		}
	}
}

// read input files from stdin
func readInput() <-chan string {
	c := make(chan string)
	go func() {
		defer close(c)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			c <- scanner.Text()
		}
	}()

	return c
}

// err calls panIc(err) unless err is nil
func must(err error) {
	if err == nil {
		return
	}
	panic(err)
}

// RenameMap provides Get
type RenameMap map[string]int

// Get returns a depuplicated filename
func (r RenameMap) Get(name string) string {
	_, fn := filepath.Split(name)

	// did we encounter the file name before?
	count, has := r[fn]
	if !has {
		r[fn] = 0
		return fn
	}

	// store the new count!
	count++
	r[fn] = count

	// build the new filename by adding a '_'
	parts := strings.SplitN(fn, ".", 2)
	parts[0] += fmt.Sprintf("_%d", count)
	fn = strings.Join(parts, ".")

	// check that we haven't encountered this one yet!
	return r.Get(fn)
}

// CopyFiles copies the file at src to the file at dst
// preserving file modes.
func CopyFile(dst, src string) error {
	// open inFile
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// open outFile
	out, err := os.Create(dst)
	defer out.Close()

	// copy and ensure that we have synced the file to disk!
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	// stat the source file!
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	// stat the dst file
	err = os.Chmod(dst, stat.Mode())
	if err != nil {
		return err
	}

	// and done!
	return nil
}

func MoveFile(dst, src string) error {
	return os.Rename(src, dst)
}

var simulate bool
var move bool
var dest string

func init() {
	defer func() {
		args := flag.Args()
		if len(args) != 1 {
			panic("Usage: collectf [-simulate] [-move] dest")
		}
		dest = args[0]
	}()
	defer flag.Parse()

	flag.BoolVar(&simulate, "simulate", simulate, "Do not copy any files instead only print what would be done")
	flag.BoolVar(&move, "move", move, "Move files instead of copying them")

}
