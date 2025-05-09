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
		'“': "\x93", '”': "\x94", '‘': "\x91", '’': "\x92",
		'–': "\x96", '—': "\x97", '…': "\x85",
		'«': "\xAB", '»': "\xBB", '†': "\x86",
		'§': "\xA7", '¶': "\xB6",
		'À': "\xC0", 'È': "\xC8", 'Ì': "\xCC", 'Ò': "\xD2", 'Ù': "\xD9",
		'à': "\xE0", 'è': "\xE8", 'ì': "\xEC", 'ò': "\xF2", 'ù': "\xF9",
		'Á': "\xC1", 'É': "\xC9", 'Í': "\xCD", 'Ó': "\xD3", 'Ú': "\xDA", 'Ý': "\xDD",
		'á': "\xE1", 'é': "\xE9", 'í': "\xED", 'ó': "\xF3", 'ú': "\xFA", 'ý': "\xFD",
		'Â': "\xC2", 'Ê': "\xCA", 'Î': "\xCE", 'Ô': "\xD4", 'Û': "\xDB",
		'â': "\xE2", 'ê': "\xEA", 'î': "\xEE", 'ô': "\xF4", 'û': "\xFB",
		'Ã': "\xC3", 'Ñ': "\xD1", 'Õ': "\xD5",
		'ã': "\xE3", 'ñ': "\xF1", 'õ': "\xF5",
		'Ä': "\xC4", 'Ë': "\xCB", 'Ï': "\xCF", 'Ö': "\xD6", 'Ü': "\xDC", 'Ÿ': "\x9F",
		'ä': "\xE4", 'ë': "\xEB", 'ï': "\xEF", 'ö': "\xF6", 'ü': "\xFC", 'ÿ': "\xFF",
		'Æ': "\xC6", 'Œ': "\x8C",
		'æ': "\xE6", 'œ': "\x9C",
		'Ç':    "\xC7",
		'ç':    "\xE7",
		0x00A0: "\xA0", // NBSP
	}
}
