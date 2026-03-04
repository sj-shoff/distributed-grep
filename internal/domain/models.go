package domain

type Match struct {
	RelLineNum int
	Text       string
}

type GlobalMatch struct {
	GlobalLineNum int
	Text          string
}

type GrepRequest struct {
	Pattern    string
	InputData  []string
	Fixed      bool
	IgnoreCase bool
	Invert     bool
}

type ChunkRequest struct {
	ChunkID    int
	Lines      []string
	Pattern    string
	Fixed      bool
	IgnoreCase bool
	Invert     bool
}

type ChunkResponse struct {
	ChunkID int
	Matches []Match
	Err     error
}
