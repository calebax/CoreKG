package decoupler

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

const (
	testCoreDSN = "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local"
)

func TestFileToPDFWithOFD(t *testing.T) {
	err := dbtools.InitMultiDBConn(map[string]string{
		"core": testCoreDSN,
	})
	if !assert.NoError(t, err) {
		return
	}

	resp, err := http.Get("https://example.com:58081/dotpen-test/apigateway/317/2026/3/16/1273e542ee6d48c584d728d6c242d81f/1273e542ee6d48c584d728d6c242d81f.ofd")
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()
	if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
		return
	}

	pdf, err := FileToPDF(context.Background(), resp.Body, "1273e542ee6d48c584d728d6c242d81f.ofd")
	if !assert.NoError(t, err) {
		return
	}
	defer pdf.Close()

	header := make([]byte, 4)
	_, err = io.ReadFull(pdf, header)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "%PDF", string(header))
}
