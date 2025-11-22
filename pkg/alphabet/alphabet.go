// Package alphabet provides various rune slices for the English alphabet.
package alphabet

// LowerCaseRunes contains all lowercase runes from the English alphabet.
var LowerCaseRunes = []rune{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
}

// UpperCaseRunes contains all uppercase runes from the English alphabet.
var UpperCaseRunes = []rune{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}

// AllRunes contains both lowercase and uppercase runes from the English
// alphabet.
var AllRunes = append(LowerCaseRunes, UpperCaseRunes...)
