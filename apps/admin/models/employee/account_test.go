package employee

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGeneratePassword(t *testing.T) {
	ret, _ := bcrypt.GenerateFromPassword([]byte("QizFRfDcJ"), 12)
	t.Logf("%s", ret)
}
