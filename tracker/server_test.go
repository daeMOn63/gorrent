package tracker

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/daeMOn63/gorrent/tracker/actions"
)

func getFreePort() int {
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.LocalAddr().(*net.UDPAddr).Port
}

var cfg = ServerConfig{
	Addr:     fmt.Sprintf(":%d", getFreePort()),
	Protocol: "udp",
}

var reader = &actions.DummyReader{}
var router = &actions.DummyRouter{}

func init() {
	s := NewServer(cfg, reader, router)
	go func() {
		s.Listen()
	}()

	time.Sleep(10 * time.Millisecond) // wait for the server to finish booting up
}

func TestServer(t *testing.T) {
	t.Run("server must read actions, dispatch them to the router and return the response", func(t *testing.T) {
		expectedPayload := []byte("abcd")
		expectedAction := &actions.DummyAction{
			IDVar: actions.ID(1),
		}
		expectedOutput := []byte("dcba")

		reader.ReadFunc = func(buf []byte) (actions.Action, error) {
			if reflect.DeepEqual(buf, expectedPayload) == false {
				t.Fatalf("Expected buf to be %v, got %v", expectedPayload, buf)
			}
			fmt.Println("read")
			return expectedAction, nil
		}

		router.HandleFunc = func(action actions.Action) ([]byte, error) {
			if reflect.DeepEqual(action, expectedAction) == false {
				t.Fatalf("Expected action to be %#v, got %#v", expectedAction, action)
			}
			fmt.Println("handle")

			return expectedOutput, nil
		}

		conn, err := net.Dial(cfg.Protocol, cfg.Addr)
		if err != nil {
			t.Fatal("could not connect to server: ", err)
		}
		defer conn.Close()

		_, err = conn.Write(expectedPayload)
		if err != nil {
			t.Fatal("could not write to server: ", err)
		}

		buf := make([]byte, 10)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal("could not read from server: ", err)
		}

		if reflect.DeepEqual(buf[:n], expectedOutput) == false {
			t.Fatalf("Expected output to be %v, got %v", expectedOutput, buf[:n])
		}
	})
}
