package main

func rotateRunes(s string, shift int) string {
	// используем rune чтобы поймать битые байты и работать с норм символами
	r := []rune(s)
	runeCount := len(r)
	if runeCount < 2 || shift%runeCount == 0 {
		return string(r)
	}
	shift = ((shift % runeCount) + runeCount) % runeCount
	// отрезаем лишнее -> приводим к одному движению с положительным shift ->
	// и еще раз отрезали лишнее если строка больше сдвига

	resultStr := make([]rune, runeCount)

	copy(resultStr, r[shift:])
	copy(resultStr[runeCount-shift:], r[:shift])

	return string(resultStr)
}
