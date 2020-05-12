package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// ** TODO - 6 - connect to d.b.
// ** TODO - 7 - run .data files to get "data"

// TODO - 2 - Implement a "path" search instead of a regular one.

// TODO/Issue/Defect - 1000 - known problem - time change on files is not working (xyzzy1000)
// TODO - 3 - Test file modification
// TODO - 5 - figure out how to test caching / file modification / auto-reload

var optPath = flag.String("path", "./tmpl1", "Path to look for templates in")
var optOutput = flag.String("output", "", "File to send output to")
var optData = flag.String("data", "", "Data file in JSON")
var optDebug = flag.String("debug", "", "comma seperated list of debug flags")
var optURLPath = flag.Bool("urlpath", false, "Simulate a URL Path")
var optSleep = flag.Bool("sleep", false, "Simulate a URL Path")
var optDbConn = flag.String("conn", "", "Database (PostgreSQL) connection string.")
var optDbName = flag.String("dbname", "", "Database (PostgreSQL) name.")
var optQuery = flag.String("sql", "", "Database (PostgreSQL) select to get data.")
var optUseSubData = flag.Bool("sub-data", false, "use .data as a field for array of data.")

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

	mdata := map[string]interface{}{}

	if *optDbConn != "" {
		db_x := ConnectToAnyDb("postgres", *optDbConn, *optDbName)
		if db_x == nil {
			fmt.Fprintf(os.Stderr, "%sUnable to connection to database: s\n", MiscLib.ColorRed, MiscLib.ColorReset)
			os.Exit(1)
		}
		data, err := SelData2(db_x.Db, *optQuery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sUnable to connection to database/failed on table select: %v%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
			os.Exit(1)
		}
		if DbOn["query"] {
			fmt.Printf("Data=%s\n", godebug.SVarI(data))
		}

		if *optUseSubData {
			mdata = map[string]interface{}{
				"data": data,
			}
		} else if len(data) == 1 {
			mdata = data[0]
		} else if len(data) > 1 {
			fmt.Printf("Warning - %d rows returend from %s, using 0th row\n", len(data), *optQuery)
			mdata = data[0]
		} else if len(data) == 0 {
			fmt.Printf("Warning - 0 rows returend from %s\n", *optQuery)
		}

	} else if *optData != "" {
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

		if *optURLPath {
			nfn := pathResolve(fn)
			fmt.Fprintf(out, "\n----------------------------- %s < %s ------------------------------\n", nfn, fn)
			fn = nfn
		} else {
			fmt.Fprintf(out, "\n----------------------------- %s ------------------------------\n", fn)
		}

		rv, err := PcrF(*optPath, fn, mdata)
		// base.html ex1.tmpl ex2.tmpl

		fmt.Fprintf(out, "rv ->%s<- err ->%s<-\n", rv, err)
		if DbOn["sleep"] {
			time.Sleep(15 * time.Second)
		}
	}
}

// if index.html?page_name=ABC then
//	render index.html with ABC ad the underying tempalte?
// if ABC then
//	just render the tempalte ABC and retur the results.
func pathResolve(fn string) (rv string) {
	rv, ok := GetVal(fn, "page_name", "index.html")
	if !ok {
		rv = "index.html"
	}
	return
}

func GetVal(fn, name, dflt string) (val string, found bool) {
	re := regexp.MustCompile("^.*\\?page_name=")
	if re.MatchString(fn) {
		found = true
		val = re.ReplaceAllString(fn, "")
		fmt.Printf("Matched orig ->%s<- new ->%s<-\n", fn, val)
	} else {
		found = true
		val = fn
	}
	return
}
