package main

import (
	"bytes"
	"compress/flate"
	"io/ioutil"
	"testing"
)

func UnFlate(input []byte) []byte {
	rdata := bytes.NewReader(input)
	r := flate.NewReader(rdata)
	s, _ := ioutil.ReadAll(r)
	return s
}

func Test_parser(t *testing.T) {
	ioutil.WriteFile("fingers.json", UnFlate(recuLoadFinger("fingers/socket", true)), 0777)
}
