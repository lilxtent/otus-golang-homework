package cursor

import (
	"unicode"
)

type StateManager struct {
	cursorState *State
}

func NewStateManager(cursorState *State) *StateManager {
	return &StateManager{
		cursorState: cursorState,
	}
}

func (cursorStateManager *StateManager) Apply(runeElement rune) {
	switch {
	case runeElement == '\\':
		applyEscapeRune(cursorStateManager.cursorState)

	case unicode.IsDigit(runeElement):
		applyDigitRune(cursorStateManager.cursorState, runeElement)

	default:
		cursorStateManager.cursorState.setSequence(runeElement)
	}
}

func (cursorStateManager *StateManager) SetRepeatTimes(times int) {
	cursorStateManager.cursorState.setRepeatTimes(times)
}

func applyEscapeRune(cursorState *State) {
	if cursorState.Escaped() {
		cursorState.setSequence('\\')
	} else {
		cursorState.Escape()
	}
}

func applyDigitRune(cursorState *State, runeElement rune) {
	if cursorState.Escaped() {
		cursorState.setSequence(runeElement)
	} else {
		cursorState.setRepeatTimes(int(runeElement - '0'))
	}
}
