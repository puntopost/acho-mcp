package assets

import _ "embed"

//go:embed juan.ans
var juanANS []byte

func JuanANS() []byte {
	return juanANS
}
