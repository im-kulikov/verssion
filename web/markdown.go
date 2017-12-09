// only knows about "[link](https://..)" markdown markup
package web

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

var (
	linkRe = regexp.MustCompile(`\[([^]]+)\]\(([^)\s]+)\)`)
)

func miniMarkdown(src string) template.HTML {
	return template.HTML(linkRe.ReplaceAllStringFunc(
		src,
		func(m string) string {
			pts := linkRe.FindStringSubmatch(m)
			if pts == nil {
				return m
			}
			url, name := pts[2], pts[1]
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return m
			}
			return fmt.Sprintf(`<a href="%s">%s</a>`,
				template.HTMLEscapeString(url),
				template.HTMLEscapeString(name),
			)
		},
	))
}
