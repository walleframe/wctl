package protobuf

import (
	"github.com/aggronmagi/wctl/protocol/ast"
	"github.com/aggronmagi/wctl/protocol/protobuf/lexer"
	"github.com/aggronmagi/wctl/protocol/protobuf/parser"
	"github.com/aggronmagi/wctl/protocol/token"
	"github.com/aggronmagi/wctl/utils"
)

func Parse(file string, src []byte) (_ *ast.YTProgram, err error) {
	l := lexer.NewLexer(src)
	l.Context = &lexer.SourceContext{Filepath: file}

	ctx := &ast.Context{
		Prog: &ast.YTProgram{},
	}

	p := parser.NewParser()
	p.Context = ctx

	res, err := p.Parse(wrapLexer(ctx, l))
	if err != nil {
		utils.Dump(err)
		return nil, err
	}

	return res.(*ast.YTProgram), nil
}

var tokDoc = parser.TokMap.Type("tok_doc")

type wLexer struct {
	ctx  *ast.Context
	scan parser.Scanner
}

func wrapLexer(ctx *ast.Context, s parser.Scanner) parser.Scanner {
	return &wLexer{
		ctx:  ctx,
		scan: s,
	}
}

func (l *wLexer) Scan() (tok *token.Token) {
	ctx := l.ctx
	tok = l.scan.Scan()
	ctx.Range(tok, &parser.TokMap)
	for tok.Type == tokDoc {
		ctx.AddDoc(tok)
		tok = l.scan.Scan()
		ctx.Range(tok, &parser.TokMap)
	}
	return
}
