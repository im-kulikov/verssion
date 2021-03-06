package web

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/alicebob/verssion/core"
)

func newCuratedHandler(base string, db core.DB, fetch Fetcher) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		r.ParseForm()
		var (
			etc   = r.Form.Get("etc")
			pages = r.Form["p"]
		)
		pm := map[string]bool{}
		for _, p := range pages {
			pm[p] = true
		}
		args := map[string]interface{}{
			"title":    "curated list",
			"etc":      etc,
			"selected": pm,
		}
		if r.Method == "POST" {
			pages, errors := readPageArgs(db, fetch, pages, etc)
			if len(pages) > 0 && len(errors) == 0 {
				id, err := db.CreateCurated()
				if err != nil {
					log.Printf("create curated: %s", err)
					http.Error(w, http.StatusText(500), 500)
					return
				}
				if err := db.CuratedSetPages(id, pages); err != nil {
					log.Printf("curated pages: %s", err)
				}

				w.Header().Set("Location", "./"+id+"/")
				w.WriteHeader(302)
				return
			}
			args["errors"] = errors
		}
		avail, err := db.Known()
		if err != nil {
			log.Printf("known: %s", err)
		}
		args["available"] = avail
		runTmpl(w, newCuratedTempl, args)
	}
}

func curatedHandler(base string, db core.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		id := p.ByName("id")
		cur, err := db.LoadCurated(id)
		if err != nil {
			log.Printf("load curated: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		if cur == nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		vs, err := db.Current(cur.Pages...)
		if err != nil {
			log.Printf("current: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		args := map[string]interface{}{
			"curated":      cur,
			"atom":         fmt.Sprintf("%s/curated/%s/atom.xml", base, id),
			"title":        cur.Title(),
			"pageversions": vs,
		}

		c := &http.Cookie{
			Name:     "curated-" + id,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		}
		w.Header().Add("Set-Cookie", c.String())

		runTmpl(w, curatedTempl, args)
	}
}

func curatedEditHandler(base string, db core.DB, fetch Fetcher) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		id := p.ByName("id")
		cur, err := db.LoadCurated(id)
		if err != nil {
			log.Printf("load curated: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		if cur == nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		r.ParseForm()
		var (
			etc    = r.Form.Get("etc")
			qPages = r.Form["p"]
		)
		selected := map[string]bool{}
		for _, p := range cur.Pages {
			selected[p] = true
		}
		args := map[string]interface{}{
			"base":         base,
			"title":        cur.Title(),
			"curated":      cur,
			"etc":          etc,
			"selected":     selected,
			"pages":        cur.Pages,
			"defaulttitle": cur.DefaultTitle(),
			"customtitle":  cur.CustomTitle,
		}
		if r.Method == "POST" {
			pages, errors := readPageArgs(db, fetch, qPages, etc)
			title := r.Form.Get("title")
			args["customtitle"] = title
			if len(errors) == 0 {
				if err := db.CuratedSetPages(id, pages); err != nil {
					log.Printf("curated pages: %s", err)
					http.Error(w, http.StatusText(500), 500)
					return
				}

				if err := db.CuratedSetTitle(id, title); err != nil {
					log.Printf("curated title: %s", err)
				}

				w.Header().Set("Location", "./")
				w.WriteHeader(302)
				return
			}

			selected := map[string]bool{}
			for _, p := range qPages {
				selected[p] = true
			}
			args["selected"] = selected
			args["errors"] = errors
		}

		seen := map[string]struct{}{}
		for _, p := range cur.Pages {
			seen[p] = struct{}{}
		}
		var av []string
		if avail, err := db.Known(); err == nil {
			for _, p := range avail {
				if _, ok := seen[p]; !ok {
					av = append(av, p)
				}
			}
		}
		args["available"] = av
		runTmpl(w, curatedEditTempl, args)
	}
}

func curatedAtomHandler(base string, db core.DB, fetch Fetcher) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		id := p.ByName("id")
		cur, err := db.LoadCurated(id)
		if err != nil {
			log.Printf("load curated: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		if cur == nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		actualPages, _ := runUpdates(db, fetch, cur.Pages)

		vs, err := db.History(actualPages...)
		if err != nil {
			log.Printf("history: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		feed := asFeed(base, "urn:uuid:"+id, cur.Title(), cur.LastUpdated, vs)
		feed.Links = []Link{
			{
				Href: fmt.Sprintf("%s/curated/%s/", base, id),
				Rel:  "alternate", // not strictly true...
				Type: "text/html",
			},
			{
				Href: fmt.Sprintf("%s/curated/%s/atom.xml", base, id),
				Rel:  "self",
				Type: "application/atom+xml",
			},
		}
		writeFeed(w, feed)

		if err := db.CuratedSetUsed(id); err != nil {
			log.Printf("curated used %q: %s", id, err)
		}
	}
}

var (
	newCuratedTempl = withBase(`
{{define "page"}}
	Create a new list. You can change it later.<br />
	<br />

	{{template "errors" .errors}}
	
	<form method="POST">
	{{template "pageselection" .}}

	<input type="submit" name="go" value="Start a list" />
	</form>
{{- end}}
`)

	curatedTempl = withBase(`
{{define "head"}}
	<link rel="alternate" type="application/atom+xml" title="Atom 1.0" href="{{.atom}}"/>
{{- end}}
{{define "page"}}
	<h2>{{.curated.Title}}</h2>
	Atom link: <a href="{{.atom}}">{{.atom}}</a><br />
	<br />
	{{- with .pageversions}}
		<table>
		<tr>
			<th class="optional">Page:</th>
			<th class="optional">Stable version:</th>
			<th class="optional">Spider timestamp:</th>
		</tr>
		{{- range .}}
			<tr>
			<td><a href="{{$.base}}/p/{{.Page}}/" title="{{.Page}}">{{title .Page}}</a></td>
			<td>{{version .StableVersion}}</td>
			<td class="optional">{{.T.Format "2006-01-02 15:04 UTC"}}</td>
			</tr>
		{{- end}}
		</table>
	{{- else}}
		No pages selected, yet.<br />
	{{- end}}
	<br />
	<a href="./edit.html">Edit the pages in this feed</a><br />
	<br />
	<br />
{{end}}
`)

	curatedEditTempl = withBase(`
{{define "page"}}
	<h2>{{.curated.Title}}</h2>
	<br />
	<br />

	{{template "errors" .errors}}


	<form method="POST">
	Title: <input type="text" size="40" name="title" value="{{.customtitle}}" placeholder="{{.defaulttitle}}" /><br />
	{{template "pageselection" .}}
	<br />
	<input type="submit" value="Update" /><br />
	</form>
{{end}}
`)
)

// read p and etc arguments
func readPageArgs(db core.DB, fetch Fetcher, pages []string, etc string) ([]string, []error) {
	var errors []error

	etcPages, etcErrors := toPages(etc)
	pages = append(pages, etcPages...)
	errors = append(errors, etcErrors...)

	finalPages, upErrors := runUpdates(db, fetch, pages)
	errors = append(errors, upErrors...)

	return unique(finalPages), errors
}
