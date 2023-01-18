package pb

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aggronmagi/wctl/xlsx/gen"
	"github.com/aggronmagi/wctl/xlsx/parser"
)

var (
	// 本地配置
	cfg = struct {
		PackageName  string
		ExportGoPath string
	}{
		PackageName:  "xlsxpb",
		ExportGoPath: "",
	}
	// 语言配置
	language = gen.NewExportConfig("pb",
		gen.WithExportDefine(exportPBDefine),
		gen.WithExportMergeDefine(exportPB2GO),
		gen.WithCheckOptions(checkOptionConfig),
	)
)

// 注册语言函数
func Language() *gen.ExportSupportConfig {
	language.StringVar(&cfg.PackageName, "pkg", cfg.PackageName, "proto包名")
	language.StringVar(&cfg.ExportGoPath, "go-path", cfg.ExportGoPath, "导出go代码路径,未设置不导出")
	// 返回语言
	return language
}

func checkOptionConfig() error {
	return nil
}

func exportPBDefine(sheet *parser.XlsxSheet, outpath string) (err error) {
	var (
		temp = template.New("pb")
	)
	temp.Funcs(UseFuncMap)
	t, err := temp.Parse(textTemplate)
	if err != nil {
		log.Println("pb template parse err", err)
		return
	}
	// 如果有特殊模板规则
	bf := &bytes.Buffer{}
	err = t.Execute(bf, &struct {
		*parser.XlsxSheet
		PackageName string
	}{
		XlsxSheet:   sheet,
		PackageName: cfg.PackageName,
	})
	if err != nil {
		log.Println("pb template execute err", err)
		return
	}
	err = gen.WriteFile(path.Join(outpath, fmt.Sprintf("%s.proto", strings.ToLower(sheet.StructName))), bf.Bytes())
	if err != nil {
		log.Println("pb write file err", err)
		return
	}

	return
}

// 生成go代码
func exportPB2GO(sheets []*parser.XlsxSheet, outpath string) (err error) {
	if cfg.ExportGoPath == "" {
		return
	}
	cfg.ExportGoPath, _ = filepath.Abs(cfg.ExportGoPath)
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(outpath)
	if err != nil {
		log.Println("chdir failed,", err)
		return err
	}

	files := getAllFiles("./", ".proto")
	os.MkdirAll(cfg.ExportGoPath, 0755)
	// log.Println("wait protoc", files)
	protoc(cfg.ExportGoPath, files)
	// 恢复工作目录
	os.Chdir(pwd)
	return
}

func getAllFiles(dir, ext string) (lists []string) {
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			//log.Println("walk error,", err)
			return nil
		}
		// 递归目录
		if info.IsDir() && info.Name() != dir {
			lists = append(lists, getAllFiles(info.Name(), ext)...)
			return nil
		}
		// 跳过
		if filepath.Ext(info.Name()) != ".proto" {
			return nil
		}

		//log.Println("path:", path, " file:", info.Name())

		lists = append(lists, info.Name())

		return nil
	})

	return
}
func protoc(outpath string, pbFileNames []string) {

	args := make([]string, 1+len(pbFileNames))
	args[0] = fmt.Sprintf("--go_out=source_relative:%s", outpath)
	//args[1] = fmt.Sprintf("--proto_path=%s", outpath)
	copy(args[1:], pbFileNames)
	cmd := exec.Command("protoc", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	//log.Println("wait exec command")
	if err := cmd.Run(); err != nil {
		log.Println("exec protoc error")
		return
	}
	// log.Println(fmt.Sprintf("%s => %s", pbFileName, outpath))
	log.Println("generate *.pb.go finish")
}
