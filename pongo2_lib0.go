package main

// xyzzy - User Context - what is the user data
// xyzzy - User Query - have a query object that can provide context
// xyzzy - Caching (xyzzy-9292)
// xyzzy - HttpServer ( w, r ) -> all PcrF w/ User Context based on user_id?

/*

Cookies:
	1. jwt_token -> user_id
	2. session_id -> UUID for per browser client data
	3. page_name -> name of page/state to display

Cycle
	Click on Item -> Action -> GET(Action/Page) -> paint into page "data-target=ID" to paint to
	Form - Display/Validate -> Click on button -> FormAction -> ACTION-tag -> GET/POST -> JSON data / page_name ===>>> render page w/ Data

--data-- example.data
	$1 = {{.user_id}}
	$2 = {{.session_id}}
	redis key err = get Name
	pg docs err = select * from t_ymux_documents where extension in ( 'xls', 'xlsx' ) and user_id = $1;
	pongo2 macro_name = "String"
	tmpl tmpl_name = "String"
	set Name = "String"
	return "String"
	<other>
--template--
	<ol>
	%{ for doc in docs %}
		<li> {{ user_details ( doc ) }} </li>
	${ endfor %}
	</ol>


--------------------------------------------------------------------------------------------------------------------------
-- A read cache - pull back file only if it has changed?
-- A "state" that is kept in Redis based on "session_id" {containing "user_id", current page}
-- A /index.html?page=PageName referrer that will re-paint so bookmarks can work.
--------------------------------------------------------------------------------------------------------------------------

1 single file? - no -

1.
	<form ... ACTION="/api/v1/do-someting" data-success-next-page="xyzzy" data-fail-msg="xyzzy">
		<button class="bind-action">
	</form>
2.
	function actionHandler ( this ) {
		// 3.
		$.ajax(...
			success: function ( data ) {
				if ( data.page_name ) {
					// 4.
					setHistory ( "/index.html?page_name="+data.page_name )
					paintHtml ( data, form.tag_next || data.tag_next, data.page_name );
				}
			error: function ( err... ) {
				}
			}
		)
	}

5. On server side - json = MergeNextPae ( original-data, "page-name" )
5. On server side http://.../index.html?page_name=X
	Paint default index.html w/ page_name as template in body.

*/

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/filelib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/pongo2"
)

func Pcr(tmpl string, data map[string]interface{}) (rv string, err error) {
	// Compile the template first (i. e. creating the AST)
	tpl, e2 := pongo2.FromString(tmpl)
	if e2 != nil {
		err = fmt.Errorf("Template failed to compile: %s", e2)
		fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), err)
		return
	}
	// Now you can render the template with the given
	// pongo2.Context how often you want to.
	out, e2 := tpl.Execute(data)
	if e2 != nil {
		err = fmt.Errorf("Template failed to run: %s", e2)
		fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), err)
		return
	}
	rv = out
	return
}

func PcrX(tmpl string, data map[string]interface{}) (rv string) {
	r, e := Pcr(tmpl, data)
	if e != nil {
		r = fmt.Sprintf("Error: %s\n", e)
		fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), e)
	}
	rv = r
	return
}

type ATmplCacheType struct {
	template    *pongo2.Template
	when        int64 // when last used
	whenModTime int64 // File sytem last modified
}

type TmplCacheType struct {
	cache     map[string]ATmplCacheType
	cacheLock sync.Mutex
	logger    *os.File
}

func NewTmplCache() (tc *TmplCacheType) {
	return &TmplCacheType{
		cache:  make(map[string]ATmplCacheType),
		logger: os.Stderr,
	}
}

