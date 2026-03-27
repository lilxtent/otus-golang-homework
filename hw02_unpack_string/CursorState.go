package hw02unpackstring

type CursorState struct {
	sequence    *[]rune
	repeatTimes *int
	escape      *bool
}

func NewCursorState() *CursorState {
	return &CursorState{
		sequence:    nil,
		repeatTimes: nil,
		escape:      nil,
	}
}

func (cursorState *CursorState) SetSequence(runes ...rune) {
	cursorState.sequence = &runes
	cursorState.escape = nil
}

func (cursorState *CursorState) GetSequence() *[]rune {
	return cursorState.sequence
}

func (cursorState *CursorState) IsSequenceSpecified() bool {
	return cursorState.sequence != nil
}

func (cursorState *CursorState) Reset() {
	cursorState.sequence = nil
	cursorState.repeatTimes = nil
	cursorState.escape = nil
}

func (cursorState *CursorState) ReadyToFlush() bool {
	return cursorState.sequence != nil && cursorState.repeatTimes != nil
}

func (cursorState *CursorState) Escaped() bool {
	return cursorState.escape != nil && *cursorState.escape
}

func (cursorState *CursorState) Escape() {
	escape := true
	cursorState.escape = &escape
}

func (cursorState *CursorState) SetRepeatTimes(repeatTimes int) {
	cursorState.repeatTimes = &repeatTimes
}

func (cursorState *CursorState) IsRepeatTimesSpecified() bool {
	return cursorState.repeatTimes != nil
}

func (cursorState *CursorState) GetRepeatTimes() int {
	if cursorState.repeatTimes == nil {
		return 1
	}

	return *cursorState.repeatTimes
}
