package encryptor

import (
	"testing"
)

func TestA(t *testing.T) {
	key := []byte("rEMyV4lXdhb1c9Js")
	defer func() {
		if err := recover(); err != nil {
			t.Logf("Recovered from panic: %v", err)
		}
	}()

	n, err := AesDecryptFromBase64(key, string("woh1Kt2u+ZPo82Yse81zpymwBpiG6lwyy923RtijJPWOAZfQuyDr1eEieNEHesIb"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Decrypted:,%d, %s", len(n), n)
}

func TestAesEncrypt(t *testing.T) {
	key := []byte("nYwlQsGauP5YMQYT")
	for _, v := range []string{
		"a",
		"aa",
		"aaa",
		"我",
		"我我",
		"我我我",
		"hello world",
	} {
		encrypted, err := AesEncrypt([]byte(key), []byte(v))
		if err != nil {
			t.Fatal(err)
		}

		decrypted, err := AesDecrypt([]byte(key), encrypted)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Encrypted: %s -> %s -> %s", v, encrypted, decrypted)
	}
}

func TestAesEncryptToBase64(t *testing.T) {
	key := []byte("nYwlQsGauP5YMQYT")
	for _, v := range []string{
		"a",
		"aa",
		"aaa",
		"我",
		"我我",
		"我我我",
		"hello world",
	} {
		encrypted, err := AesEncryptToBase64([]byte(key), []byte(v))
		if err != nil {
			t.Fatal(err)
		}

		decrypted, err := AesDecryptFromBase64([]byte(key), encrypted)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Encrypted: %s -> %s -> %s", v, encrypted, decrypted)
	}
}

func TestAesEncryptToBase58(t *testing.T) {
	key := []byte("nYwlQsGauP5YMQYT")
	for _, v := range []string{
		"a",
		"aa",
		"aaa",
		"我",
		"我我",
		"我我我",
		"hello world",
	} {
		encrypted, err := AesEncryptToBase58([]byte(key), []byte(v))
		if err != nil {
			t.Fatal(err)
		}

		decrypted, err := AesDecryptFromBase58([]byte(key), encrypted)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Encrypted: %s -> %s -> %s", v, encrypted, decrypted)
	}
}

func TestDecryptAesCBC(t *testing.T) {
	key := []byte("nYwlQsGauP5YMQYT")
	for _, v := range []string{
		"a",
		"aa",
		"aaa",
		"我",
		"我我",
		"我我我",
		"hello world",
	} {
		encrypted, err := EncryptAesCBC([]byte(key), []byte(v))
		if err != nil {
			t.Fatal(err)
		}

		decrypted, err := DecryptAesCBC([]byte(key), encrypted)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Encrypted: %s -> %s -> %s", v, encrypted, decrypted)
	}
}
