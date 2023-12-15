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
package builder

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/walleframe/wctl/builder/buildpb"
	"github.com/walleframe/wctl/utils"
	"google.golang.org/protobuf/proto"

	"github.com/walleframe/wctl/protocol/ast"
)

type cmdPluginGenerator struct {
	cmd  string
	name string
	args []string
}

// Generate 生成代码接口
func (gen *cmdPluginGenerator) Generate(prog *ast.YTProgram) (outs []*Output, err error) {
	if utils.ShowDetail() {
		fmt.Println("ready to generate")
	}
	// 创建命令
	cmd := exec.Command(gen.cmd, gen.args...)
	if utils.ShowDetail() {
		fmt.Println("get pipe")
	}
	// 输入输出流
	// 使用stdin发送生成请求
	writer, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	// 使用stdout读取生成文件信息
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	// 捕获stderr输出,打印到当前stdout
	pluginOutput, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	if utils.ShowDetail() {
		fmt.Println("build rq")
	}
	// 构造请求
	req := &buildpb.BuildRQ{}
	req.Files = append(req.Files, prog.File)
	req.Programs = make(map[string]*buildpb.FileDesc)
	for _, v := range prog.GetFileDescWithImports() {
		req.Programs[v.File] = v
	}
	// 序列化请求
	data, err := proto.Marshal(req)
	if err != nil {
		return
	}
	if utils.ShowDetail() {
		fmt.Println("write rq")
	}
	// data, err := json.Marshal(req)
	go func() {
		// 写入请求数据
		if utils.ShowDetail() {
			fmt.Println("ready to write cmd request. size:", len(data))
		}
		_, err = writer.Write(data)
		defer writer.Close()
		if utils.ShowDetail() {
			fmt.Println("write cmd request finish", err)
		}
		if err != nil {
			fmt.Println("write rq failed.", err)
			return
		}
	}()
	if utils.ShowDetail() {
		fmt.Println("ready for catch stderr")
	}
	// 捕获输出
	stdout := newCapturingPassThroughWriter(os.Stdout)
	go func() {
		io.Copy(stdout, pluginOutput)
	}()
	go func() {

	}()
	if utils.ShowDetail() {
		fmt.Println("start cmd...")
	}

	reply := &buildpb.BuildRS{}
	// 开始命令
	err = cmd.Start()
	if err != nil {
		fmt.Println("execute cmd plugin failed.", err)
		return
	}
	// 读取全部结果
	data, err = ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println("read result error", err)
		return
	}
	if utils.ShowDetail() {
		fmt.Println("read result len", len(data))
	}

	if utils.ShowDetail() {
		fmt.Println("wait cmd finish...")
	}
	// 等待执行
	err = cmd.Wait()
	if err != nil {
		return
	}
	if utils.ShowDetail() {
		fmt.Println("unmarshal result...")
	}

	// 解析结果
	err = proto.Unmarshal(data, reply)
	if err != nil {
		return
	}
	//
	for _, v := range reply.Result {
		outs = append(outs, &Output{
			File: v.File,
			Data: v.Data,
		})
	}
	if utils.ShowDetail() {
		fmt.Println("plugin-cmd ", gen.cmd, ", finish ")
	}

	// io.Pipe()
	return
}

// Union 唯一标识符 用于标识不同插件
func (gen *cmdPluginGenerator) Union() string {
	return gen.name
}

// NewCmdPluginGenerater 新建命令行插件-代码生成器
func NewCmdPluginGenerater(cmd string) (err error) {

	path, err := exec.LookPath(cmd)
	if err != nil {
		return
	}
	if utils.Debug() {
		fmt.Println("find command plugin [", cmd, "] in path[", path, "].")
	}
	gen := &cmdPluginGenerator{
		cmd:  cmd,
		name: "cmd-plugin-" + cmd,
	}

	if utils.Debug() {
		gen.args = append(gen.args, "--debug")
	}
	if utils.ShowDetail() {
		gen.args = append(gen.args, "--debug-detail")
	}

	if last, ok := factory[gen.Union()]; ok {
		fmt.Println("WARN 使用命令行插件 替换插件:", last.Union(), cmd)
	}
	// 保存生成器
	factory[gen.Union()] = gen
	// 生效插件
	addUse(gen)
	return
}

// capturingPassThroughWriter is a writer that remembers
// data written to it and passes it to w
type capturingPassThroughWriter struct {
	buf bytes.Buffer
	w   io.Writer
}

// newCapturingPassThroughWriter creates new capturingPassThroughWriter
func newCapturingPassThroughWriter(w io.Writer) *capturingPassThroughWriter {
	return &capturingPassThroughWriter{
		w: w,
	}
}
func (w *capturingPassThroughWriter) Write(d []byte) (int, error) {
	w.buf.Write(d)
	return w.w.Write(d)
}

// Bytes returns bytes written to the writer
func (w *capturingPassThroughWriter) Bytes() []byte {
	return w.buf.Bytes()
}
