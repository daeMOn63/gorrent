package actions

// Actions
const (
	AnnounceID ID = 0x1
)

// ID define type for holding action ids
type ID uint8

// Action interface describe methodes needed for any actions
type Action interface {
	ID() ID
}

// DummyAction is a configurable action
type DummyAction struct {
	IDVar ID
}

// ID returns IDVar value
func (d *DummyAction) ID() ID {
	return d.IDVar
}
