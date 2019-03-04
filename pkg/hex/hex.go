package hex

import (
	"bytes"
	"fmt"
	"strings"
)

func InsertNth(s string, n int) string {
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune('-')
		}
	}
	return buffer.String()
}

func HexToStream(str string) string {
	if len(str)%2 != 0 {
		fmt.Println("hex should be even")
		return ""
	}
	str = InsertNth(str, 2)
	strs := strings.Split(str, "-")
	for i, s := range strs {
		strs[i] = fmt.Sprintf("0x%s,", s)
	}
	return fmt.Sprintln(strs)
}

func HexRotation(str string) string {
	if len(str)%2 != 0 {
		fmt.Println("hex should be even")
		return ""
	}
	str = InsertNth(str, 2)
	strs := strings.Split(str, "-")
	for i, s := range strs {
		strs[i] = fmt.Sprintf("0x%s,", s)
	}
	for i, j := 0, len(strs)-1; i < j; i, j = i+1, j-1 {
		strs[i], strs[j] = strs[j], strs[i]
	}

	return fmt.Sprintln(strs)
}
