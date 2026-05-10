package structscomposition1

import "testing"

func TestNewServerValidPort(t *testing.T) {
	s := NewServer(8080)
	if s == nil {
		t.Fatal("expected non-nil server for valid port")
	}

	if got := s.Port(); got != 8080 {
		t.Fatalf("expected port 8080, got %d", got)
	}
}

func TestNewServerInvalidPorts(t *testing.T) {
	tests := []int{0, -1, 70000}

	for _, port := range tests {
		port := port
		t.Run("invalid_port", func(t *testing.T) {
			s := NewServer(port)
			if s != nil {
				t.Fatalf("expected nil for invalid port %d, got %#v", port, s)
			}
		})
	}
}

func TestNilServerPort(t *testing.T) {
	var s *server
	if got := s.Port(); got != 0 {
		t.Fatalf("expected 0 for nil server, got %d", got)
	}
}
