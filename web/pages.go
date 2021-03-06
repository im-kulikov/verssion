package web

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/alicebob/verssion/core"
)

var matchpage = regexp.MustCompile(`^(?:(?i:https?://en.wikipedia.org)/wiki/)?(\S+)$`)

// from textarea to pages
func toPages(q string) ([]string, []error) {
	var (
		ps     []string
		errors []error
	)
	for _, l := range strings.Split(q, "\n") {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		if m := matchpage.FindStringSubmatch(l); m != nil {
			ps = append(ps, m[1])
		} else {
			errors = append(errors, fmt.Errorf("invalid page: %q", l))
		}
	}
	return ps, errors
}

func unique(ps []string) []string {
	m := map[string]struct{}{}
	for _, p := range ps {
		m[p] = struct{}{}
	}
	res := make([]string, 0, len(m))
	for p := range m {
		res = append(res, p)
	}
	sort.Strings(res)
	return res
}

func runUpdates(db core.DB, fetch Fetcher, pages []string) ([]string, []error) {
	var (
		ret    []string
		errors []error
	)

	for _, p := range pages {
		if n, err := loadPage(p, db, fetch); err != nil {
			log.Printf("update %q: %s", p, err)
			errors = append(errors, err)
		} else {
			ret = append(ret, n.Page)
		}
	}
	return ret, errors
}
