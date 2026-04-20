package output

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// Mode selects the output renderer.
type Mode int

const (
	ModeAuto   Mode = iota // choose based on TTY
	ModeJSON               // force JSON
	ModePretty             // force pretty table
)

// Writer is the minimal contract every renderer satisfies.
// Concrete writers (*JSONWriter, *TableWriter) also expose
// per-command methods (e.g., WritePortfolio); callers typecheck
// when they need those.
type Writer interface {
	WriteError(command string, err error) error
}

// ChooseMode resolves ModeAuto based on whether stdout is a TTY.
// `isTTY` is injected rather than probed so tests can exercise both paths.
func ChooseMode(m Mode, isTTY bool) Mode {
	if m != ModeAuto {
		return m
	}
	if isTTY {
		return ModePretty
	}
	return ModeJSON
}

// IsStdoutTTY reports whether os.Stdout is attached to a terminal.
func IsStdoutTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

// NewWriter returns a writer for the resolved mode.
// Callers usually call NewWriter(ChooseMode(flag, IsStdoutTTY()), os.Stdout).
func NewWriter(m Mode, out io.Writer) Writer {
	switch m {
	case ModeJSON:
		return &JSONWriter{Out: out}
	case ModePretty:
		return &TableWriter{Out: out}
	default:
		// Defensive: default to JSON if mode is unresolved.
		return &JSONWriter{Out: out}
	}
}
