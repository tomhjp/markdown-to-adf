package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tomhjp/markdown-to-adf/renderer"
)

var output = flag.String("o", "", "output file to write, defaults to stdout if not set")

func usage() {
	fmt.Fprintf(os.Stderr, "usage: markdown-to-adf [flags] path\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}

	source, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Printf("Reading file failed: %v", err)
		os.Exit(1)
	}

	var w io.WriteCloser
	if *output == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(*output)
		if err != nil {
			fmt.Printf("Creating output file failed: %v", err)
			os.Exit(1)
		}
		defer w.Close()
	}

	if err = renderer.Render(w, source); err != nil {
		fmt.Printf("Rendering adf failed: %v", err)
		os.Exit(1)
	}

	if *output != "" {
		fmt.Println("Output file created successfully.")
	}
}