func (tc *TmplCacheType) TmplCache(name string) (tpl *pongo2.Template, cMt int64, err error) {
	var found bool
	tc.cacheLock.Lock()
	defer tc.cacheLock.Unlock()
	t, found := tc.cache[name]
	if !found {
		err = fmt.Errorf("Tempalte Not Found")
		if DbOn["cache"] {
			fmt.Fprintf(os.Stderr, "%s%s Not found in cache%s\n", MiscLib.ColorCyan, name, MiscLib.ColorReset)
		}
	} else {
		if DbOn["cache"] {
			fmt.Fprintf(os.Stderr, "%s%s found in cache%s\n", MiscLib.ColorRed, name, MiscLib.ColorReset)
		}
	}
	t.when = int64(time.Now().Unix()) // Update Timestamp of Use
	tc.cache[name] = t
	return t.template, t.whenModTime, nil
}

func (tc *TmplCacheType) SaveInCache(name string, tpl *pongo2.Template, mTime int64) {
	if DbOn["cache"] {
		fmt.Fprintf(os.Stderr, "%sLoad %s to cache%s\n", MiscLib.ColorCyan, name, MiscLib.ColorReset)
	}
	tc.cacheLock.Lock()
	defer tc.cacheLock.Unlock()
	tc.cache[name] = ATmplCacheType{
		template:    tpl,
		when:        time.Now().Unix(),
		whenModTime: mTime, // Set Timestamp for 1st use
	}
}

func (tc *TmplCacheType) FlushCache() {
	tc.cache = make(map[string]ATmplCacheType)
}

// xyzzy discard old - time of save? LRU?
// Set timeout - for how often to check - and oldest use.

var tc *TmplCacheType = NewTmplCache()

// Look in the tmplDir for the tmplName.  If found then run the template.
// 1. How to handle errors
//		a. Log them
//		b. return an error message
//		c. return an error code? 418?
func PcrF(tmplDir, tmplName string, data map[string]interface{}) (rv string, err error) {
	base_fn := filepath.Join(tmplDir, tmplName)
	// fn := base_fn + ".tmpl"
	fn := base_fn
	data_fn := base_fn + ".data"
	_ = data_fn
	var tpl *pongo2.Template
	var cMt int64
	if tpl, cMt, err = tc.TmplCache(fn); err != nil {
		if !filelib.Exists(filepath.Join(fn)) {
			err = fmt.Errorf("Unable to find file")
			fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), err)
			return
		}
		finfo, e2 := os.Stat(fn)
		if e2 != nil {
			err = fmt.Errorf("Unable to access file: %s", e2)
			fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), err)
			return
		}
		mt := finfo.ModTime().Unix()
		fmt.Printf("cMt=%d mt=%d\n", cMt, mt)
		if cMt != mt {
			if DbOn["cache"] {
				fmt.Fprintf(os.Stderr, "%sLoad %s to cache - timestamp changed%s\n", MiscLib.ColorYellow, tmplName, MiscLib.ColorReset)
			}
			tpl, e2 = pongo2.FromFile(fn)
			if e2 != nil {
				err = fmt.Errorf("Unable to parse template:%s", fn)
				fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), err)
				return
			}
			tc.SaveInCache(fn, tpl, mt)
		}
	} else {
		if !filelib.Exists(filepath.Join(fn)) {
			err = fmt.Errorf("Unable to find file")
			fmt.Fprintf(tc.logger, "AT: %s file: ->%s<- err: %s\n", godebug.LF(), fn, err)
			return
		}
		finfo, e2 := os.Stat(fn)
		if e2 != nil {
			err = fmt.Errorf("Unable to access file: %s", e2)
			fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), err)
			return
		}
		tpl, e2 = pongo2.FromFile(fn)
		if e2 != nil {
			err = fmt.Errorf("Unable to parse template:%s", fn)
			fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), err)
			return
		}
		mt := finfo.ModTime().Unix()
		tc.SaveInCache(fn, tpl, mt)
	}

	out, e2 := tpl.Execute(data)
	if e2 != nil {
		err = fmt.Errorf("Template (%s) failed to run: %s", fn, e2)
		fmt.Fprintf(tc.logger, "AT: %s err: %s\n", godebug.LF(), err)
		return
	}

	rv = out
	return
}
