package actions

import "errors"

var (
	// ErrNoHandler is returned when no handler can be found on the router with this
	// action ID.
	ErrNoHandler = errors.New("no handler for this action")
)

// Handler is the interface for all action handlers
type Handler interface {
	Handle(action Action) ([]byte, error)
}

// Router is a Handler where sub Handlers can be registered
type Router interface {
	Handler
	Register(ID, Handler)
}

// Router defines an action Router, where action handlers
// can be registered for each action IDs
type router struct {
	routes map[ID]Handler
}

var _ Router = &router{}
var _ Router = &DummyRouter{}
var _ Handler = &DummyHandler{}

// NewRouter create a new action router
func NewRouter() Router {
	return &router{
		routes: make(map[ID]Handler),
	}
}

// Register add a new handler for given action ID
func (r *router) Register(id ID, handler Handler) {
	r.routes[id] = handler
}

// Handle calls the associated action Handler from the given action's ID
func (r *router) Handle(action Action) ([]byte, error) {
	var handler Handler
	var ok bool

	if handler, ok = r.routes[action.ID()]; !ok {
		return nil, ErrNoHandler
	}

	return handler.Handle(action)
}

// DummyRouter provides a configurable Router
type DummyRouter struct {
	HandleFunc   func(action Action) ([]byte, error)
	RegisterFunc func(ID, Handler)
}

// Handle calls HandleFunc
func (d *DummyRouter) Handle(action Action) ([]byte, error) {
	return d.HandleFunc(action)
}

// Register calls RegisterFunc
func (d *DummyRouter) Register(id ID, handler Handler) {
	d.RegisterFunc(id, handler)
}

// DummyHandler provides a configurable Handler
type DummyHandler struct {
	HandleFunc func(action Action) ([]byte, error)
}

// Handle calls HandleFunc
func (d *DummyHandler) Handle(action Action) ([]byte, error) {
	return d.HandleFunc(action)
}
