// mappings.go
//
// Definitions for mapping between character sets.
package qris

func utf8ToNormalized(data string, normTable map[rune]string) string {
	var result string
	if normTable == nil {
		result = data
	} else {
		rs := []rune(data)
		for _, r := range rs { // check for runes to be replaced by a single rune
			if newRunes, replace := normTable[r]; replace {
				result += newRunes
			} else {
				result += string(r)
			}
		}
	}
	return result
}

func utf8ToAscii() map[rune]string {
	return map[rune]string{
		'“': `"`, '”': `"`, '‘': `'`, '’': `'`,
		'–': `-`, '—': `--`, '…': `...`,
		'«': `<<`, '»': `>>`, '†': ``,
		'§': "ss", '¶': "pp",
		'À': `A`, 'È': `E`, 'Ì': `I`, 'Ò': `O`, 'Ù': `U`,
		'à': `a`, 'è': `e`, 'ì': `i`, 'ò': `o`, 'ù': `u`,
		'Á': `A`, 'É': `E`, 'Í': `I`, 'Ó': `O`, 'Ú': `U`, 'Ý': `Y`,
		'á': `a`, 'é': `e`, 'í': `i`, 'ó': `o`, 'ú': `u`, 'ý': `y`,
		'Â': `A`, 'Ê': `E`, 'Î': `I`, 'Ô': `O`, 'Û': `U`,
		'â': `a`, 'ê': `e`, 'î': `i`, 'ô': `o`, 'û': `u`,
		'Ã': `A`, 'Ñ': `N`, 'Õ': `O`,
		'ã': `a`, 'ñ': `n`, 'õ': `o`,
		'Ä': `A`, 'Ë': `E`, 'Ï': `I`, 'Ö': `O`, 'Ü': `U`, 'Ÿ': `Y`,
		'ä': `a`, 'ë': `e`, 'ï': `i`, 'ö': `o`, 'ü': `u`, 'ÿ': `y`,
		'Æ': `ae`, 'Œ': `OE`,
		'æ': `ae`, 'œ': `oe`,
		'Ç':    `C`,
		'ç':    `c`,
		0x00A0: ` `, // NBSP
	}
}

func utf8ToAnsi() map[rune]string {
	return map[rune]string{
		'€': "\x80", '‚': "\x82", 'ƒ': "\x83", '„': "\x84", '…': "\x85",
		'†': "\x86", '‡': "\x87", 'ˆ': "\x88", '‰': "\x89", 'Š': "\x8A",
		'‹': "\x8B", 'Œ': "\x8C", 'Ž': "\x8E",
		'‘': "\x91", '’': "\x92", '“': "\x93", '”': "\x94", '•': "\x95",
		'–': "\x96", '—': "\x97", '˜': "\x98", '™': "\x99", 'š': "\x9A",
		'›': "\x9B", 'œ': "\x9C", 'ž': "\x9E", 'Ÿ': "\x9F", '¡': "\xA1",
		'¢': "\xA2", '£': "\xA3", '¤': "\xA4", '¥': "\xA5", '¦': "\xA6",
		'§': "\xA7", '¨': "\xA8", '©': "\xA9", 'ª': "\xAA", '«': "\xAB",
		'¬': "\xAC", '®': "\xAE", '¯': "\xAF", '°': "\xB0", '±': "\xB1",
		'²': "\xB2", '³': "\xB3", '´': "\xB4", 'µ': "\xB5", '¶': "\xB6",
		'·': "\xB7", '¸': "\xB8", '¹': "\xB9", 'º': "\xBA", '»': "\xBB",
		'¼': "\xBC", '½': "\xBD", '¾': "\xBE", '¿': "\xBF",
		'À': "\xC0", 'Á': "\xC1", 'Â': "\xC2", 'Ã': "\xC3", 'Ä': "\xC4", 'Å': "\xC5",
		'Æ': "\xC6", 'Ç': "\xC7",
		'È': "\xC8", 'É': "\xC9", 'Ê': "\xCA", 'Ë': "\xCB",
		'Ì': "\xCC", 'Í': "\xCD", 'Î': "\xCE", 'Ï': "\xCF",
		'Ð': "\xD0", 'Ñ': "\xD1",
		'Ò': "\xD2", 'Ó': "\xD3", 'Ô': "\xD4", 'Õ': "\xD5", 'Ö': "\xD6",
		'×': "\xD7", 'Ø': "\xD8",
		'Ù': "\xD9", 'Ú': "\xDA", 'Û': "\xDB", 'Ü': "\xDC",
		'Ý': "\xDD", 'Þ': "\xDE", 'ß': "\xDF",
		'à': "\xE0", 'á': "\xE1", 'â': "\xE2", 'ã': "\xE3", 'ä': "\xE4", 'å': "\xE5",
		'æ': "\xE6", 'ç': "\xE7",
		'è': "\xE8", 'é': "\xE9", 'ê': "\xEA", 'ë': "\xEB",
		'ì': "\xEC", 'í': "\xED", 'î': "\xEE", 'ï': "\xEF",
		'ð': "\xF0", 'ñ': "\xF1",
		'ò': "\xF2", 'ó': "\xF3", 'ô': "\xF4", 'õ': "\xF5", 'ö': "\xF6",
		'÷': "\xF7", 'ø': "\xF8",
		'ù': "\xF9", 'ú': "\xFA", 'û': "\xFB", 'ü': "\xFC",
		'ý': "\xFD", 'þ': "\xFE", 'ÿ': "\xFF",
		0x00A0: "\xA0", // NBSP : non-breaking space
		0x00AD: "\xAD", // SHY : soft hyphen
	}
}
