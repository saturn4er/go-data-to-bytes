package d2b

func bytesToStr(bytes []byte) string {
	for key, value := range bytes {
		if value == '\u0000' {
			return string(bytes[:key])
		}
	}
	return string(bytes[:])
}
