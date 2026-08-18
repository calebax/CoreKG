package utils

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUnixMicro(t *testing.T) {
	timeStamp := GetUnixMicro(time.Hour * 24 * 90)
	timeStampStr := strconv.FormatInt(timeStamp, 10)
	t.Log("GetUnixMicro: ", timeStampStr)

	// 正确转换
	tsInt, err := strconv.ParseInt(timeStampStr, 10, 64)
	assert.Nil(t, err)
	datetime := time.UnixMicro(tsInt)
	t.Log("datetime: ", datetime.Format(time.DateTime))
}
