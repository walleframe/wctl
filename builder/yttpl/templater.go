/*
   Copyright © 2020 aggronmagi <czy463@163.com>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package yttpl

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/aggronmagi/wctl/builder"
	"github.com/aggronmagi/wctl/protocol/ast"
	"github.com/aggronmagi/wctl/utils"
	"gopkg.in/yaml.v3"
)

// 模板生成器配置
type config struct {
	// 模板生成器名称. 全局唯一
	Union string
	// 模板文件所在文件夹
	Path string
	// 模板文件名后缀
	Suffix string
	// 模板配置
	Tpl []string
}

// 模板生成器参数
type tplArg struct {
	*ast.YTProgram
	// 输出文件名.(含后缀)
	Out   string
	gofmt bool
}

type tplGenerater struct {
	tpl *template.Template
	cfg *config
	arg *tplArg
}

// Generate 生成代码接口
func (gen *tplGenerater) Generate(prog *ast.YTProgram) (outs []*builder.Output, err error) {
	gen.arg.YTProgram = prog
	buf := &bytes.Buffer{}
	var data []byte

	// 遍历配置模板
	for _, v := range gen.cfg.Tpl {
		gen.arg.gofmt = false
		gen.arg.Out = ""
		// 执行模板
		err = gen.tpl.ExecuteTemplate(buf, v, gen.arg)
		if err != nil {
			return
		}
		data = buf.Bytes()
		if gen.arg.gofmt {
			data, err = format.Source(data)
			if err != nil {
				fmt.Println("Error: 格式化输出错误:", err)
				data = buf.Bytes()
			}
		}
		// 保存输出
		outs = append(outs, &builder.Output{
			File: gen.arg.Out,
			Data: data,
		})
		fmt.Println(string(data))
		fmt.Println("out", gen.arg.Out)
	}
	return
}

// Union 唯一标识符 用于标识不同插件
func (gen *tplGenerater) Union() string {
	return gen.cfg.Union
}

// NewTemplateGenerator 新建template生成器
func NewTemplateGenerator(cfgName string) (err error) {
	// 读取配置文件
	data, err := ioutil.ReadFile(cfgName)
	if err != nil {
		return
	}
	// 转换路径
	path, err := filepath.Abs(filepath.Clean(filepath.Dir(cfgName)))
	if err != nil {
		return
	}
	return NewTemplateGeneratorByCfgData(data, path)
}

// NewTemplateGeneratorByCfgData 新建template生成器
func NewTemplateGeneratorByCfgData(data []byte, path string) (err error) {
	// 解析配置
	cfg := &config{}
	err = yaml.Unmarshal(data, cfg)
	if utils.Debug() {
		fmt.Println(cfg)
	}
	if err != nil {
		return
	}

	// 生成器
	tpl := &tplGenerater{
		cfg: cfg,
		arg: &tplArg{},
	}

	tg := template.New(cfg.Union).Funcs(template.FuncMap{
		"out": func(out string) (none bool, err error) {
			tpl.arg.Out = out
			return
		},
		"gofmt": func() (none bool, err error) {
			tpl.arg.gofmt = true
			return
		},
	})
	tg.Funcs(gTplFunc)
	// 解析go模板文件
	tg, err = tg.ParseGlob(path + "/*." + cfg.Suffix)
	if err != nil {
		return
	}
	tpl.tpl = tg

	// 注册生成器
	builder.RegisterGenerater(tpl)
	// 开启生成器
	err = builder.EnableGenerator(tpl.Union())
	if err != nil {
		return
	}
	return
}

var gTplFunc = template.FuncMap{
	"normal": gTplNormalize,
}

func gTplNormalize(name string) (final string, err error) {
	buf := &strings.Builder{}
	buf.Grow(len(name))
	index := strings.IndexByte(name, '.')
	if index >= 0 {
		buf.WriteString(name[:index+1])
		name = name[index+1:]
	}

	change := false
	for k, v := range name {
		if k == 0 {
			v = unicode.ToTitle(v)
			buf.WriteRune(v)
			continue
		}
		if v == '_' {
			change = true
			continue
		}
		if change {
			v = unicode.ToTitle(v)
		}
		buf.WriteRune(v)
	}
	return buf.String(), nil
}
