package assets

import _ "embed"

//go:embed word_hans_utf8.txt
var LocationNameHansBytes []byte

//go:embed word_hant_utf8.txt
var LocationNameHantBytes []byte
