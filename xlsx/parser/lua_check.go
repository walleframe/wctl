package parser

import (
	"errors"
	"log"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"go.uber.org/multierr"
)

func PrepareCheckTable(errs *[]error) *lua.LState {
	l := lua.NewState()
	var lock sync.Mutex
	l.SetGlobal("error", l.NewFunction(func(l *lua.LState) int {
		buf := strings.Builder{}
		for i := 1; i <= l.GetTop(); i++ {
			if i > 1 {
				buf.WriteByte(' ')
			}
			buf.WriteString(l.ToString(i))
		}
		lock.Lock()
		*errs = append(*errs, errors.New(buf.String()))
		lock.Unlock()
		return 0
	}))
	return l
}

func SetLuaCheckTable(L *lua.LState, table *XlsxSheet) (err error) {
	ud := L.NewUserData()
	ud.Value = table
	ltb := L.NewTable()
	L.SetGlobal(table.SheetName, ltb)
	for k, field := range table.AllType {
		col := k
		mt := L.NewTable()
		L.SetField(mt, "range", L.NewClosure(luaFuncClosureRange, ud, lua.LNumber(col)))
		L.SetField(mt, "find", L.NewClosure(luaFuncClosureFind, ud, lua.LNumber(col)))
		L.SetField(mt, "get", L.NewClosure(luaFuncClosureGet, ud, lua.LNumber(col)))
		L.SetField(ltb, field.Name, mt)
	}
	return

}

func LuaCheckTable(L *lua.LState, check *XlsxCheckSheet, innerErrors *[]error) (errs error) {
	for lable, script := range check.LuaScripts {
		err := L.DoString(script)
		if err != nil {
			log.Println("lua script invalid,check your script")
			log.Printf("%s:%#v\n", lable, err)
			log.Println(script)

			log.Println("check failed", err)
			_ = lable
			errs = multierr.Append(errs, errors.New(lable))
			continue
		}
		if len(*innerErrors) > 0 {
			log.Println("check ", lable, " failed")
			for _, v := range *innerErrors {
				log.Println("\t", v)
			}
			*innerErrors = (*innerErrors)[:0]
			errs = multierr.Append(errs, errors.New(lable))
		}
	}
	return
}

func luaFuncClosureRange(l *lua.LState) int {
	// 函数参数
	rf := l.CheckFunction(1)
	// upvalue
	ud := l.CheckUserData(lua.UpvalueIndex(1))
	col := l.CheckInt(lua.UpvalueIndex(2))
	sheet := ud.Value.(*XlsxSheet)
	l.SetTop(0)
	// 遍历数据
	for i := 0; i < len(sheet.AllData); i++ {
		// call rf(cell-raw-data, sheet-user-data, row, col)
		l.Push(rf)
		l.Push(lua.LString(sheet.AllData[i][col].Raw))
		//l.Push(ud)
		//l.Push(lua.LNumber(i))
		//l.Push(lua.LNumber(col))
		err := l.PCall(1, 0, nil)
		if err != nil {
			log.Println("exec error,", err)
		}
		l.SetTop(0)
	}
	return 0
}

func luaFuncClosureFind(l *lua.LState) int {
	ud := l.CheckUserData(lua.UpvalueIndex(1))
	col := l.CheckInt(lua.UpvalueIndex(2))
	sheet := ud.Value.(*XlsxSheet)

	switch l.GetTop() {
	case 1:
		// 对比原始数据
		value := l.ToString(1)
		find := false
		for i := 0; i < len(sheet.AllData); i++ {
			if value == sheet.AllData[i][col].Raw {
				find = true
				break
			}
		}
		l.Push(lua.LBool(find))
		return 1
	}
	return 0
}

func luaFuncClosureGet(l *lua.LState) int {
	ud := l.CheckUserData(lua.UpvalueIndex(1))
	col := l.CheckInt(lua.UpvalueIndex(2))
	sheet := ud.Value.(*XlsxSheet)

	row := l.CheckInt(1)

	if row >= len(sheet.AllData) {
		l.Push(lua.LString(sheet.AllData[row][col].Raw))
	} else {
		l.Push(lua.LNil)
	}
	return 1
}
