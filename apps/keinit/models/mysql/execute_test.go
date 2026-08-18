package mysql

import (
	"bytes"
	"testing"
)

func TestParseSQLFile(t *testing.T) {
	table := map[string][]string{
		"INSERT a;": {"INSERT a;"},

		`INSERT a
		;`: {"INSERT a ;"},

		`INSERT a

		;`: {"INSERT a ;"},

		`INSERT a
		-- help
		;`: {"INSERT a ;"},

		`-- help
		INSERT a
		;`: {"INSERT a ;"},

		`INSERT b;
		-- help
		INSERT a
		;`: {"INSERT b;", "INSERT a ;"},
	}

	for k, v := range table {
		r := bytes.NewReader([]byte(k))
		sqls, err := parseSQLReader(r)
		if err != nil {
			t.Fatal(err)
		}
		if len(sqls) != len(v) {
			t.Fatalf("want %d, got %d", len(v), len(sqls))
		}
		for i, s := range v {
			if sqls[i].Line != s {
				t.Fatalf("want %s, got %v", s, sqls[i])
			}
		}
	}
}
