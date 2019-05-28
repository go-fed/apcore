package apcore

import (
	"strings"
)

const (
	// Clarke says something shorter than usual.
	// Line 1 usable range is [86:138] (len=52)
	// Line 2 usable range is [156:209] (len=53)
	// Line 3 usable range is [227:280] (len=53)
	clarkeShort = `               ______________________________________________________
       \__/   /                                                      \
\------(oo)  /                                                       |
 ||    (__) <                                                        |
 ||w--||     \-------------------------------------------------------/`
	// Clarke says something longer.
	// Lines 1 - 3 are the same as above
	// Line 4 usable range is [298:351] (len=53)
	clarkeLongBegin = `               ______________________________________________________
       \__/   /                                                      \
\------(oo)  /                                                       |
 ||    (__) <                                                        |
 ||w--||     |                                                       |`
	// Line usable range is [15:68] (len=53)
	clarkeLongMiddle = `             |                                                       |`
	clarkeLongEnd    = `             \-------------------------------------------------------/`
)

func replace(input, replace string, offset int) string {
	in := []byte(input)
	repl := []byte(replace)
	return string(append(in[:offset],
		append(repl,
			in[offset+len(repl):]...)...))
}

func clarkeSays(moo string) string {
	moo = strings.TrimSpace(strings.ReplaceAll(moo, "\n", " "))
	words := strings.Split(moo, " ")
	lines := make([][]string, 0, 1)
	var line []string
	var length int
	for _, word := range words {
		maxlen := 53
		if len(lines) == 0 {
			maxlen = 52
		}
		sl := 0
		if len(line) > 0 {
			sl = 1
		}
		if length+len(word)+sl > maxlen {
			lines = append(lines, line)
			line = []string{word}
			length = len(word)
		} else {
			line = append(line, word)
			length += len(word) + sl
		}
	}
	lines = append(lines, line)

	var s string
	switch len(lines) {
	case 1:
		// Middle line
		s = clarkeShort
		s = replace(s, strings.Join(lines[0], " "), 156)
	case 2:
		// Middle and bottom line
		s = clarkeShort
		s = replace(s, strings.Join(lines[0], " "), 156)
		s = replace(s, strings.Join(lines[1], " "), 226)
	case 3:
		// Top, middle and bottom line
		s = clarkeShort
		s = replace(s, strings.Join(lines[0], " "), 86)
		s = replace(s, strings.Join(lines[1], " "), 156)
		s = replace(s, strings.Join(lines[2], " "), 227)
	default:
		// Long paragraph.
		s = clarkeLongBegin
		s = replace(s, strings.Join(lines[0], " "), 86)
		s = replace(s, strings.Join(lines[1], " "), 156)
		s = replace(s, strings.Join(lines[2], " "), 227)
		s = replace(s, strings.Join(lines[3], " "), 298)
		if len(lines) > 4 {
			for i := 4; i < len(lines); i++ {
				m := clarkeLongMiddle
				m = replace(m, strings.Join(lines[i], " "), 15)
				s += "\n"
				s += m
			}
		}
		s += "\n"
		s += clarkeLongEnd
	}
	s += "\n"
	return s
}
