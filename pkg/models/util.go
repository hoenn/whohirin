package models

func incSelectionWithWrap(i, inc, max int) int {
	return (i + inc) % max
}

func decSelectionWithWrap(i, dec, max int) int {
	if i-dec < 0 {
		return max - 1
	}
	return i - 1
}
