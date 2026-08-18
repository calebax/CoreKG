package kellm

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type testStreamWriter struct {
	header     http.Header
	body       bytes.Buffer
	flushCount int
	statusCode int
}

func (w *testStreamWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *testStreamWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *testStreamWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *testStreamWriter) Flush() {
	w.flushCount++
}

func TestCopyStreamBodyPreservesSSEBlocks(t *testing.T) {
	input := "data: {\"id\":\"1\"}\n\n" +
		"data: {\"id\":\"2\"}\n\n" +
		"data: [DONE]\n\n"

	writer := &testStreamWriter{}
	if err := copyStreamBody(writer, io.NopCloser(bytes.NewBufferString(input))); err != nil {
		t.Fatalf("copyStreamBody() error = %v", err)
	}

	if writer.body.String() != input {
		t.Fatalf("copyStreamBody() body = %q, want %q", writer.body.String(), input)
	}
	if writer.flushCount != 4 {
		t.Fatalf("copyStreamBody() flushCount = %d, want 4", writer.flushCount)
	}
}
