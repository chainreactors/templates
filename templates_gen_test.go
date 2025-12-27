package main

import (
	"bytes"
	"compress/flate"
	en "github.com/chainreactors/utils/encode"
	"io/ioutil"
	"testing"
)

func UnFlate(input []byte) []byte {
	rdata := bytes.NewReader(input)
	r := flate.NewReader(rdata)
	s, _ := ioutil.ReadAll(r)
	return s
}

func Decode(input string) []byte {
	b := en.Base64Decode(input)
	return UnFlate(b)
}

func Test_parser(t *testing.T) {
	ioutil.WriteFile("fingers.json", Decode(recuLoadFinger("fingers/socket")), 0777)
}
