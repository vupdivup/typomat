// Package keystroke provides structures and methods to represent and manage
// individual keystrokes in a typing session.
package keystroke

// Status represents the status of a keystroke.
type Status int

const (
	// StatusUnentered indicates the keystroke has not been entered yet.
	StatusUnentered Status = iota

	// StatusCorrect indicates the keystroke was entered correctly.
	StatusCorrect

	// StatusIncorrect indicates the keystroke was entered incorrectly.
	StatusIncorrect

	// StatusDeleted indicates the keystroke was deleted.
	StatusDeleted

	// StatusAmended indicates the keystroke was initially incorrect but later
	// corrected.
	StatusAmended
)

// Keystroke represents a single keystroke in the typing session.
type Keystroke struct {
	want         rune
	got          rune
	wasIncorrect bool
	status       Status
}

// New creates a new Keystroke with the expected rune.
func New(want rune) *Keystroke {
	return &Keystroke{
		want:   want,
		status: StatusUnentered,
	}
}

// Record records the user's input rune and updates the keystroke status.
func (k *Keystroke) Record(got rune) {
	k.got = got

	if got == k.want {
		if k.wasIncorrect {
			k.status = StatusAmended
		} else {
			k.status = StatusCorrect
		}
	} else {
		k.status = StatusIncorrect
		k.wasIncorrect = true
	}
}

// Delete marks the keystroke as deleted.
func (k *Keystroke) Delete() {
	k.status = StatusDeleted
}

// Status returns the current status of the keystroke.
func (k *Keystroke) Status() Status {
	return k.status
}
