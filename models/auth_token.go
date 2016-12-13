package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/elithrar/simple-scrypt"
	. "github.com/eywa/configs"
	"io"
	"time"
)

type AuthToken struct {
	Username    string    `json:"username"`
	TokenString string    `json:"token_string"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

var CheckSumInvalidErr = errors.New("invalid checksum.")
var AuthTokenExpiredErr = errors.New("auth token expired.")

type checksumedToken struct {
	Token    string `json:"token"`
	Checksum string `json:"checksum"`
}

func NewAuthToken(u, p string) (*AuthToken, error) {
	asBytes, err := scrypt.GenerateFromPassword([]byte(p), scrypt.DefaultParams)
	if err != nil {
		return nil, err
	}

	return &AuthToken{
		Username:    u,
		TokenString: base64.URLEncoding.EncodeToString(asBytes),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(Config().Security.Dashboard.TokenExpiry.Duration),
	}, nil
}

func (t *AuthToken) Encrypt() (string, error) {
	asBytes, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	asString := base64.URLEncoding.EncodeToString(asBytes)
	h := md5.New()
	io.WriteString(h, asString)
	cs := base64.URLEncoding.EncodeToString(h.Sum(nil))
	tk := &checksumedToken{
		Token:    asString,
		Checksum: cs,
	}
	asBytes, err = json.Marshal(tk)
	if err != nil {
		return "", err
	}
	asBytes, err = aesEncrypt(asBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(asBytes), nil
}

func DecryptAuthToken(str string) (*AuthToken, error) {
	asBytes, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	asBytes, err = aesDecrypt(asBytes)
	if err != nil {
		return nil, err
	}

	tk := &checksumedToken{}
	err = json.Unmarshal(asBytes, tk)
	if err != nil {
		return nil, err
	}

	h := md5.New()
	io.WriteString(h, tk.Token)
	cs := base64.URLEncoding.EncodeToString(h.Sum(nil))
	if cs != tk.Checksum {
		return nil, CheckSumInvalidErr
	}

	asBytes, err = base64.URLEncoding.DecodeString(tk.Token)
	if err != nil {
		return nil, err
	}

	t := &AuthToken{}
	err = json.Unmarshal(asBytes, t)
	if err != nil {
		return nil, err
	}

	if t.ExpiresAt.Before(time.Now()) {
		return nil, AuthTokenExpiredErr
	}

	return t, nil
}

func aesEncrypt(b []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(Config().Security.Dashboard.AES.KEY))
	if err != nil {
		return nil, err
	}
	iv := []byte(Config().Security.Dashboard.AES.IV)[:aes.BlockSize]

	encrypter := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(b))
	encrypter.XORKeyStream(encrypted, b)

	return encrypted, nil
}

func aesDecrypt(b []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(Config().Security.Dashboard.AES.KEY))
	if err != nil {
		return nil, err
	}
	iv := []byte(Config().Security.Dashboard.AES.IV)[:aes.BlockSize]

	decrypter := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(b))
	decrypter.XORKeyStream(decrypted, b)

	return decrypted, nil
}
