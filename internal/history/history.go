// Package history reads shell history files and finds past calls of a command.
package history

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Time   *time.Time `json:"time,omitempty"` // nil when the shell did not record it (plain bash)
	Line   string     `json:"line"`
	Source string     `json:"source"`
}

// Sources lists the history files that exist: histfile (usually $HISTFILE),
// then the default zsh, bash and fish locations.
func Sources(home, histfile, xdgData string) []string {
	if xdgData == "" {
		xdgData = filepath.Join(home, ".local", "share")
	}
	cands := []string{histfile,
		filepath.Join(home, ".zsh_history"),
		filepath.Join(home, ".bash_history"),
		filepath.Join(xdgData, "fish", "fish_history"),
	}
	var out []string
	seen := map[string]bool{}
	for _, p := range cands {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			out = append(out, p)
		}
	}
	return out
}

// Load parses every path and returns the entries oldest first. Entries without
// a timestamp come before timestamped ones, each group in file order.
func Load(paths []string) ([]Entry, error) {
	var all []Entry
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		es, err := parse(f, p)
		_ = f.Close() // read-only
		if err != nil {
			return nil, err
		}
		all = append(all, es...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		a, b := all[i].Time, all[j].Time
		switch {
		case a == nil:
			return b != nil
		case b == nil:
			return false
		}
		return a.Before(*b)
	})
	return all, nil
}

var zshExtended = regexp.MustCompile(`^: (\d+):\d+;(.*)$`)

// parse picks the format from the file name, then from the first line.
func parse(r io.Reader, path string) ([]Entry, error) {
	br := bufio.NewReader(r)
	name := strings.ToLower(filepath.Base(path))
	switch {
	case strings.Contains(name, "fish"):
		return parseFish(br, path)
	case strings.Contains(name, "bash"):
		return parseBash(br, path)
	case strings.Contains(name, "zsh"):
		return parseZsh(br, path)
	}
	head, _ := br.Peek(64)
	first, _, _ := strings.Cut(string(head), "\n")
	switch {
	case zshExtended.MatchString(first):
		return parseZsh(br, path)
	case strings.HasPrefix(first, "- cmd:"):
		return parseFish(br, path)
	}
	return parseBash(br, path)
}

func readLines(r *bufio.Reader, fn func(string)) error {
	for {
		line, err := r.ReadString('\n')
		if line != "" {
			fn(strings.TrimRight(line, "\r\n"))
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func unix(s string) *time.Time {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	t := time.Unix(n, 0)
	return &t
}

// unmetafy reverses zsh's on-disk escaping of special bytes (0x83 then byte^32).
func unmetafy(s string) string {
	if !strings.Contains(s, "\x83") {
		return s
	}
	b := []byte(s)
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] == 0x83 && i+1 < len(b) {
			i++
			out = append(out, b[i]^32)
			continue
		}
		out = append(out, b[i])
	}
	return string(out)
}

// parseZsh handles both EXTENDED_HISTORY (": ts:dur;cmd") and plain lines.
// A line ending in a backslash continues on the next line.
func parseZsh(r *bufio.Reader, src string) ([]Entry, error) {
	var out []Entry
	var cur *Entry
	cont := false
	err := readLines(r, func(raw string) {
		raw = unmetafy(raw)
		if cont && cur != nil {
			cur.Line += "\n" + strings.TrimSuffix(raw, "\\")
			cont = strings.HasSuffix(raw, "\\")
			return
		}
		e := Entry{Source: src}
		if m := zshExtended.FindStringSubmatch(raw); m != nil {
			e.Time, e.Line = unix(m[1]), m[2]
		} else {
			e.Line = raw
		}
		cont = strings.HasSuffix(e.Line, "\\")
		e.Line = strings.TrimSuffix(e.Line, "\\")
		out = append(out, e)
		cur = &out[len(out)-1]
	})
	return out, err
}

// parseBash handles plain lines and HISTTIMEFORMAT's "#<unix>" lines.
func parseBash(r *bufio.Reader, src string) ([]Entry, error) {
	var out []Entry
	var ts *time.Time
	err := readLines(r, func(line string) {
		if strings.HasPrefix(line, "#") {
			if t := unix(line[1:]); t != nil {
				ts = t
				return
			}
		}
		if strings.TrimSpace(line) == "" {
			return
		}
		out = append(out, Entry{Time: ts, Line: line, Source: src})
		ts = nil
	})
	return out, err
}

var fishUnescape = strings.NewReplacer(`\\`, `\`, `\n`, "\n")

// parseFish handles fish's YAML-like "- cmd: ...\n  when: <unix>" records.
func parseFish(r *bufio.Reader, src string) ([]Entry, error) {
	var out []Entry
	err := readLines(r, func(line string) {
		switch {
		case strings.HasPrefix(line, "- cmd: "):
			out = append(out, Entry{Line: fishUnescape.Replace(line[len("- cmd: "):]), Source: src})
		case strings.HasPrefix(line, "  when: ") && len(out) > 0:
			out[len(out)-1].Time = unix(strings.TrimSpace(line[len("  when: "):]))
		}
	})
	return out, err
}
