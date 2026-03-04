// internal/config/flags.go
// Парсинг CLI флагов: метод ParseFlags overrides Config.

package config

import (
	"flag"
)

func ParseFlags(cfg *Config) {
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

	cfg.ServerMode = *serverMode
	cfg.Addr = *addr
	if cfg.Addr == "" {
		cfg.Addr = cfg.ServerAddr
	}
	cfg.Addrs = *addrsStr
	if cfg.Addrs == "" {
		cfg.Addrs = cfg.DefaultAddrs
	}
	cfg.GrepOptions = GrepOptions{
		Fixed:      *fixed,
		IgnoreCase: *ignoreCase,
		Invert:     *invert,
		LineNum:    *lineNum,
		After:      *after,
		Before:     *before,
		Context:    *context,
		Count:      *count,
	}
	cfg.Args = flag.Args()
}
