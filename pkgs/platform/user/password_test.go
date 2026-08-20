package user

import (
	"context"
	"fmt"
	"testing"
)

func TestVerifyPassword(t *testing.T) {
	pwd := "zhengzihao123"
	encPwd := "CHANGE_ME_PASSWORD_HASH"
	res := VerifyPassword(context.Background(), pwd, encPwd)
	fmt.Println("res: ", res)
}
