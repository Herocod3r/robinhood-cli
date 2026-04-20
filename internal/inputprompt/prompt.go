// Package inputprompt provides small stdin/stdout prompt helpers used by `rh login`.
//
// The Prompter is injectable so tests can feed scripted input and interactive
// flows can inject real stdin/stdout. Password reads bypass In/Out and use
// x/term directly when stdin is a TTY; tests force the non-TTY branch.
package inputprompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompter reads interactive user input.
type Prompter struct {
	In              io.Reader
	Out             io.Writer
	MaxEmptyRetries int // 0 = default 3
}

// New returns a Prompter bound to os.Stdin / os.Stderr (prompts go to stderr
// so stdout stays clean for structured output).
func New() *Prompter {
	return &Prompter{In: os.Stdin, Out: os.Stderr}
}

// Text prompts and reads a line, trimming whitespace. Retries up to MaxEmptyRetries times.
func (p *Prompter) Text(label string) (string, error) {
	retries := p.MaxEmptyRetries
	if retries <= 0 {
		retries = 3
	}
	r := bufio.NewReader(p.In)
	for i := 0; i < retries; i++ {
		if _, err := fmt.Fprint(p.Out, label); err != nil {
			return "", err
		}
		line, err := r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line != "" {
			return line, nil
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	return "", errors.New("empty input after retries")
}

// Password prompts and reads a line with echo suppressed. Falls back to plain
// read when stdin is not a TTY (tests, pipelines).
func (p *Prompter) Password(label string) (string, error) {
	if _, err := fmt.Fprint(p.Out, label); err != nil {
		return "", err
	}
	f, ok := p.In.(*os.File)
	if ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		if _, werr := fmt.Fprintln(p.Out); werr != nil {
			return "", werr
		}
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	// Non-TTY fallback.
	r := bufio.NewReader(p.In)
	line, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// Choose prints a menu and returns the zero-based index of the selection.
func (p *Prompter) Choose(label string, options []string) (int, error) {
	if len(options) == 0 {
		return 0, errors.New("no options")
	}
	if _, err := fmt.Fprintln(p.Out, label); err != nil {
		return 0, err
	}
	for i, o := range options {
		if _, err := fmt.Fprintf(p.Out, "  [%d] %s\n", i+1, o); err != nil {
			return 0, err
		}
	}
	for {
		s, err := p.Text("> ")
		if err != nil {
			return 0, err
		}
		var idx int
		if _, err := fmt.Sscanf(s, "%d", &idx); err == nil && idx >= 1 && idx <= len(options) {
			return idx - 1, nil
		}
		_, _ = fmt.Fprintln(p.Out, "enter a number from the list")
	}
}
