package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walleframe/wctl/protocol/token"
)

func TestPreDoc(t *testing.T) {
	ctx := &Context{
		Prog:        &YTProgram{},
		LastElement: nil,
		Docs: []*token.Token{
			{
				Lit: []byte("// pre doc"),
				Pos: token.Pos{
					Line: 1,
				},
			},
			{
				Lit: []byte("// tail doc"),
				Pos: token.Pos{
					Line: 2,
				},
			},
		},
	}

	doc := ctx.PreDoc(2)
	assert.NotNil(t, doc, "need return documents")
	assert.EqualValues(t, 0, len(ctx.Docs), "no more docs")
	assert.Contains(t, doc.Doc[0], "pre", "pre doc")
	assert.Contains(t, doc.TailDoc, "tail", "tail doc")

	ctx = &Context{
		Prog:        &YTProgram{},
		LastElement: nil,
		Docs: []*token.Token{
			{
				Lit: []byte("// pre doc"),
				Pos: token.Pos{
					Line: 1,
				},
			},
			{
				Lit: []byte("// tail doc"),
				Pos: token.Pos{
					Line: 2,
				},
			},
			{
				Lit: []byte("// import doc"),
				Pos: token.Pos{
					Line: 3,
				},
			},
		},
	}

	doc = ctx.PreDoc(2)
	assert.NotNil(t, doc, "need return documents")
	assert.EqualValues(t, 1, len(ctx.Docs), "left import docs")
	assert.Contains(t, doc.Doc[0], "pre", "pre doc")
	assert.Contains(t, doc.TailDoc, "tail", "tail doc")
}

func TestPreDoc2(t *testing.T) {
	ctx := &Context{
		Prog:        &YTProgram{},
		LastElement: nil,
		Docs: []*token.Token{
			{
				Lit: []byte("// pre doc"),
				Pos: token.Pos{
					Line: 1,
				},
			},
			{
				Lit: []byte("// tail doc"),
				Pos: token.Pos{
					Line: 2,
				},
			},
			{
				Lit: []byte("// 4line"),
				Pos: token.Pos{
					Line: 4,
				},
			},
			{
				Lit: []byte("// 5line"),
				Pos: token.Pos{
					Line: 5,
				},
			},
		},
	}

	doc := ctx.PreDoc(5)
	assert.NotNil(t, doc, "need return documents")
	assert.EqualValues(t, 2, len(ctx.Docs), "no more docs")
	assert.Contains(t, doc.Doc[0], "4line", "pre doc")
	assert.Contains(t, doc.TailDoc, "5line", "tail doc")

}
