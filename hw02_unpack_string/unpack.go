package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"

	"github.com/lilxtent/otus-golang-homework/hw02_unpack_string/cursor"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	return loopOverString(input)
}

func loopOverString(input string) (string, error) {
	stringBuilder := &strings.Builder{}
	cursorState := cursor.NewState()
	cursorStateManager := cursor.NewStateManager(cursorState)

	for _, runeElement := range input {
		if cursorState.IsRepeatTimesSpecified() && !cursorState.IsSequenceSpecified() {
			return "", ErrInvalidString
		}

		if isSequenceEnded(cursorState, runeElement) {
			cursorStateManager.SetRepeatTimes(1)
		}

		if cursorState.ReadyToFlush() {
			if err := flush(stringBuilder, cursorState); err != nil {
				return "", err
			}
		}

		cursorStateManager.Apply(runeElement)
	}

	if err := handleCursorStateAfterLoop(cursorStateManager, cursorState, stringBuilder); err != nil {
		return "", err
	}

	return stringBuilder.String(), nil
}

func flush(stringsBuilder *strings.Builder, cursorState *cursor.State) error {
	if cursorState == nil {
		return errors.New("cursorState state cannot be nil")
	}

	if !cursorState.IsSequenceSpecified() {
		return errors.New("cursorState.Sequence must be specified")
	}

	stringToWrite := getStringToWrite(cursorState)
	_, err := stringsBuilder.WriteString(*stringToWrite)
	cursorState.Reset()

	return err
}

func getStringToWrite(cursorState *cursor.State) *string {
	sequenceAsString := string(*cursorState.GetSequence())

	if cursorState.IsRepeatTimesSpecified() {
		repeatedString := strings.Repeat(sequenceAsString, cursorState.GetRepeatTimes())
		return &repeatedString
	}

	return &sequenceAsString
}

func isSequenceEnded(cursor *cursor.State, runeElement rune) bool {
	if cursor.IsSequenceSpecified() && runeElement == '\\' {
		return true
	}

	return cursor.IsSequenceSpecified() && !cursor.IsRepeatTimesSpecified() &&
		runeElement != '\\' && !unicode.IsDigit(runeElement)
}

func handleCursorStateAfterLoop(cursorStateManager *cursor.StateManager, cursorState *cursor.State,
	stringBuilder *strings.Builder,
) error {
	if cursorState.IsSequenceSpecified() && !cursorState.IsRepeatTimesSpecified() {
		cursorStateManager.SetRepeatTimes(1)
	}

	if cursorState.ReadyToFlush() {
		if err := flush(stringBuilder, cursorState); err != nil {
			return err
		}
	}

	if cursorState.IsSequenceSpecified() || cursorState.IsRepeatTimesSpecified() || cursorState.Escaped() {
		return ErrInvalidString
	}

	return nil
}
