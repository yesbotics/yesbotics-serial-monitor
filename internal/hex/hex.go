package hex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func GetHexString(data string) string {
	var hexBuilder strings.Builder

	for i, char := range data {
		if i > 0 {
			hexBuilder.WriteString(" ")
		}
		hexBuilder.WriteString(fmt.Sprintf("%02X", char))
	}

	return hexBuilder.String()
}

func ReplaceHexValuesToLatin1(message string) string {
	re := regexp.MustCompile(`#([0-9a-fA-F]{2})`)
	matches := re.FindAllStringSubmatchIndex(message, -1)

	var newString string
	lastIndex := 0
	for _, match := range matches {
		// Add the non-matching part
		newString += message[lastIndex:match[0]]

		// Convert hex to Latin-1 character
		hexValue := message[match[2]:match[3]]
		number, _ := strconv.ParseInt(hexValue, 16, 32)
		newString += string(rune(number))

		lastIndex = match[1]
	}

	// Add the remaining part of the string
	newString += message[lastIndex:]

	return newString
}
