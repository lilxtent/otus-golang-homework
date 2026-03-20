package hw02unpackstring

type CursorState struct {
	LastLetter   *rune
	RepeatTimes  int
	EscapeNext   bool
	ReadyToWrite bool
}

func NewCursorState() *CursorState {
	return &CursorState{
		LastLetter:   nil,
		RepeatTimes:  -1,
		EscapeNext:   false,
		ReadyToWrite: false,
	}
}

func (cursorState *CursorState) Reset() {
	cursorState.LastLetter = nil
	cursorState.RepeatTimes = -1
	cursorState.EscapeNext = false
	cursorState.ReadyToWrite = false
}
