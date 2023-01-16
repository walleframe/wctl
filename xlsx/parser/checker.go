package parser

import (
	"errors"
	"fmt"
	"math"

	lua "github.com/yuin/gopher-lua"
)

var checkLState = lua.NewState()

const (
	checkListName = "__check_list"
)

// 数据检测
type ValueCheck func(val interface{}) (err error)

var (
	ErrCheckTypeInvalid = errors.New("check type not support")
)

// LuaHelperGetCheckList lua注册辅助,获取checklist指针
func LuaHelperGetCheckList(L *lua.LState) *[]ValueCheck {
	val := L.GetGlobal(checkListName)
	if val == nil {
		L.RaiseError("get check-list failed")
		return nil
	}

	if ud, ok := val.(*lua.LUserData); !ok {
		L.RaiseError("get check-list convert LUserData failed")
		return nil
	} else if list, ok := ud.Value.(*[]ValueCheck); ok {
		return list
	}
	L.RaiseError("get check-list failed")
	return nil
}

// CheckOptionInteger 有指定数据的检测
func CheckOptionIntege(L *lua.LState) int {
	if L.GetTop() < 1 {
		return 0
	}

	var vals []int64
	for i := 1; i <= L.GetTop(); i++ {
		vals = append(vals, L.CheckInt64(i))
	}
	// TODO: L.CheckNumber(1) float支持
	checker := func(in interface{}) (err error) {
		switch n := in.(type) {
		case int64:
			for _, v := range vals {
				if v == n {
					return nil
				}
			}
		case uint64:
			// 此处转换成int64,应该可以满足需求,不考虑溢出情况
			for _, v := range vals {
				if v == int64(n) {
					return nil
				}
			}
		case float64:
			// TODO: float 支持
		case []interface{}: // slice 支持
		case map[interface{}]interface{}: // map - value 检测支持
		default:
			return ErrCheckTypeInvalid
		}

		return fmt.Errorf("value %v not in option list %v", in, vals)
	}

	list := LuaHelperGetCheckList(L)
	*list = append(*list, checker)
	return 0
}

// CheckMaxInteger 检测最大值上限
func CheckMaxInteger(L *lua.LState) int {
	if L.GetTop() < 1 {
		L.RaiseError("need 1 args")
		return 0
	}
	max := L.CheckInt64(1)

	checker := func(val interface{}) (err error) {
		n, ok := val.(int64)
		if !ok {
			return ErrCheckTypeInvalid
		}
		if n > max {
			return fmt.Errorf("max value limit %d,got %d", max, n)
		}
		return nil
	}
	list := LuaHelperGetCheckList(L)
	*list = append(*list, checker)
	return 0
}

// CheckIntegerRange 检测数值有效范围
func CheckIntegerRange(L *lua.LState) int {

	if L.GetTop() != 2 {
		L.RaiseError("need 2 args")
		return 0
	}
	min := L.CheckInt64(1)
	max := L.CheckInt64(2)

	if min > max {
		min, max = max, min
	}

	list := LuaHelperGetCheckList(L)

	checker := func(val interface{}) (err error) {
		n, ok := val.(int64)
		if !ok {
			return ErrCheckTypeInvalid
		}

		if n < min || n > max {
			return fmt.Errorf("value range(%d,%d),got %d", min, max, n)
		}
		return nil
	}
	*list = append(*list, checker)
	return 0
}

func innerIntCheck(min, max int64) ValueCheck {
	if min > max {
		min, max = max, min
	}
	return func(val interface{}) (err error) {
		n, ok := val.(int64)
		if !ok {
			return ErrCheckTypeInvalid
		}

		if n < min || n > max {
			return fmt.Errorf("value range(%d,%d),got %d", min, max, n)
		}
		return nil
	}
}

func innerUintCheck(max uint64) ValueCheck {
	return func(val interface{}) (err error) {
		n, ok := val.(uint64)
		if !ok {
			return ErrCheckTypeInvalid
		}

		if n > max {
			return fmt.Errorf("value range(0,%d),got %d", max, n)
		}
		return nil
	}
}

func innerFload32Check() ValueCheck {
	return func(val interface{}) (err error) {
		n, ok := val.(float64)
		if !ok {
			return ErrCheckTypeInvalid
		}

		if n < math.SmallestNonzeroFloat32 || n > math.MaxFloat32 {
			return fmt.Errorf("value out of range float32,got %f", n)
		}
		return nil
	}
}


func RegisterCheck(name string, f lua.LGFunction) {
	checkLState.SetGlobal(name, checkLState.NewFunction(f))
}

func RegisterDefaultChecker() {
	RegisterCheck("range", CheckIntegerRange)
	RegisterCheck("option", CheckOptionIntege)
	RegisterCheck("max", CheckMaxInteger)
	//min
	// size(5,1) // 最大5个 最少1个
	// atleast(5)
}

func ParseCheker(check string) (chekers []ValueCheck, err error) {
	if len(check) < 1 {
		return
	}
	list := checkLState.NewUserData()
	list.Value = &chekers
	checkLState.SetGlobal(checkListName, list)
	err = checkLState.DoString(check)
	if err != nil {
		return nil, err
	}
	return chekers, err
}
