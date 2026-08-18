package types

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/insmtx/corekg/pkgs/utils/encryptor"
	uuid "github.com/satori/go.uuid"
	"github.com/ygpkg/yg-go/logs"
)

// Password 密码
type Password string

var secretKey = []byte("q84n7hz7k4b4tcb0")

const (
	encryptorPrefix = "enc:"
)

func (p Password) ToStringPtr() *string {
	str := string(p)
	return &str
}

type Secret string

// Enc 加密
func (s Secret) Enc(ctx context.Context) Secret {
	if strings.HasPrefix(string(s), encryptorPrefix) {
		return s
	}
	enc, err := encryptor.BlowfishEncryptToBase58(secretKey, []byte(s))
	if err != nil {
		logs.ErrorContextf(ctx, "encryptor.BlowfishEncryptToBase58(%s) error: %v", string(s), err)
		return s
	}
	return Secret(encryptorPrefix + enc)
}

// Dec 解密
func (s Secret) Dec(ctx context.Context) Secret {
	if !strings.HasPrefix(string(s), encryptorPrefix) {
		return s
	}
	enc := strings.TrimPrefix(string(s), encryptorPrefix)
	dec, err := encryptor.BlowfishDecryptFromBase58(secretKey, enc)
	if err != nil {
		logs.ErrorContextf(ctx, "encryptor.BlowfishDecryptFromBase58(%s) error: %v", enc, err)
		return s
	}
	return Secret(dec)
}

// SafeID 是一个安全的ID
type SafeID uint

// Enc 加密的ID
func (id SafeID) Enc(ctx context.Context) string {
	encStr, err := enc(big.NewInt(int64(id)).Bytes())
	if err != nil {
		logs.ErrorContextf(ctx, "enc(%v) error: %v", id, err)
		return ""
	}
	return encStr
}
func (id *SafeID) Dec(ctx context.Context, idstr string) {
	decStr, err := dec(idstr)
	if err != nil {
		logs.ErrorContextf(ctx, "dec(%s) error: %v", idstr, err)
		return
	}
	a := new(big.Int)
	decID := a.SetBytes(decStr).Int64()

	*id = SafeID(decID)
}

// MarshalJSON .
func (id SafeID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.Enc(context.TODO()))
}

// UnmarshalJSON .
func (id *SafeID) UnmarshalJSON(data []byte) error {
	var idstr string
	err := json.Unmarshal(data, &idstr)
	if err != nil {
		return err
	}
	decStr, err := dec(idstr)
	if err != nil {
		logs.ErrorContextf(context.TODO(), "dec(%s) error: %v", idstr, err)
		return err
	}
	a := new(big.Int)
	decID := a.SetBytes(decStr).Int64()

	*id = SafeID(decID)
	return nil
}

// enc 加密
func enc(data []byte) (string, error) {
	enc, err := encryptor.BlowfishEncryptToBase58(secretKey, data)
	if err != nil {
		return enc, err
	}
	return enc, nil
}

func dec(str string) ([]byte, error) {
	dec, err := encryptor.BlowfishDecryptFromBase58(secretKey, str)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}

// GenerateID 生成ID
func GenerateID() string {
	return hex.EncodeToString(uuid.Must(uuid.NewV4(), nil).Bytes())
}
