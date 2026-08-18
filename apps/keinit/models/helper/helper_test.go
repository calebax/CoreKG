package helper

import (
	"fmt"
	"sort"
	"testing"
)

func TestSort(t *testing.T) {
	sqlFiles := []string{
		"insert_agent_i18n",
		"a.10.0__insert_agent_i18n",
		"v1.9_0__create_table",
		"v1.10_1__insert_agent_i18n",
		"v1.10.0__insert_agent_i18n",
		"v1.8_2__create_user",
		"v1.8_3__create_user",
		"v1.8_0__create_user",
		"v2.8_1__create_user",
		"zinsert_agent_i18n",
	}

	sort.Sort(VersionSlice(sqlFiles))

	for _, file := range sqlFiles {
		fmt.Println(file)
	}
}
