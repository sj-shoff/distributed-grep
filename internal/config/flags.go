package config

import (
	"flag"
)

type Flags struct {
	ServerMode bool
	Addr       string
	Addrs      string
	Fixed      bool
	IgnoreCase bool
	Invert     bool
	LineNum    bool
	After      int
	Before     int
	Context    int
	Count      bool
	Args       []string
}

func ParseFlags(cfg *Config) *Flags {
	serverMode := flag.Bool("server", false, "Run in server mode")
	addr := flag.String("addr", "", "Address to listen on in server mode")
	addrsStr := flag.String("addrs", "", "Comma-separated list of server addresses (client mode)")

	fixed := flag.Bool("F", false, "Interpret pattern as fixed string")
	ignoreCase := flag.Bool("i", false, "Ignore case")
	invert := flag.Bool("v", false, "Invert match")
	lineNum := flag.Bool("n", false, "Print line numbers")
	after := flag.Int("A", 0, "Print N lines after match")
	before := flag.Int("B", 0, "Print N lines before match")
	context := flag.Int("C", 0, "Print N lines of context")
	count := flag.Bool("c", false, "Print only count of matches")

	flag.Parse()

	return &Flags{
		ServerMode: *serverMode,
		Addr:       *addr,
		Addrs:      *addrsStr,
		Fixed:      *fixed,
		IgnoreCase: *ignoreCase,
		Invert:     *invert,
		LineNum:    *lineNum,
		After:      *after,
		Before:     *before,
		Context:    *context,
		Count:      *count,
		Args:       flag.Args(),
	}
}
