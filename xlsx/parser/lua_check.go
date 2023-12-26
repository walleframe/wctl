package parser

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"go.uber.org/multierr"
)

func PrepareCheckTable(errs *[]error) *lua.LState {
	l := lua.NewState()
	var lock sync.Mutex
	l.SetGlobal("failed", l.NewFunction(func(l *lua.LState) int {
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
	l.SetGlobal("split", l.NewFunction(func(l *lua.LState) int {
		data := l.CheckString(1)
		sep1 := l.CheckString(2)
		sep2 := ""
		if l.GetTop() > 2 {
			sep2 = l.CheckString(3)
		}
		list1 := strings.Split(data, sep1)
		if sep2 == "" {
			t := l.NewTable()
			for _, v := range list1 {
				t.Append(lua.LString(strings.TrimSpace(v)))
			}
			l.Push(t)
			return 1
		}
		t := l.NewTable()
		for _, v1 := range list1 {
			list2 := strings.Split(v1, sep2)
			t2 := l.NewTable()
			for _, v := range list2 {
				t2.Append(lua.LString(strings.TrimSpace(v)))
			}
			t.Append(t2)
		}
		l.Push(t)
		return 1
	}))
	return l
}

func SetLuaCheckTable(L *lua.LState, table *XlsxSheet) (err error) {
	ud := L.NewUserData()
	ud.Value = table
	ltb := L.NewTable()
	L.SetField(ltb, "range", L.NewClosure(luaFuncClosureMulRange, ud))
	L.SetField(ltb, "get", L.NewClosure(luaFuncClosureMulGet, ud))
	for k, field := range table.AllType {
		col := k
		mt := L.NewTable()
		L.SetField(mt, "range", L.NewClosure(luaFuncClosureRange, ud, lua.LNumber(col)))
		L.SetField(mt, "find", L.NewClosure(luaFuncClosureFind, ud, lua.LNumber(col)))
		L.SetField(mt, "get", L.NewClosure(luaFuncClosureGet, ud, lua.LNumber(col)))
		L.SetField(ltb, field.Name, mt)
	}
	L.SetGlobal(table.SheetName, ltb)
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

func luaFuncClosureMulRange(l *lua.LState) int {
	if l.GetTop() < 1 {
		l.ArgError(1, "must input function and columns type")
	}
	// 函数参数
	cols := make([]string, 0, l.GetTop()-1)
	for i := 1; i <= l.GetTop()-1; i++ {
		cols = append(cols, l.CheckString(i))
	}
	if len(cols) < 1 {
		l.RaiseError("need input columns name")
	}
	//log.Println("mul-range", cols)
	rf := l.CheckFunction(l.GetTop())
	// upvalue
	ud := l.CheckUserData(lua.UpvalueIndex(1))
	sheet := ud.Value.(*XlsxSheet)

	colIds := make([]int, 0, len(cols))
	for ai, cn := range cols {
		find := false
		for cid, ct := range sheet.AllType {
			if ct.Name == cn {
				find = true
				colIds = append(colIds, cid)
				break
			}
		}
		if !find {
			l.ArgError(ai+1, fmt.Sprintf("%s not in sheet %s", cn, sheet.SheetName))
		}
	}
	// log.Println("mul-range-ids", cols)
	l.SetTop(0)
	// 遍历数据
	for i := 0; i < len(sheet.AllData); i++ {
		// call rf(cell-raw-data, sheet-user-data, row, col)
		l.Push(rf)
		for _, cid := range colIds {
			l.Push(lua.LString(sheet.AllData[i][cid].Raw))
		}
		l.Push(lua.LNumber(i))
		//l.Push(lua.LNumber(col))
		err := l.PCall(len(colIds)+1, 0, nil)
		if err != nil {
			log.Println("exec error,", err)
		}
		l.SetTop(0)
	}
	return 0
}

func luaFuncClosureMulGet(l *lua.LState) int {
	row := l.CheckInt(1)
	// 函数参数
	cols := make([]string, 0, l.GetTop()-1)
	for i := 1; i < l.GetTop()-1; i++ {
		cols = append(cols, l.CheckString(i))
	}
	if len(cols) < 1 {
		l.RaiseError("need input columns name")
	}

	ud := l.CheckUserData(lua.UpvalueIndex(1))
	sheet := ud.Value.(*XlsxSheet)

	colIds := make([]int, 0, len(cols))
	for ai, cn := range cols {
		find := false
		for cid, ct := range sheet.AllType {
			if ct.Name == cn {
				find = true
				colIds = append(colIds, cid)
				break
			}
		}
		if !find {
			l.ArgError(ai+1, fmt.Sprintf("%s not in sheet %s", cn, sheet.SheetName))
		}
	}
	l.SetTop(0)

	if row < len(sheet.AllData) {
		for _, cid := range colIds {
			l.Push(lua.LString(sheet.AllData[row][cid].Raw))
		}
		return len(colIds)
	} else {
		l.Push(lua.LNil)
		return 1
	}
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
		l.Push(lua.LNumber(i))
		//l.Push(lua.LNumber(col))
		err := l.PCall(2, 0, nil)
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

	if row < len(sheet.AllData) {
		l.Push(lua.LString(sheet.AllData[row][col].Raw))
	} else {
		l.Push(lua.LNil)
	}
	return 1
}
