package structscomposition1

type server struct {
	port int
}

// NewServer constructs server safely and prevents direct construction outside package.
func NewServer(port int) *server {
	if port <= 0 || port > 65535 {
		return nil
	}
	return &server{port: port}
}

func (s *server) Port() int {
	if s == nil {
		return 0
	}
	return s.port
}	