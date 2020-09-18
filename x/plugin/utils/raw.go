package utils

import "fmt"

func HexString(input []byte) (output string) {
	return fmt.Sprintf("%x", input)
}
