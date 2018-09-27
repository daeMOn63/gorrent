package actions

import (
	"errors"
	"reflect"
	"testing"
)

func TestRouter(t *testing.T) {
	t.Run("Router properly routes", func(t *testing.T) {
		expectedAction := &DummyAction{IDVar: 1}
		expectedOut := []byte("abc")
		expectedErr := errors.New("handlererr")

		r := NewRouter()

		r.Register(ID(1), &DummyHandler{
			HandleFunc: func(action Action) ([]byte, error) {
				if reflect.DeepEqual(action, expectedAction) == false {
					t.Fatalf("Expected action to be %#v, got %#v", expectedAction, action)
				}

				return expectedOut, expectedErr
			},
		})

		r.Register(ID(2), &DummyHandler{
			HandleFunc: func(action Action) ([]byte, error) {
				t.Fatalf("Call was not expected")

				return nil, nil
			},
		})

		out, err := r.Handle(expectedAction)
		if reflect.DeepEqual(out, expectedOut) == false {
			t.Fatalf("Expected out to be %v, got %v", expectedOut, out)
		}
		if err != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, err)
		}
	})

	t.Run("Router returns error on unknow actions", func(t *testing.T) {
		r := NewRouter()
		a := &DummyAction{IDVar: 1}

		out, err := r.Handle(a)
		if out != nil {
			t.Fatalf("Expected out to be nil, got %v", out)
		}

		if err != ErrNoHandler {
			t.Fatalf("Expected err to be %s, got %s", ErrNoHandler, err)
		}
	})
}

func TestDummyRouter(t *testing.T) {
	t.Run("DummyRouter.Handle calls HandleFunc", func(t *testing.T) {
		expectedAction := &DummyAction{
			IDVar: 1,
		}
		expectedOut := []byte("abc")
		expectedErr := errors.New("handlerr")

		d := &DummyRouter{
			HandleFunc: func(action Action) ([]byte, error) {
				if reflect.DeepEqual(action, expectedAction) == false {
					t.Fatalf("Expected action to be %#v, got %#v", expectedAction, action)
				}

				return expectedOut, expectedErr
			},
		}

		out, err := d.Handle(expectedAction)
		if err != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, err)
		}

		if reflect.DeepEqual(out, expectedOut) == false {
			t.Fatalf("Expected out to be %v, got %v", expectedOut, out)
		}
	})

	t.Run("DummyRouter.Register calls RegisterFunc", func(t *testing.T) {
		expectedID := ID(1)
		expectedHandler := &DummyHandler{}

		registerFuncCalled := false

		d := &DummyRouter{
			RegisterFunc: func(id ID, handler Handler) {
				if id != expectedID {
					t.Fatalf("Expected id to be %d, got %d", expectedID, id)
				}

				if reflect.DeepEqual(handler, expectedHandler) == false {
					t.Fatalf("Expected handler to be %#v, got %#v", expectedHandler, handler)
				}

				registerFuncCalled = true
			},
		}

		d.Register(expectedID, expectedHandler)

		if registerFuncCalled == false {
			t.Fatalf("Expected registerFunc to be called")
		}
	})
}
