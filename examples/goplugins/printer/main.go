package main

import (
	"fmt"

	"github.com/walleframe/wctl/builder"
	"github.com/walleframe/wctl/protocol/ast"
	"github.com/walleframe/wctl/utils"
)

// go plugin
func main() {

}

// NewGenerator 新建生成器
func NewGenerator() builder.Generater {
	return &printGenerater{}
}

type printGenerater struct{}

// Generate 生成代码接口
func (gen *printGenerater) Generate(prog *ast.YTProgram) (outs []*builder.Output, err error) {
	fmt.Println(utils.Sdump(prog, "recv"))
	return
}

// Union 唯一标识符 用于标识不同插件
func (gen *printGenerater) Union() string {
	return "printer"
}
