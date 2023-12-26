package xlsxgen

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/multierr"

	"github.com/walleframe/wctl/xlsx/gen"
	"github.com/walleframe/wctl/xlsx/parser"

	"github.com/walleframe/wctl/xlsx/gen/golang"
	"github.com/walleframe/wctl/xlsx/gen/jsondata"
	"github.com/walleframe/wctl/xlsx/gen/pb"
	"github.com/walleframe/wctl/xlsx/gen/wpb"
)

var cfg = struct {
	Recursion  bool
	FileExtern string
	// 检测数据合理性脚本的lua目录
	VerifyScriptPath string
}{
	FileExtern: ".xlsx",
	Recursion:  true,
}

const (
	// Version 命令行工具版本
	Version = "0.0.1"

	Help = `excel 配置导出工具 
`
	Example = `
wctl xlsx x1.xlsx x2.xlsx ./xxx/ --json-data=./json-data ...
`
)

func Flags(genCmd *pflag.FlagSet) {
	// 参数不排序
	genCmd.SortFlags = false

	genCmd.BoolVarP(&cfg.Recursion, "recursion", "r", cfg.Recursion, "递归深层目录,用于扫描xlsx文件")
	genCmd.StringVar(&cfg.FileExtern, "ext", cfg.FileExtern, "xlsx文件后缀")
	genCmd.StringVar(&cfg.VerifyScriptPath, "verify-script-path", cfg.VerifyScriptPath, "检测数据合理性脚本的lua目录")

	// 注册语言
	langCaches = append(langCaches,
		// json-data
		jsondata.Language(),
		// golang pb
		pb.Language(),
		// wpb
		wpb.Language(),
		// golang code
		golang.Language(),
		// //
		// luatpl.Language(),
	)
	// 注册标记
	for _, cfg := range langCaches {
		cfg.SetFlagSet(genCmd)
	}
}

var langCaches []*gen.ExportSupportConfig

// RunCommand run generate command
func RunCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.PrintErrln("invalid args, should input xlsx file or path.")
		cmd.Help()
		return
	}
	err := fixConfig()
	if err != nil {
		return
	}
	//
	parser.RegisterDefaultChecker()
	parser.RegisterDefaultType()
	////////////////////////////////////////////////////////////
	// 解析全部xlsx文件 - TODO:并发处理
	var errs error
	for _, arg := range args {
		if dir, err := isDir(arg); err != nil {
			log.Println("open file failed", arg, err)
			errs = multierr.Append(errs, err)
			continue
		} else if dir {
			// 输入的目录
			err = parseDir(arg)
			if err != nil {
				errs = multierr.Append(errs, err)
			}
		} else {
			// 输入的文件
			err = parseFile(arg)
			if err != nil {
				errs = multierr.Append(errs, err)
			}
		}
	}
	// 是否解析出错
	if errs != nil {
		log.Println("parse xlsx files failed")
		return
	}
	// 限制检测
	errs = parser.LimitCheck(GlobalCache.AllTables)
	if errs != nil {
		log.Println("inner limit")
		return
	}
	////////////////////////////////////////////////////////////
	// 检测配置
	var luaErrors []error = make([]error, 0, 256)
	var L = parser.PrepareCheckTable(&luaErrors)
	defer L.Close()
	// 设置sheet数据
	for _, table := range GlobalCache.AllTables {
		err := parser.SetLuaCheckTable(L, table)
		if err != nil {
			log.Println("set lua data failed. from ", table.FromFile, table.SheetName)
			for _, v := range multierr.Errors(err) {
				log.Println("\t", v)
			}
			errs = multierr.Append(errs, err)
		}
	}
	// 是否出错
	if errs != nil {
		log.Println("prepare data for check failed")
		return
	}
	// 检测数据合理性脚本的lua目录
	if cfg.VerifyScriptPath != "" {
		filepath.Walk(cfg.VerifyScriptPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			// 不递归目录
			if info.IsDir() {
				return nil
			}
			// 过滤隐藏文件及非lua文件
			if strings.HasPrefix(info.Name(), ".") || filepath.Ext(info.Name()) != ".lua" {
				return nil
			}
			// 加载文件
			fname := filepath.Join(filepath.Clean(cfg.VerifyScriptPath), info.Name())
			data, err := os.ReadFile(fname)
			if err != nil {
				log.Println("load lua check file ", fname, err)
				return nil
			}
			// 保存
			GlobalCache.AllChecks = append(GlobalCache.AllChecks, &parser.XlsxCheckSheet{
				FromFile:   fname,
				Sheet:      fname,
				LuaScripts: map[string]string{fname: string(data)},
			})

			return nil
		})
	}
	// 检测sheet数据
	for _, check := range GlobalCache.AllChecks {
		err := parser.LuaCheckTable(L, check, &luaErrors)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	// 是否出错
	if errs != nil {
		log.Println("data not valid,check please!")
		return
	}
	////////////////////////////////////////////////////////////
	// 生成代码 TODO: 并发生成
	for _, lang := range langCaches {
		if !lang.HasSetFlag() {
			continue
		}
		cc := &gen.ExportOption{
			Outpath:    lang.OutpathType(),
			TypePath:   lang.OutpathType(),
			DataPath:   lang.OutpathData(),
			ExportFlag: lang.ExportFlag(),
		}
		// 单项类型导出
		if lang.Opts.ExportDefine != nil {
			for _, data := range GlobalCache.AllTables {
				err = lang.Opts.ExportDefine(data, cc)
				if err != nil {
					log.Println("sheet", data.SheetName, " export language ", lang.Language, " define failed")
					for _, v := range multierr.Errors(err) {
						log.Println("\t", v)
					}
					errs = multierr.Append(errs, err)
				}

			}
		}
		// 合并类型导出
		if lang.Opts.ExportMergeDefine != nil {
			err = lang.Opts.ExportMergeDefine(GlobalCache.AllTables, cc)
			if err != nil {
				log.Println("language ", lang.Language, " export merge type failed")
				for _, v := range multierr.Errors(err) {
					log.Println("\t", v)
				}
				errs = multierr.Append(errs, err)
			}
		}
	}
	// 生成数据
	for _, lang := range langCaches {
		if !lang.HasSetFlag() {
			continue
		}
		cc := &gen.ExportOption{
			Outpath:    lang.OutpathData(),
			TypePath:   lang.OutpathType(),
			DataPath:   lang.OutpathData(),
			ExportFlag: lang.ExportFlag(),
		}
		// 单项类型导出
		if lang.Opts.ExportData != nil {
			for _, data := range GlobalCache.AllTables {
				err = lang.Opts.ExportData(data, cc)
				if err != nil {
					log.Println("sheet", data.SheetName, " export language ", lang.Language, " define failed")
					for _, v := range multierr.Errors(err) {
						log.Println("\t", v)
					}
					errs = multierr.Append(errs, err)
				}

			}
		}
		// 合并类型导出
		if lang.Opts.ExportMergeData != nil {
			err = lang.Opts.ExportMergeData(GlobalCache.AllTables, cc)
			if err != nil {
				log.Println("language ", lang.Language, " export merge type failed")
				for _, v := range multierr.Errors(err) {
					log.Println("\t", v)
				}
				errs = multierr.Append(errs, err)
			}
		}
	}
	if errs != nil {
		log.Println("xlsx files export failed")
		return
	}
	log.Println("generate finish")
}

