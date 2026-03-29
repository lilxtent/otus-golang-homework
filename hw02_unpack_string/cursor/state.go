package cursor

type State struct {
	sequence    *[]rune
	repeatTimes *int
	escape      *bool
}

func NewState() *State {
	return &State{
		sequence:    nil,
		repeatTimes: nil,
		escape:      nil,
	}
}

func (state *State) setSequence(runes ...rune) {
	state.sequence = &runes
	state.escape = nil
}

func (state *State) GetSequence() *[]rune {
	return state.sequence
}

func (state *State) IsSequenceSpecified() bool {
	return state.sequence != nil
}

func (state *State) Reset() {
	state.sequence = nil
	state.repeatTimes = nil
	state.escape = nil
}

func (state *State) ReadyToFlush() bool {
	return state.sequence != nil && state.repeatTimes != nil
}

func (state *State) Escaped() bool {
	return state.escape != nil && *state.escape
}

func (state *State) Escape() {
	escape := true
	state.escape = &escape
}

func (state *State) setRepeatTimes(repeatTimes int) {
	state.repeatTimes = &repeatTimes
}

func (state *State) IsRepeatTimesSpecified() bool {
	return state.repeatTimes != nil
}

func (state *State) GetRepeatTimes() int {
	if state.repeatTimes == nil {
		return 1
	}

	return *state.repeatTimes
}
