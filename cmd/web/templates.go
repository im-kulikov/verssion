package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strings"

	libw "github.com/alicebob/verssion/w"
)

var (
	baseTempl = template.Must(
		template.New("base").
			Funcs(template.FuncMap{
				"title": libw.Title,
				"link": func(s string) string {
					return *baseURL + s
				},
				"version": func(s string) template.HTML {
					h := template.HTMLEscapeString(s)
					return template.HTML(strings.Replace(h, "\n", "<br />", -1))
				},
			}).Parse(`<!DOCTYPE html>
<html>
    <head>
        <title>{{ .title }}</title>
        <style type="text/css">
td {
    vertical-align: top;
}
        </style>
        {{- template "head" . }}
    </head>
    <body>
        <a href="{{link "/"}}">Home</a><br />
        <hr />
        {{- template "page" . }}
    </body>
</html>
{{define "head"}}
{{- end}}
{{define "page"}}
{{- end}}

{{define "errors"}}
    {{- with .}}
        Some problems:<br />
        {{- range .}}
            {{.}}<br />
        {{- end}}
        <br />
        <br />
    {{- end}}
{{end}}

{{define "pageselection"}}
    {{- if .pages}}
    Selected pages:<br />
    {{- range .pages}}
        <input type="checkbox" name="p" value="{{.}}" id="p{{.}}"{{if (index $.selected .)}} CHECKED{{end}}/><label for="p{{.}}" title="{{.}}"> {{title .}}</label><br />
    {{- end}}
    <br />
    {{- end}}

    Add some pages:<br />
    {{- range .available}}
        <input type="checkbox" name="p" value="{{.}}" id="p{{.}}"{{if (index $.selected .)}} CHECKED{{end}}/><label for="p{{.}}" title="{{.}}"> {{title .}}</label><br />
    {{- end}}
    <br />

    Or add other en.wikipedia.org pages (either the full URL or the part after <code>/wiki/</code>). One per line.<br />
    <textarea name="etc" cols="80" rows="4">{{.etc}}</textarea><br />
{{end}}
`))
)

func extend(t *template.Template) *template.Template {
	return template.Must(t.Clone())
}

func runTmpl(w http.ResponseWriter, t *template.Template, args interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	b := &bytes.Buffer{}
	if err := t.Execute(b, args); err != nil {
		log.Printf("template: %s", err)
		http.Error(w, "internal server error", 500)
		return
	}
	w.Write(b.Bytes())
}
