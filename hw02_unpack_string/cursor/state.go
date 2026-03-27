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

func (State *State) setSequence(runes ...rune) {
	State.sequence = &runes
	State.escape = nil
}

func (State *State) GetSequence() *[]rune {
	return State.sequence
}

func (State *State) IsSequenceSpecified() bool {
	return State.sequence != nil
}

func (State *State) Reset() {
	State.sequence = nil
	State.repeatTimes = nil
	State.escape = nil
}

func (State *State) ReadyToFlush() bool {
	return State.sequence != nil && State.repeatTimes != nil
}

func (State *State) Escaped() bool {
	return State.escape != nil && *State.escape
}

func (State *State) Escape() {
	escape := true
	State.escape = &escape
}

func (State *State) setRepeatTimes(repeatTimes int) {
	State.repeatTimes = &repeatTimes
}

func (State *State) IsRepeatTimesSpecified() bool {
	return State.repeatTimes != nil
}

func (State *State) GetRepeatTimes() int {
	if State.repeatTimes == nil {
		return 1
	}

	return *State.repeatTimes
}
