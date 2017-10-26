package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alicebob/w/w"
	"github.com/julienschmidt/httprouter"
)

var (
	baseURL = flag.String("base", "http://localhost:3141", "base URL")
	dbURL   = flag.String("db", "postgresql:///w", "postgres URL")
	listen  = flag.String("listen", ":3141", "http listen")
	updates = flag.Bool("update", true, "update pages")
)

func main() {
	flag.Parse()
	pages := flag.Args()
	if len(pages) != 0 {
		fmt.Fprintf(os.Stderr, "no args accepted\n")
		os.Exit(2)
	}

	db, err := w.NewPostgres(*dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pg: %s\n", err)
		os.Exit(2)
	}

	up := newUpdate(db)
	if !*updates {
		up = nil
	}

	r := httprouter.New()
	r.GET("/", indexHandler(db))
	r.GET("/adhoc/", adhocHandler(db, up))
	r.GET("/adhoc/atom.xml", adhocAtomHandler(db, up))
	r.GET("/v/:page/", pageHandler(db, up))
	r.GET("/v/:page/atom.xml", pageAtomHandler(db, up))
	fmt.Printf("listening on %s...\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, r))
}

func indexHandler(db w.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		es, _ := db.Recent()
		runTmpl(w, indexTempl, map[string]interface{}{
			"title":   "hello world",
			"entries": es,
		})
	}
}

func adhocHandler(db w.DB, up *update) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		pages := r.URL.Query()["p"]
		sort.Strings(pages)
		if up != nil {
			for _, p := range pages {
				if err := up.Update(p); err != nil {
					log.Printf("update %q: %s", p, err)
				}
			}
		}

		vs, err := db.History(pages...)
		if err != nil {
			log.Printf("history: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		runTmpl(w, adhocTempl, map[string]interface{}{
			"title":    strings.Join(pages, ", "),
			"pages":    pages,
			"versions": vs,
			"atom":     adhocURL(pages),
		})
	}
}

func adhocAtomHandler(db w.DB, up *update) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		pages := r.URL.Query()["p"]
		sort.Strings(pages)
		if up != nil {
			for _, p := range pages {
				if err := up.Update(p); err != nil {
					log.Printf("update %q: %s", p, err)
				}
			}
		}

		vs, err := db.History(pages...)
		if err != nil {
			log.Printf("history: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		var es []Entry
		for _, v := range vs {
			es = append(es, Entry{
				ID:      asuri(v.Page, v.StableVersion),
				Title:   v.Page + ": " + v.StableVersion,
				Updated: v.T,
				Content: v.StableVersion, // TODO: prev version?
			})
		}
		url := adhocURL(pages)
		var update time.Time
		if len(vs) > 0 {
			update = vs[len(vs)-1].T
		}
		writeFeed(w, url, strings.Join(pages, ", "), update, es)
	}
}

func pageHandler(db w.DB, up *update) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		page := p.ByName("page")
		if up != nil {
			if err := up.Update(page); err != nil {
				log.Printf("update %q: %s", page, err)
			}
		}

		vs, err := db.History(page)
		if err != nil {
			log.Printf("history: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		runTmpl(w, pageTempl, map[string]interface{}{
			"title":    page,
			"page":     page,
			"versions": vs,
		})
	}
}

func pageAtomHandler(db w.DB, up *update) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		page := p.ByName("page")
		if up != nil {
			if err := up.Update(page); err != nil {
				log.Printf("update %q: %s", page, err)
			}
		}

		vs, err := db.History(page)
		if err != nil {
			log.Printf("history: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		if len(vs) == 0 {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		var (
			es   []Entry
			prev = "?"
		)
		for _, v := range vs {
			es = append(es, Entry{
				ID:      asuri(page, v.StableVersion),
				Title:   page + ": " + v.StableVersion,
				Updated: v.T,
				Content: fmt.Sprintf("%s -> %s", prev, v.StableVersion),
			})
			prev = v.StableVersion
		}
		writeFeed(w, asuri(page), page, vs[len(vs)-1].T, es)
	}
}

func writeFeed(w http.ResponseWriter, id, title string, update time.Time, es []Entry) {
	feed := Feed{
		XMLNS:   "http://www.w3.org/2005/Atom",
		ID:      id,
		Title:   title,
		Updated: update,
		Author: Author{
			Name: "Wikipedia",
		},
		Entries: es,
	}
	w.Header().Set("Content-Type", "application/atom+xml")
	w.Write([]byte(xml.Header))
	e := xml.NewEncoder(w)
	e.Indent("", "\t")
	e.Encode(feed)
}

func adhocURL(pages []string) string {
	u, err := url.Parse(*baseURL)
	if err != nil {
		panic(err)
	}
	u.Path += "/adhoc/atom.xml"
	u.RawQuery = url.Values{
		"p": pages,
	}.Encode()
	return u.String()
}