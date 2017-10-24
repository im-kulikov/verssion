package w

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestFindTables(t *testing.T) {
	type cas struct {
		Html string
		Want []Table
	}
	cases := []cas{
		{
			Html: `<table>string</table>`,
			Want: []Table{
				Table{
					Rows: [][]string(nil),
				},
			},
		},
		{
			Html: `<html><body><table><tr><td>foo</td><td>bar</td></table>`,
			Want: []Table{
				{
					Rows: [][]string{{"foo", "bar"}},
				},
			},
		},
	}
	for i, c := range cases {
		d, err := FindTables(bytes.NewBufferString(c.Html))
		if err != nil {
			t.Fatal(err)
		}
		if have, want := d, c.Want; !reflect.DeepEqual(have, want) {
			t.Errorf("case %d: have %#v, want %#v", i, have, want)
		}
	}
}

func TestFindTablesReal(t *testing.T) {
	r, err := os.Open("./data/git.html")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	ts, err := FindTables(r)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := len(ts), 6; have != want {
		t.Errorf("have %v, want %v", have, want)
	}
	t1 := Table{Rows: [][]string{
		[]string{""},
		[]string{"A command-line session showing repository creation, addition of a file, and remote synchronization"},
		[]string{"Original author(s)", "Linus Torvalds"},
		[]string{"Developer(s)", "Junio Hamano and others"},
		[]string{"Initial release", "7 April 2005; 12 years ago (2005-04-07)"},
		[]string{""},
		[]string{"Stable release", "2.14.2 / 22 September 2017; 32 days ago (2017-09-22)"},
		[]string{""},
		[]string{"Repository", "git-scm.com/downloads"},
		[]string{"Development status", "Active"},
		[]string{"Written in", "C, Shell, Perl, Tcl, Python"},
		[]string{"Operating system", "POSIX: Linux, Windows, macOS"},
		[]string{"Platform", "IA-32, x86-64"},
		[]string{"Available in", "English"},
		[]string{"Type", "Version control"},
		[]string{"License", "GNU GPL v2 and GNU LGPL v2.1"},
		[]string{"Website", "git-scm.com"},
	}}
	if have, want := ts[0], t1; !reflect.DeepEqual(have, want) {
		t.Errorf("have %#v\nwant %#v", have, want)
	}
	// for _, tb := range ts {
	// t.Logf("table: %#v\n\n", tb)
	// }
}
