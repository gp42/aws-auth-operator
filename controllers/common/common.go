package common

import (
	"crypto/sha512"
	"encoding/base64"
	"sort"
	"strings"
)

func Sha512SumFromStringSorted(d string) string {
	// Sort lines to get consistent results independent of map order
	var sorted sort.StringSlice
	sorted = strings.Split(d, "\n")
	sorted.Sort()

	sum := sha512.Sum512([]byte(strings.Join(sorted, "\n")))
	return string(base64.URLEncoding.EncodeToString(sum[:]))
}
