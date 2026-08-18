package user

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/insmtx/corekg/pkgs/utils/encryptor"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/settings"
)

func TestEnc(t *testing.T) {
	decrypt, err := encryptor.AesEncryptToBase64([]byte("G2HX+F+411lK/O8c"), []byte("12345678"))
	if err != nil {
		t.Errorf("Failed to decrypt data: %v", err)
	}
	fmt.Printf("%s\n", decrypt)
}

func TestEnc2(t *testing.T) {
	// 1. Define the Base64-encoded ciphertext string
	cipherTextB64 := "b5N/OJgDxFA7zVJoT0Z4pw=="

	// 2. Base64-decode the string to get the raw ciphertext bytes
	cipherText, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		t.Fatalf("Failed to decode Base64 string: %v", err)
	}

	// 3. Pass the raw ciphertext bytes to the decryption function
	decrypt, err := encryptor.AesDecrypt([]byte("G2HX+F+411lK/O8c"), cipherText)
	if err != nil {
		t.Errorf("Failed to decrypt data: %v", err)
	}
	fmt.Printf("%s\n", string(decrypt))
}

func DBInit() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}
}

func TestSetSetting(t *testing.T) {
	DBInit()
	st := &settings.SettingItem{
		Group:     "account",
		Key:       "aeskey",
		Name:      "隐私信息对称加密key",
		Value:     "G2HX+F+411lK/O8c",
		ValueType: settings.ValueSecret,
	}
	if err := dbtools.Core().Create(&st).Error; err != nil {
		panic(err)
	}
}
