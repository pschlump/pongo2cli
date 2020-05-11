package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// TODO - 3 - Test file modification
// TODO - 5 - figure out how to test caching / file modification / auto-reload
// TODO - 7 - run .data files to get "data"

// TODO - 6 - connect to d.b.

// TODO - 2 - Implement a "path" search instead of a regular one.

var optPath = flag.String("path", "./tmpl1", "Path to look for templates in")
var optOutput = flag.String("output", "", "File to send output to")
var optData = flag.String("data", "", "Data file in JSON")
var optDebug = flag.String("debug", "", "comma seperated list of debug flags")

var out *os.File
var DbOn map[string]bool

func init() {
	out = os.Stdout
	DbOn = make(map[string]bool)
}

func main() {

	flag.Parse()

	if *optDebug != "" {
		ss := strings.Split(*optDebug, ",")
		for _, s := range ss {
			DbOn[s] = true
		}
	}

	fns := flag.Args()

	mdata := map[string]interface{}{
		"name": "kegan",
	}

	if *optData != "" {
		buf, err := ioutil.ReadFile(*optData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to open data file (%s) for input: %s\n", *optData, err)
			os.Exit(1)
		}

		err = json.Unmarshal(buf, &mdata)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse data file (%s) for input: %s\n", *optData, err)
			os.Exit(1)
		}
	}

	for _, fn := range fns {

		fmt.Fprintf(out, "\n----------------------------- %s ------------------------------\n", fn)
		rv, err := PcrF(*optPath, fn, mdata)
		// base.html ex1.tmpl ex2.tmpl

		fmt.Fprintf(out, "rv ->%s<- err ->%s<-\n", rv, err)
		time.Sleep(15 * time.Second)
	}

}
