package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	resultString := strings.Builder{}
	cursorState := NewCursorState()
	index := 0

	if input == "" {
		return "", nil
	}

	if unicode.IsDigit(rune(input[0])) {
		return "", ErrInvalidString
	}

	for {
		if index >= len(input) {
			if cursorState.EscapeNext {
				return "", ErrInvalidString
			}

			stringToWrite := CreateString(*cursorState.LastLetter, 1)
			resultString.WriteString(stringToWrite)
			cursorState.Reset()

			break
		}

		if cursorState.ReadyToWrite {
			stringToWrite := CreateString(*cursorState.LastLetter, cursorState.RepeatTimes)
			resultString.WriteString(stringToWrite)
			cursorState.Reset()
		}

		currentSymbol := rune(input[index])

		if unicode.IsLetter(currentSymbol) || index >= len(input) {
			if cursorState.LastLetter != nil {
				cursorState.RepeatTimes = 1
				cursorState.ReadyToWrite = true

				continue
			}

			cursorState.LastLetter = &currentSymbol

			index++
			continue
		}

		if unicode.IsDigit(currentSymbol) {
			if cursorState.EscapeNext {
				cursorState.LastLetter = &currentSymbol
			} else {
				digit, atoiErr := strconv.Atoi(string(currentSymbol))

				if atoiErr != nil {
					return "", ErrInvalidString
				}

				cursorState.RepeatTimes = digit
				cursorState.ReadyToWrite = true
			}

			index++
			continue
		}

		if currentSymbol == '\\' {
			cursorState.EscapeNext = true

			index++
			continue
		}

		cursorState.ReadyToWrite = true
	}

	return resultString.String(), nil
}

func CreateString(symbol rune, repeat int) string {
	if repeat == 1 {
		return string(symbol)
	}

	return strings.Repeat(string(symbol), repeat)
}
