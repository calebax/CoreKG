package types

import (
	"context"
	"encoding/json"

	"github.com/ygpkg/yg-go/logs"
)

// StringArray 字符串数组，用于Mysql数据库和接口之间的转换
type StringArray string

func NewStringArray(v []string) StringArray {
	if len(v) == 0 {
		return StringArray("")
	}
	bs, _ := json.Marshal(v)
	return StringArray(bs)
}

// MarshalJSON .
func (i StringArray) MarshalJSON() ([]byte, error) {
	arr := i.Slice()
	return json.Marshal(arr)
}

// UnmarshalJSON .
func (i *StringArray) UnmarshalJSON(data []byte) error {
	it := StringArray(string(data))
	*i = it
	return nil
}

// Slice .
func (i StringArray) Slice() (arr []string) {
	arr = []string{}
	if i == "" {
		return
	}

	err := decodeJSON(string(i), &arr)
	if err != nil {
		logs.ErrorContextf(context.TODO(), "array %s decode failed, %s", i, err)
		return
	}
	return
}

func decodeJSON(bs string, v interface{}) error {
	return json.Unmarshal([]byte(bs), v)
}

func (i *StringArray) Append(value string) {
	arr := i.Slice()
	arr = append(arr, value)
	*i = NewStringArray(arr)
}

func (s *StringArray) Remove(item string) {
	arr := s.Slice()
	var newArray []string
	for _, v := range arr {
		if v != item {
			newArray = append(newArray, v)
		}
	}
	*s = NewStringArray(newArray)
}
