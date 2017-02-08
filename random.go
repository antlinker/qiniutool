package main

import (
	"math/rand"
	"time"
)

var _zfc = []byte("0123456789abcdefghijklmnopqrstuvwxyz")
var _zfclen = len(_zfc)
var _zfcrand = rand.New(rand.NewSource(time.Now().Unix()))

// CreateRandomString 创建随机字符串包含0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ中的字母数字
func CreateRandomString(size int) string {

	out := make([]byte, 0, size)
	for i := 0; i < size; i++ {
		out = append(out, _zfc[_zfcrand.Int()%_zfclen])
	}
	return string(out)
}
