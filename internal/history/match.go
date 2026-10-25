package history

import (
	"path/filepath"
	"strings"
)

type token struct {
	text string
	op   bool // unquoted shell operator: | || & && ; ( )
}

// tokenize splits a command line into words and operators, honoring quotes and
// backslashes. It is not a shell parser: it only has to find where commands start.
func tokenize(line string) []token {
	var toks []token
	var cur strings.Builder
	has := false
	flush := func() {
		if has {
			toks = append(toks, token{text: cur.String()})
			cur.Reset()
			has = false
		}
	}
	rs := []rune(line)
	var quote rune
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else if r == '\\' && quote == '"' && i+1 < len(rs) {
				i++
				cur.WriteRune(rs[i])
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, has = r, true
		case r == '\\' && i+1 < len(rs):
			i++
			cur.WriteRune(rs[i])
			has = true
		case r == ' ' || r == '\t' || r == '\n':
			flush()
		case r == '(' || r == ')':
			flush()
			toks = append(toks, token{text: string(r), op: true})
		case r == '|' || r == ';' || r == '&':
			s := cur.String()
			redirect := r == '&' && (strings.HasSuffix(s, ">") || strings.HasSuffix(s, "<") || (i+1 < len(rs) && rs[i+1] == '>'))
			if redirect {
				cur.WriteRune(r)
				has = true
				continue
			}
			flush()
			j := i
			for j+1 < len(rs) && (rs[j+1] == '|' || rs[j+1] == ';' || rs[j+1] == '&') {
				j++
			}
			toks = append(toks, token{text: string(rs[i : j+1]), op: true})
			i = j
		default:
			cur.WriteRune(r)
			has = true
		}
	}
	flush()
	return toks
}

var wrappers = map[string]bool{
	"sudo": true, "doas": true, "time": true, "command": true, "exec": true,
	"nohup": true, "noglob": true, "nocorrect": true, "builtin": true, "env": true, "nice": true,
}

func isAssignment(w string) bool {
	i := strings.IndexByte(w, '=')
	if i <= 0 {
		return false
	}
	for j, r := range w[:i] {
		letter := r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		digit := j > 0 && r >= '0' && r <= '9'
		if !letter && !digit {
			return false
		}
	}
	return true
}

// Match reports whether line calls cmd, directly or after env assignments and
// wrappers such as sudo, in any pipeline or && / ; segment. The first
// len(args) words after the command must equal args.
func Match(line, cmd string, args []string) bool {
	cmd = filepath.Base(cmd)
	var seg []string
	check := func() bool {
		i := 0
		for i < len(seg) {
			w := seg[i]
			switch {
			case isAssignment(w):
			case wrappers[filepath.Base(w)]:
				// skip the wrapper's own flags: sudo -E, env -i, nice -n (value not handled)
				for i+1 < len(seg) && strings.HasPrefix(seg[i+1], "-") {
					i++
				}
			default:
				if filepath.Base(w) != cmd {
					return false
				}
				rest := seg[i+1:]
				if len(rest) < len(args) {
					return false
				}
				for k, a := range args {
					if rest[k] != a {
						return false
					}
				}
				return true
			}
			i++
		}
		return false
	}
	for _, t := range tokenize(line) {
		if t.op {
			if check() {
				return true
			}
			seg = seg[:0]
			continue
		}
		seg = append(seg, t.text)
	}
	return check()
}

// Filter returns the entries that call cmd with the given leading args.
func Filter(es []Entry, cmd string, args []string) []Entry {
	var out []Entry
	for _, e := range es {
		if Match(e.Line, cmd, args) {
			out = append(out, e)
		}
	}
	return out
}
