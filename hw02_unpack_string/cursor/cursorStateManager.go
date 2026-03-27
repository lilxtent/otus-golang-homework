package cursor

import (
	"unicode"
)

type CursorStateManager struct {
	cursorState *CursorState
}

func NewCursorStateManager(cursorState *CursorState) *CursorStateManager {
	return &CursorStateManager{
		cursorState: cursorState,
	}
}

func (cursorStateManager *CursorStateManager) Apply(runeElement rune) error {
	var err error = nil

	if runeElement == '\\' {
		err = applyEscapeRune(cursorStateManager.cursorState)
	} else if unicode.IsDigit(runeElement) {
		err = applyDigitRune(cursorStateManager.cursorState, runeElement)
	} else {
		cursorStateManager.cursorState.setSequence(runeElement)
	}

	return err
}

func (cursorStateManager *CursorStateManager) SetRepeatTimes(times int) {
	cursorStateManager.cursorState.setRepeatTimes(times)
}

func applyEscapeRune(cursorState *CursorState) error {
	if cursorState.Escaped() {
		cursorState.setSequence('\\')
	} else {
		cursorState.Escape()
	}

	return nil
}

func applyDigitRune(cursorState *CursorState, runeElement rune) error {
	if cursorState.Escaped() {
		cursorState.setSequence(runeElement)
	} else {
		cursorState.setRepeatTimes(int(runeElement - '0'))
	}

	return nil
}
