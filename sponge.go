/*
sponge: soak up standard input and write to a file

Synopsis

sed '...' file | grep '...' | sponge file

Description

sponge reads standard input and writes it out to the specified file. Unlike a shell redirect, sponge soaks up all its input before opening the output file. This allows constricting pipelines that read from and write to the same file.

If no output file is specified, sponge outputs to stdout.

Source: https://linux.die.net/man/1/sponge
*/
package main // go.jonnrb.io/sponge

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func getSink() io.Writer {
	switch len(os.Args) {
	case 0:
		panic("congrats: you won the game")
	case 1:
		return os.Stdout
	case 2:
		dst := os.Args[1]
		f, err := os.Create(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening %s: %v\n", dst, err)
			os.Exit(1)
		}
		return f
	default:
		fmt.Fprintf(os.Stderr, "usage: %s [file]\n", os.Args[0])
		os.Exit(2)
		panic("you must be doing really well for yourself")
	}
}

func sponge(in io.Reader, out io.Writer) error {
	b := bytes.Buffer{}
	if _, err := io.Copy(&b, in); err != nil {
		return err
	}
	if _, err := io.Copy(out, &b); err != nil {
		return err
	}
	return nil
}

func main() {
	src, sink := os.Stdin, getSink()
	if err := sponge(src, sink); err != nil {
		fmt.Fprintf(os.Stderr, "error during pipe: %v\n", err)
		os.Exit(3)
	}
}
