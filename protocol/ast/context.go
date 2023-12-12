package ast

import (
	"github.com/aggronmagi/wctl/protocol/token"
	"github.com/aggronmagi/wctl/utils"
)

type Context struct {
	Prog *YTProgram
	//Doc         *YTDoc
	LastElement interface{}
	Docs        []*token.Token
}

func (ctx *Context) Range(tok *token.Token, tm *token.TokenMap) {
	utils.Trackln("scan ", tm.Id(tok.Type), tok.String())
}

func (ctx *Context) AddDoc(tok *token.Token) {
	ctx.Docs = append(ctx.Docs, tok)
}

func (ctx *Context) PreDoc(tokLine int) *YTDoc {
	pre, tail := false, false
	doc := &YTDoc{}

	// 前面的全部注释
	preDocs := make([]*token.Token, 0, len(ctx.Docs))
	trimIdx := -1

	for k, d := range ctx.Docs {
		preDocs = append(preDocs, d)
		// 当前token前的最后一条注释
		if d.Line+1 == tokLine {
			trimIdx = k
		}
		// 尾注释
		if d.Line == tokLine {
			doc.TailDoc = d.IDValue()
			ctx.Docs = append(ctx.Docs[:k], ctx.Docs[k+1:]...)
			tail = true
			break
		}
	}
	// 去除前面的全部token
	if trimIdx >= 0 {
		// ctx.Docs = append(ctx.Docs[:0], ctx.Docs[trimIdx+1:]...)

		// 前置注释必须是连续的, 所以中间如果跨行的注释,忽略跨行前的注释文档.
		preIdx, preLine := trimIdx, preDocs[trimIdx].Line
		for i := trimIdx - 1; i >= 0; i-- {
			d := preDocs[i]
			if d.Line+1 >= preLine {
				preLine = d.Line
				preIdx = i
			}
		}
		// 连续的前置注释
		for i := preIdx; i <= trimIdx; i++ {
			doc.Doc = append(doc.Doc, preDocs[i].IDValue())
			pre = true
		}

		ctx.Docs = append(ctx.Docs[:preIdx], ctx.Docs[trimIdx+1:]...)
	}

	if pre || tail {
		return doc
	}
	return nil
}
