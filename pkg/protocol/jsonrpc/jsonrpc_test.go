package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestDecodeMessage(t *testing.T) {
	t.Run("request", func(t *testing.T) {
		msg, err := DecodeMessage([]byte(`{"jsonrpc":"2.0","method":"ping","params":{},"id":"1"}`))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := msg.(Request); !ok {
			t.Fatalf("expected Request, got %T", msg)
		}
	})

	t.Run("response", func(t *testing.T) {
		msg, err := DecodeMessage([]byte(`{"jsonrpc":"2.0","result":"pong","id":"1"}`))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := msg.(Response); !ok {
			t.Fatalf("expected Response, got %T", msg)
		}
	})

	t.Run("event", func(t *testing.T) {
		msg, err := DecodeMessage([]byte(`{"jsonrpc":"2.0","method":"videoInputState","params":{"state":"ok"}}`))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := msg.(Event); !ok {
			t.Fatalf("expected Event, got %T", msg)
		}
	})

	t.Run("error_response", func(t *testing.T) {
		msg, err := DecodeMessage([]byte(`{"jsonrpc":"2.0","error":{"code":-32601,"message":"method not found"},"id":"2"}`))
		if err != nil {
			t.Fatal(err)
		}
		resp, ok := msg.(Response)
		if !ok {
			t.Fatalf("expected Response, got %T", msg)
		}
		if resp.Error == nil {
			t.Fatal("expected error field")
		}
		if resp.Error.Code != -32601 {
			t.Errorf("error code = %d, want -32601", resp.Error.Code)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := DecodeMessage([]byte(`{"foo":"bar"}`))
		if err != ErrUnknownMessage {
			t.Fatalf("expected ErrUnknownMessage, got %v", err)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := DecodeMessage([]byte(`not json`))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestNewRequest(t *testing.T) {
	req := NewRequest("ping", nil, 1)
	if req.JSONRPC != Version {
		t.Errorf("version = %q, want %q", req.JSONRPC, Version)
	}
	if req.Method != "ping" {
		t.Errorf("method = %q, want ping", req.Method)
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := DecodeMessage(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := msg.(Request); !ok {
		t.Fatalf("round-trip: expected Request, got %T", msg)
	}
}

func TestNewResponse(t *testing.T) {
	resp := NewResponse("abc", map[string]string{"key": "val"})
	if resp.JSONRPC != Version {
		t.Errorf("version = %q", resp.JSONRPC)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := DecodeMessage(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := msg.(Response); !ok {
		t.Fatalf("expected Response, got %T", msg)
	}
}

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse(42, -32600, "invalid request", nil)
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != -32600 {
		t.Errorf("code = %d, want -32600", resp.Error.Code)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := DecodeMessage(data)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := msg.(Response)
	if !ok {
		t.Fatalf("expected Response, got %T", msg)
	}
	if r.Error == nil || r.Error.Code != -32600 {
		t.Errorf("round-trip error mismatch")
	}
}

func TestNewEvent(t *testing.T) {
	evt := NewEvent("stateUpdate", map[string]bool{"connected": true})
	if evt.JSONRPC != Version {
		t.Errorf("version = %q", evt.JSONRPC)
	}
	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := DecodeMessage(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := msg.(Event); !ok {
		t.Fatalf("expected Event, got %T", msg)
	}
}

func TestMustVersion(t *testing.T) {
	if err := MustVersion("2.0"); err != nil {
		t.Errorf("MustVersion(2.0) = %v, want nil", err)
	}
	if err := MustVersion("1.0"); err == nil {
		t.Error("MustVersion(1.0) expected error")
	}
}

func TestCompact(t *testing.T) {
	input := `{  "key" :  "value"  }`
	got := Compact([]byte(input))
	want := `{"key":"value"}`
	if got != want {
		t.Errorf("Compact = %q, want %q", got, want)
	}

	got = Compact([]byte(`not json`))
	if got != "not json" {
		t.Errorf("Compact(invalid) = %q, want passthrough", got)
	}
}
