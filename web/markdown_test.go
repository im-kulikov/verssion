package web

import (
	"html/template"
	"testing"
)

func TestMarkdown(t *testing.T) {
	type cas struct {
		Src  string
		HTML string
	}
	for i, c := range []cas{
		{
			Src:  "foo bar",
			HTML: "foo bar",
		},
		{
			Src:  "foo [foo](http://bar)",
			HTML: `foo <a href="http://bar">foo</a>`,
		},
		{
			Src:  "foo [more words!!](http://bar)",
			HTML: `foo <a href="http://bar">more words!!</a>`,
		},
		{
			Src:  "foo [foo](http://foo)[bar](http://bar/foo/etc.html)",
			HTML: `foo <a href="http://foo">foo</a><a href="http://bar/foo/etc.html">bar</a>`,
		},
		{
			Src:  "foo [foo](mailto://bar)",
			HTML: "foo [foo](mailto://bar)",
		},
		{
			Src:  "foo [<b>foo!](http://bar)",
			HTML: `foo <a href="http://bar">&lt;b&gt;foo!</a>`,
		},
		{
			Src:  "[mariadb.org](https://mariadb.org/), [mariadb.com](https://mariadb.com/)",
			HTML: `<a href="https://mariadb.org/">mariadb.org</a>, <a href="https://mariadb.com/">mariadb.com</a>`,
		},
	} {
		if have, want := miniMarkdown(c.Src), template.HTML(c.HTML); have != want {
			t.Errorf("case %d: have %q, want %q", i, have, want)
		}
	}
}
