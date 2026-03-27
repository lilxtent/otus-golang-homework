package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	cursor := CursorState{}
	stringBuilder := strings.Builder{}

	for _, runeElement := range input {
		if IsSequenceEnded(&cursor, runeElement) {
			cursor.SetRepeatTimes(1)
		}

		if cursor.ReadyToFlush() {
			err := Flush(&stringBuilder, &cursor)
			if err != nil {
				return "", err
			}
		}

		if runeElement == '\\' {
			if cursor.Escaped() {
				cursor.SetSequence(runeElement)
			} else {
				cursor.Escape()
			}
		} else if unicode.IsDigit(runeElement) {
			if cursor.IsRepeatTimesSpecified() {
				return "", ErrInvalidString
			}

			if cursor.Escaped() {
				cursor.SetSequence(runeElement)
			} else {
				cursor.SetRepeatTimes(int(runeElement - '0'))
			}
		} else {
			cursor.SetSequence(runeElement)
		}
	}

	if cursor.IsSequenceSpecified() && !cursor.IsRepeatTimesSpecified() {
		cursor.SetRepeatTimes(1)
	}

	if cursor.ReadyToFlush() {
		err := Flush(&stringBuilder, &cursor)
		if err != nil {
			return "", err
		}
	}

	if cursor.IsSequenceSpecified() || cursor.IsRepeatTimesSpecified() || cursor.Escaped() {
		return "", ErrInvalidString
	}

	return stringBuilder.String(), nil
}

func Flush(stringsBuilder *strings.Builder, cursorState *CursorState) error {
	if cursorState == nil {
		return errors.New("cursorState state cannot be nil")
	}

	if !cursorState.IsSequenceSpecified() {
		return errors.New("cursorState.Sequence must be specified")
	}

	stringToWrite := GetStringToWrite(cursorState)
	_, err := stringsBuilder.WriteString(*stringToWrite)
	cursorState.Reset()

	return err
}

func GetStringToWrite(cursorState *CursorState) *string {
	sequenceAsString := string(*cursorState.GetSequence())

	if cursorState.IsRepeatTimesSpecified() {
		repeatedString := strings.Repeat(sequenceAsString, cursorState.GetRepeatTimes())
		return &repeatedString
	} else {
		return &sequenceAsString
	}
}

func IsSequenceEnded(cursor *CursorState, runeElement rune) bool {
	if cursor.IsSequenceSpecified() && runeElement == '\\' {
		return true
	}

	return cursor.IsSequenceSpecified() && !cursor.IsRepeatTimesSpecified() && runeElement != '\\' && !unicode.IsDigit(runeElement)
}