// 解析目录中的xlsx文件
func parseDir(dir string) (errs error) {
	dir = filepath.Clean(dir)
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Println("read path failed", err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// 遍历多层目录
		if !cfg.Recursion {
			if filepath.Join(dir, info.Name()) != path {
				return nil
			}
		}

		// 解析文件
		err = parseFile(path)
		if err != nil {
			errs = multierr.Append(errs, err)
			return nil
		}

		return nil
	})
	return
}

// 解析具体的xlsx文件
func parseFile(fname string) (err error) {
	// 文件名以.开头,隐藏的文件,忽略
	if strings.HasPrefix(filepath.Base(fname), ".") {
		return nil
	}
	// 文件名后缀匹配
	if cfg.FileExtern != "" && !strings.HasSuffix(filepath.Ext(fname), cfg.FileExtern) {
		return nil
	}
	// 解析配置
	datas, checks, err := parser.LoadXlsx(fname)
	if err != nil {
		return
	}
	// 缓存信息
	GlobalCache.AllTables = append(GlobalCache.AllTables, datas...)
	GlobalCache.AllChecks = append(GlobalCache.AllChecks, checks...)
	return
}

var GlobalCache = struct {
	AllTables []*parser.XlsxSheet
	AllChecks []*parser.XlsxCheckSheet
}{}

func isDir(fname string) (bool, error) {
	file, err := os.Stat(fname)
	if err != nil {
		return false, err
	}
	return file.IsDir(), nil
}

func fixConfig() (errs error) {
	if cfg.FileExtern != "" && !strings.HasPrefix(cfg.FileExtern, ".") {
		cfg.FileExtern = "." + cfg.FileExtern
	}

	for _, lang := range langCaches {
		if !lang.HasSetFlag() {
			continue
		}
		err := lang.Opts.CheckOptions()
		if err != nil {
			log.Println("language ", lang.Language, " check options failed", err)
			errs = multierr.Append(errs, err)
		}
	}
	return
}
