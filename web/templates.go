package web

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/alicebob/verssion/core"
)

var (
	baseTempl = template.Must(
		template.New("base").
			Funcs(template.FuncMap{
				"title": core.Title,
				"version": func(s string) template.HTML {
					h := template.HTMLEscapeString(s)
					t := template.HTML(strings.Replace(h, "\n", "<br />", -1))
					return t
				},
			}).Parse(`<!DOCTYPE html>
<html>
<head>
	<title>{{ .title }}</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="shortcut icon" href="{{.base}}/s/favicon.png" type="image/png" sizes="16x16 24x24 32x32 64x64">
	<link rel="apple-touch-icon" href="{{.base}}/s/favicon.png">
	<style type="text/css">
body {
	margin: 0;
}
table {
	width: 100%;
	border-collapse: collapse;
}
th {
	text-align: left;
	font-weight: normal;
}
td, th {
    vertical-align: top;
	padding: 0;
	padding-bottom: 3px;
}
td:first-child {
	width: 250px;
}
textarea {
	box-sizing: border-box;
	width: 100%;
}
h2 {
	border-bottom: 1px solid #ddd;
}
a, a:visited {
	color: #357cb7;
}
a:hover {
	color: black;
}
.head {
	background-color: #35b7b1;
	padding: 0.5em 0;
}
.head a {
	color: black;
	text-decoration: none;
}
.head a:hover {
	text-decoration: underline;
}
.body, .head div {
	margin: 0 auto;
	max-width: 760px;
	padding: 0 0.5em;
}
.body p {
	text-align: justify;
}

@media only screen and (max-width: 700px) {
	table, thead, tbody, tr, th, td {
		display: block;
	}

	.optional {
		display: none;
	}
}
	</style>
	{{- block "head" .}}{{end}}
</head>
<body>
	<div class="head">
		<div>
        <a href="{{.base}}/">Home</a>
        - <a href="{{.base}}/curated/">New feed</a>
        - <a href="{{.base}}/p/">All pages</a>
		</div>
	</div>
	<div class="body">
        {{- block "page" .}}{{end}}
	</div>
</body>
</html>

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
    <textarea name="etc" rows="4">{{.etc}}</textarea><br />
{{end}}
`))
)

func withBase(s string) *template.Template {
	return template.Must(template.Must(baseTempl.Clone()).Parse(s))
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
