package license

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/insmtx/corekg/pkgs/platform/admintype"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func TestChecker(t *testing.T) {
	rawLicense := `Mlmd6yrfBMGTZa4pKyjMc2hQMP6Q/HGDo3gvF0usAc6ywa6I81UzHYIzU+cYlu1m/kxgC2JNx/BBavLsQW0SnEe/rMEOQNVB7IuITOsQgD1T6vyxD8yjaUYQRXjGEGR7Y6siUvDHcs5cHpihhkMRF4b/bRGckLVu83pICKOj/+K/GiejVlIL4TmbvTQ8xsanHtbG8A/4fnoTJdMCarLvvKKpqoJBlACNx4vOdyoEIbZO7BAMenZMCHZZTQ+SPXRliAqUbmWHCBOT8i0+EgMZlFyTjMaXnJOfw7CS/QqDvyp0ZXp+P0c9mndVtXO6RACD98m5hpmP+UsFcP3ScDvXgw==.eyJpZCI6MCwic2VyaWFsIjoiZTE4ZTIyNTYtZGVlYy00OWI4LWE3NzEtOTk1ZTg0ZDU2OTNjIiwiZW52Ijoia3ViZXJuZXRlcyIsInVpZCI6ImM1MDBkYmI3LWYyOWUtNDcyZC04NTZiLTM1MDgyYjEyMGZjYSIsInN1YmplY3QiOiJoM2MiLCJpc3N1ZXIiOiJ5eWd1IiwiZXhwaXJlZF9hdCI6IjIwMjUtMTAtMDNUMjE6Mjg6MDUuNjgyMjc0NDY1KzA4OjAwIiwic2VlZCI6ImJiNDU3NmNkLTkwYjYtNDRhMi04MzVmLTI2NTA5YTdjNDZiMCJ9`
	key, err := license.ParsePublicKey(`
-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA2yzBf3x0puJOadJpnvyqNR0RCRMC5TSUO1EnY5SjDjtqHbap2Lkn
96kkvcyT2sEY/Wr4/KwhkBRET5uZnvzEOQYiyQJIJbVtVQ+jjuUMiOih2PppWt34
9ypLjhcbzAwanbUMM2kfsI8yPBem0fm6ewgKhJVOQWqG8p13+Jnr39a42vKh9M3Q
IndbNOSEvdX77L+73n98UlIpgrkzj0383jHCXy3hTL61ZOjk8rqhGjTorq/LaQQt
Wyc1h8nyKDAnm6RX40BzzP+TiaZyW6sgmK0ugUcMSQMy4S542DgG/VnAayYLMCl3
WMfu/sTo2IfTqfZrTDCpA9+5HcAb4ordHwIDAQAB
-----END RSA PUBLIC KEY-----`)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.TODO()
	parts := strings.Split(rawLicense, ".")
	if len(parts) != 2 {
		logs.ErrorContextf(ctx, "Invalid license format: %s", rawLicense)
		return
	}
	signature, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		logs.ErrorContextf(ctx, "invalid base64 for signature: %v", err)
		return
	}
	jsonData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		logs.ErrorContextf(ctx, "invalid base64 for data: %v", err)
		return
	}

	hashed := sha256.Sum256(jsonData)
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], signature); err != nil {
		logs.ErrorContextf(ctx, "invalid signature: %v", err)
		return
	}

	var meta admintype.Meta
	if err := json.Unmarshal(jsonData, &meta); err != nil {
		logs.ErrorContextf(ctx, "failed to unmarshal license metadata: %v", err)
		return
	}
	return
}

//test for hash chains
//| 10 | 2025-09-04 14:51:08.539 | 29fdf8d2e7eefc7da075e078a538cce75d5b3fe64a5e84a6f25b1c6fd770cf92 | 01a9fe6bea3cd0b71d79b76d5005803a51c7978d1027e82a35fa684b784e62c4
//|     1 | License is valid                      | 2025-09-04 14:51:08.539 | 2025-09-04 14:51:08.539 | NULL                    |
//| 11 | 2025-09-05 12:00:00.284 | 29fdf8d2e7eefc7da075e078a538cce75d5b3fe64a5e84a6f25b1c6fd770cf92 | ec7970fdbcc2d52c5a30a887d50407d241a29414e309f0fb6b5a23000e5f16cb
//|    -1 | license hash chain tampered           | 2025-09-05 12:00:00.284 | 2025-09-05 12:00:00.284 | NULL

func TestHashChains(t *testing.T) {
	ctx := context.TODO()
	ss := strings.Split("xUcr+kgW8PMLg7maktFkzx6K7GsMmfy8uXh3uX9ofOE++uZV2P06e5aHn4r0tL9O3Qy+ffRraacos7G5JwP5a4mjOt6AhrWboHLmNIECjzUl+Ew0e/APQNuMRFyG5Z3FkGdxGmYy24rG6nrnCdKTap+o2LmzkG1tHsq+obhIszdYnS5zhpTUX9n5NsRJkBvxyQbKmvtZ4zpj0F6tb8JiqHgZnTGAM/TI+hFQ+mGCq/C27jIagP1BsO0KOl6hbpuwhzrgztmweOjkAZ56c/1Gusp3vtBW5dZIUvXsKJhzR7jurrDBlWrsg3mESVsV6p+AQqe5mMlq38Dr33e3SSJLSA==.eyJpZCI6MCwic2VyaWFsIjoiZjQyYmNjYjctYzBmOC00MWIwLTk0ZmItYzg0NTAxMGNkN2MxIiwiZW52Ijoia3ViZXJuZXRlcyIsInVpZCI6ImM1MDBkYmI3LWYyOWUtNDcyZC04NTZiLTM1MDgyYjEyMGZjYSIsInN1YmplY3QiOiJoM2MiLCJpc3N1ZXIiOiJ5eWd1IiwiZXhwaXJlZF9hdCI6IjIwMjUtMTAtMDRUMTI6MDI6NDkuODU3NTU5MzM3KzA4OjAwIiwic2VlZCI6IjcxYzI1NDFhLWE1Y2MtNDhhYy1hYTdkLTUzY2IzOGVjN2RjMyJ9", ".")
	jsonData, err := base64.StdEncoding.DecodeString(ss[1])
	if err != nil {
		logs.ErrorContextf(ctx, "invalid base64 for data: %v", err)
		t.Fatal(err)
	}

	var meta admintype.Meta
	if err := json.Unmarshal(jsonData, &meta); err != nil {
		logs.ErrorContextf(ctx, "failed to unmarshal license metadata: %v", err)
		t.Fatal(err)
	}
}

func TestHashSum(t *testing.T) {
	uid := "c500dbb7-f29e-472d-856b-35082b120fca"
	seed := "71c2541a-a5cc-48ac-aa7d-53cb38ec7dc3"
	hmacKey := sha256.Sum256([]byte(uid + seed))
	data := []byte("2025-09-08 14:59:13.000" + "78a2f7bdee47ed77d147f8ba859094baf30ea75922951762ffab106d315dfd55")
	//bc7390d824101e92051b809f1a5a3d4ca74a956a66fddf24451a847df2c345b3	009
	//bc1a2b80cd04954715ae48a1ac23483777d0ca557ec0b1432c602537727190f8	000

	mac := hmac.New(sha256.New, hmacKey[:])
	mac.Write(data)
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	fmt.Println(expectedHash)
}

func InitDB() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}
}
