package main

import (
	"crypto/md5"
	"fmt"
	"io"
)

func M5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%032x", h.Sum(nil))
}

func Sign_str(sercet, timestemp string) string {
	var ret string
	ret = M5("n" + sercet + "e" + timestemp + "0")
	return ret
}
