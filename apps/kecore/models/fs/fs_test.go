package fs

import (
	"context"
	"testing"
)

func TestSplitHost(t *testing.T) {
	// print(SplitHost("http://localhost:8080/aa/bb/c.jpg"))
	println(SpliceUrl(context.Background(), "aa/bb/c.jpg", "http://39.175.132.229:18020/search?search=%257B%2522text%2522%253A%2522%2522%252C%2522img%2522%253A%2522%252Fcorekg-bucket%252Fyg-chat%252F20250822%252F1-CCdx11JuQ.png%2522%252C%2522type%2522%253A%2522all%2522%257D"))
}
