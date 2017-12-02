package web

import (
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	libw "github.com/alicebob/verssion/w"
)

func indexHandler(base string, db libw.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		es, err := db.Recent(12)
		if err != nil {
			log.Printf("current all: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		curated, err := readCuratedCookies(r, db)
		if err != nil {
			log.Printf("readCuratedCookies: %s", err)
		}
		runTmpl(w, indexTempl, map[string]interface{}{
			"base":    base,
			"title":   "VeRSSion",
			"entries": es,
			"curated": curated,
		})
	}
}

func readCuratedCookies(r *http.Request, db libw.DB) ([]libw.Curated, error) {
	var (
		curs    []libw.Curated
		lastErr error
	)
	for _, cookie := range r.Cookies() {
		t := strings.SplitN(cookie.Name, "-", 2)
		if len(t) != 2 || t[0] != "curated" {
			continue
		}
		c, err := db.LoadCurated(t[1])
		if err != nil {
			lastErr = err
		} else {
			if c != nil {
				curs = append(curs, *c)
			}
		}
	}
	return curs, lastErr
}

var (
	indexTempl = withBase(`
{{define "page"}}
<h2>What</h2>
<div><p>
Verssion(*) tracks stable version of software projects (e.g.: databases, editors, JS frameworks), and makes that available as an RSS (atom) feed. The main use-case is for dev-ops and developers who use a lot of open source software projects, and who like to keep an eye on releases. Without making that a fulltime job, and without signing up for dozens of e-mail lists. Turns out wikipedia is a great source for version information, so that's what we use.<br />
You can create feeds for your own use, or share them with colleagues.<br />
*) working title<br />
</p>
<br />
<a href="https://github.com/alicebob/verssion/">Full source</a> for issues and PRs.<br />
</div>
<br />

<h2>Feed</h2>
Make a feed which combines multiple projects in a single feed:<br />
<a href="./curated/">Create new feed!</a><br />
	{{- if .curated}}
		<br />
		Your recent feeds:<br />
		{{- range .curated}}
			- <a href="{{$.base}}/curated/{{.ID}}">{{.Title}}</a><br />
		{{- end}}
	{{- end}}
<br />

<h2>Updates</h2>
	<table>
	{{- range .entries}}
		<tr>
			<td><a href="./p/{{.Page}}/">{{title .Page}}</a></td>
			<td>{{version .StableVersion}}</td>
		</tr>
	{{- end}}
		<tr>
			<td><a href="./p/">...</a></td>
			<td></td>
		</tr>
	</table>
	<br />

{{- end}}
`)
)