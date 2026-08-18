package nebulagraph

type TagDesc struct {
	Field   string `nebula:"Field"`
	Type    string `nebula:"Type"`
	Null    string `nebula:"Null"`
	Default string `nebula:"Default"`
	Comment string `nebula:"Comment"`
}
type TagDescList []TagDesc

func (t *TagDescList) NameMap() map[string]TagDesc {
	m := make(map[string]TagDesc)
	for _, v := range *t {
		m[v.Field] = v
	}
	return m
}
