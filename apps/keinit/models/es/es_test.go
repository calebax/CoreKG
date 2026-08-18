package es

import (
	"context"
	"testing"

	"github.com/ygpkg/yg-go/logs"
)

func TestES(t *testing.T) {
	res, err := parseMultipleDSLFile("/usr/local/goProject/src/roc/scripts/es/v1.7.0__mapping.dsl")
	if err != nil {
		logs.ErrorContextf(context.TODO(), "parseMultipleDSLFile err: %v", err)
	}
	for _, v := range res {
		logs.InfoContextf(context.TODO(), "parseMultipleDSLFile res: %+v", v)
	}
}
