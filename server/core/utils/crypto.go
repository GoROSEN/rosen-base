package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

func GetPass(passMd5 string) string {
	h := sha256.New()
	h.Write([]byte(passMd5))
	return hex.EncodeToString(h.Sum(nil))
}

func GetSignature(params, appsecret string) string {

	// 算法：signature = md5(md5(params...)+APPSECRET)
	h := md5.New()
	h.Write([]byte(params))
	s := hex.EncodeToString(h.Sum(nil)) + appsecret
	h = md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
