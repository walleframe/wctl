// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"io/ioutil"
	"unicode/utf8"

	"github.com/aggronmagi/wctl/protocol/token"
)

const (
	NoState    = -1
	NumStates  = 134
	NumSymbols = 160
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
	src, err := ioutil.ReadFile(fpath)
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
0: '.'
1: 'p'
2: 'a'
3: 'c'
4: 'k'
5: 'a'
6: 'g'
7: 'e'
8: ';'
9: ','
10: 'i'
11: 'm'
12: 'p'
13: 'o'
14: 'r'
15: 't'
16: '='
17: 'e'
18: 'n'
19: 'u'
20: 'm'
21: '{'
22: '}'
23: 'm'
24: 'e'
25: 's'
26: 's'
27: 'a'
28: 'g'
29: 'e'
30: '['
31: ']'
32: 'r'
33: 'e'
34: 'p'
35: 'e'
36: 'a'
37: 't'
38: 'e'
39: 'd'
40: 'm'
41: 'a'
42: 'p'
43: '<'
44: '>'
45: 'i'
46: 'n'
47: 't'
48: '8'
49: 'u'
50: 'i'
51: 'n'
52: 't'
53: '8'
54: 'i'
55: 'n'
56: 't'
57: '1'
58: '6'
59: 'u'
60: 'i'
61: 'n'
62: 't'
63: '1'
64: '6'
65: 'i'
66: 'n'
67: 't'
68: '3'
69: '2'
70: 'u'
71: 'i'
72: 'n'
73: 't'
74: '3'
75: '2'
76: 'i'
77: 'n'
78: 't'
79: '6'
80: '4'
81: 'u'
82: 'i'
83: 'n'
84: 't'
85: '6'
86: '4'
87: 's'
88: 't'
89: 'r'
90: 'i'
91: 'n'
92: 'g'
93: 'b'
94: 'y'
95: 't'
96: 'e'
97: 's'
98: 'b'
99: 'o'
100: 'o'
101: 'l'
102: 's'
103: 'e'
104: 'r'
105: 'v'
106: 'i'
107: 'c'
108: 'e'
109: 'o'
110: 'n'
111: 'e'
112: 'w'
113: 'a'
114: 'y'
115: ':'
116: 'n'
117: 'o'
118: 't'
119: 'i'
120: 'f'
121: 'y'
122: 't'
123: 'w'
124: 'o'
125: 'w'
126: 'a'
127: 'y'
128: '('
129: ')'
130: 'p'
131: 'r'
132: 'o'
133: 'j'
134: 'e'
135: 'c'
136: 't'
137: '+'
138: '-'
139: '_'
140: '/'
141: '*'
142: '*'
143: '/'
144: '/'
145: '/'
146: '\n'
147: '`'
148: '`'
149: '"'
150: '"'
151: ' '
152: '\t'
153: '\n'
154: '\r'
155: '#'
156: '\n'
157: '0'-'9'
158: 'a'-'z'
159: .
*/
