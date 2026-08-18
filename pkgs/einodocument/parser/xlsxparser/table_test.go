package xlsxparser

import "testing"

func TestNormalizeHeader(t *testing.T) {
	fieldMap := map[string]string{
		"123":    "c_123",
		"c_123":  "c_123",
		"123_c":  "c_123_c",
		"姓名":     "xing_ming",
		"a姓名":    "a_xing_ming",
		"姓名a":    "xing_ming_a",
		"姓a名":    "xing_a_ming",
		"姓名a名":   "xing_ming_a_ming",
		"姓名a名b":  "xing_ming_a_ming_b",
		"姓名a名b_": "xing_ming_a_ming_b",
		"ID":     "id",
	}
	for ori := range fieldMap {
		res := NormalizeHeader(ori)
		t.Logf("ori: %s, res: %s", ori, res)
	}
}
