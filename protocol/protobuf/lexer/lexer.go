// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"os"
	"unicode/utf8"

	"github.com/walleframe/wctl/protocol/token"
)

const (
	NoState    = -1
	NumStates  = 92
	NumSymbols = 106
)

type Lexer struct {
	src     []byte
	pos     int
	line    int
	column  int
	Context token.Context
}

func NewLexer(src []byte) *Lexer {
	lexer := &Lexer{
		src:     src,
		pos:     0,
		line:    1,
		column:  1,
		Context: nil,
	}
	return lexer
}

// SourceContext is a simple instance of a token.Context which
// contains the name of the source file.
type SourceContext struct {
	Filepath string
}

func (s *SourceContext) Source() string {
	return s.Filepath
}

func NewLexerFile(fpath string) (*Lexer, error) {
	src, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	lexer := NewLexer(src)
	lexer.Context = &SourceContext{Filepath: fpath}
	return lexer, nil
}

func (l *Lexer) Scan() (tok *token.Token) {
	tok = &token.Token{}
	if l.pos >= len(l.src) {
		tok.Type = token.EOF
		tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = l.pos, l.line, l.column
		tok.Pos.Context = l.Context
		return
	}
	start, startLine, startColumn, end := l.pos, l.line, l.column, 0
	tok.Type = token.INVALID
	state, rune1, size := 0, rune(-1), 0
	for state != -1 {
		if l.pos >= len(l.src) {
			rune1 = -1
		} else {
			rune1, size = utf8.DecodeRune(l.src[l.pos:])
			l.pos += size
		}

		nextState := -1
		if rune1 != -1 {
			nextState = TransTab[state](rune1)
		}
		state = nextState

		if state != -1 {

			switch rune1 {
			case '\n':
				l.line++
				l.column = 1
			case '\r':
				l.column = 1
			case '\t':
				l.column += 4
			default:
				l.column++
			}

			switch {
			case ActTab[state].Accept != -1:
				tok.Type = ActTab[state].Accept
				end = l.pos
			case ActTab[state].Ignore != "":
				start, startLine, startColumn = l.pos, l.line, l.column
				state = 0
				if start >= len(l.src) {
					tok.Type = token.EOF
				}

			}
		} else {
			if tok.Type == token.INVALID {
				end = l.pos
			}
		}
	}
	if end > start {
		l.pos = end
		tok.Lit = l.src[start:end]
	} else {
		tok.Lit = []byte{}
	}
	tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = start, startLine, startColumn
	tok.Pos.Context = l.Context

	return
}

func (l *Lexer) Reset() {
	l.pos = 0
}

/*
Lexer symbols:
0: ';'
1: 's'
2: 'y'
3: 'n'
4: 't'
5: 'a'
6: 'x'
7: '='
8: 'p'
9: 'a'
10: 'c'
11: 'k'
12: 'a'
13: 'g'
14: 'e'
15: 'i'
16: 'm'
17: 'p'
18: 'o'
19: 'r'
20: 't'
21: 'e'
22: 'n'
23: 'u'
24: 'm'
25: '{'
26: '}'
27: 'o'
28: 'p'
29: 't'
30: 'i'
31: 'o'
32: 'n'
33: 'm'
34: 'e'
35: 's'
36: 's'
37: 'a'
38: 'g'
39: 'e'
40: 'm'
41: 'a'
42: 'p'
43: '<'
44: ','
45: '>'
46: 'r'
47: 'e'
48: 'p'
49: 'e'
50: 'a'
51: 't'
52: 'e'
53: 'd'
54: 's'
55: 'e'
56: 'r'
57: 'v'
58: 'i'
59: 'c'
60: 'e'
61: 'r'
62: 'p'
63: 'c'
64: '('
65: ')'
66: 'r'
67: 'e'
68: 't'
69: 'u'
70: 'r'
71: 'n'
72: 's'
73: '_'
74: '.'
75: '['
76: ']'
77: '<'
78: '>'
79: '`'
80: '`'
81: '"'
82: '"'
83: '+'
84: '-'
85: '0'
86: 'x'
87: '/'
88: '*'
89: '*'
90: '/'
91: '/'
92: '/'
93: '\n'
94: ' '
95: '\t'
96: '\n'
97: '\r'
98: '#'
99: '\n'
100: '0'-'9'
101: 'a'-'z'
102: 'A'-'Z'
103: 'a'-'f'
104: 'A'-'F'
105: .
*/
