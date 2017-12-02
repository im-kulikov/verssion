package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	libw "github.com/alicebob/verssion/w"
)

func allPagesHandler(base string, db libw.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		all, err := db.CurrentAll()
		if err != nil {
			log.Printf("current all: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		runTmpl(w, allPagesTempl, map[string]interface{}{
			"base":  base,
			"title": "Pages overview",
			"pages": all,
		})
	}
}

func pageHandler(base string, db libw.DB, fetch Fetcher) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		page := p.ByName("page")
		cur, err := loadPage(page, db, fetch)
		if err != nil {
			if p, ok := err.(libw.ErrNotFound); ok {
				log.Printf("not found %q: %s", page, err)
				w.WriteHeader(404)
				runTmpl(w, pageNotFoundTempl, map[string]interface{}{
					"base":      base,
					"title":     libw.Title(p.Page),
					"wikipedia": WikiURL(p.Page),
					"page":      p.Page,
				})
				return
			}
			log.Printf("update %q: %s", page, err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if page != cur.Page {
			w.Header().Set("Location", fmt.Sprintf("%s/p/%s/", base, cur.Page))
			w.WriteHeader(302)
			return
		}

		vs, err := db.History(cur.Page)
		if err != nil {
			log.Printf("history: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		runTmpl(w, pageTempl, map[string]interface{}{
			"base":      base,
			"title":     libw.Title(cur.Page),
			"atom":      adhocURL(base, []string{cur.Page}),
			"wikipedia": WikiURL(cur.Page),
			"current":   cur,
			"page":      cur.Page,
			"versions":  vs,
		})
	}
}

var (
	allPagesTempl = withBase(`
{{define "page"}}
	<h2>All known pages</h2>
	<table>
	{{- range .pages}}
		<tr>
			<td><a href="./{{.Page}}/" title="{{.Page}}">{{title .Page}}</a></td>
			<td>{{version .StableVersion}}</td>
		</tr>
	{{- end}}
	</table>
	<br />
	<a href="{{.base}}/curated/">Make a custom feed</a><br />
{{- end}}
`)

	pageTempl = withBase(`
{{define "head"}}
	<link rel="alternate" type="application/atom+xml" title="Atom 1.0" href="{{.atom}}"/>
{{- end}}
{{define "page"}}
	<h2>{{.title}}</h2>
	<table>
		<tr>
			<td>Wikipedia:</td>
			<td><a href="{{.wikipedia}}">{{.wikipedia}}</a></td>
		</tr>
		<tr>
			<td>Homepage:</td>
			<td>{{with .current.Homepage}}<a href="https://{{.}}">https://{{.}}</a>{{- end}}</td>
		</tr>
		<tr>
			<td>Current stable version:</td>
			<td>{{with .current.StableVersion}}{{version .}}{{- end}}</td>
		</tr>
	</table>
    <br />
    <br />
	<h2>Version history</h2>
	<table>
	<tr>
		<th>Spider timestamp:</th>
		<th>Version:</th>
	</tr>
	{{- range .versions}}
		<tr>
			<td>{{.T.Format "2006-01-02 15:04 UTC"}}</td>
			<td>{{version .StableVersion}}</td>
		</tr>
	{{- end}}
	</table>
	<br />
	RSS link: <a href="{{.atom}}">Atom feed</a><br />
	Latest spider check: {{if not .current.T.IsZero}}{{.current.T.Format "2006-01-02 15:04 UTC"}}{{- end}}<br />
{{- end}}
`)

	pageNotFoundTempl = withBase(`
{{define "page"}}
	Page not found: {{.page}}<br />
	Maybe you can create it on <a href="{{.wikipedia}}">Wikipedia</a><br />
    <br />
{{- end}}
`)
)