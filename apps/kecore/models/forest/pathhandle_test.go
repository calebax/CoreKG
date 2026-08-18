package forest

import "testing"

func TestSplitPath2(t *testing.T) {
	paths := []string{
		"",
		"a.pdf",
		"a/",
		"a/b/",
		"a/b/c.pdf",
		"/a/",
	}
	for _, p := range paths {
		dirs, f := SplitPath(p)
		t.Logf("%s--%v-%d-%v", p, dirs, len(dirs), f)
	}
}
