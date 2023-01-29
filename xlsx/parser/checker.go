package parser

import (
	"errors"
	"fmt"
	"math"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

var (
	checkLState = lua.NewState()
	checkLock   sync.Mutex
)

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

func checkNumber(in interface{}, f func(val float64) error) (err error) {
	switch val := in.(type) {
	case int64:
		return f(float64(val))
	case uint64:
		return f(float64(val))
	case float64:
		return f(float64(val))
	case []interface{}:
		for _, av := range val {
			err = checkNumber(av, f)
			if err != nil {
				return err
			}
		}
	case map[interface{}]interface{}:
		for _, av := range val {
			err = checkNumber(av, f)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%w %#v", ErrCheckTypeInvalid, in)
	}
	return nil
}

// LuaCheckFuncOption 有指定数据的检测
func LuaCheckFuncOption(L *lua.LState) int {
	if L.GetTop() < 1 {
		return 0
	}

	var vals []float64
	for i := 1; i <= L.GetTop(); i++ {
		vals = append(vals, float64(L.CheckNumber(i)))
	}

	checker := func(in interface{}) (err error) {
		return checkNumber(in, func(val float64) error {
			for _, v := range vals {
				if v == val {
					return nil
				}
			}
			return fmt.Errorf("value %v not in option list %v", in, vals)
		})
	}

	list := LuaHelperGetCheckList(L)
	*list = append(*list, checker)
	return 0
}

// LuaCheckFuncMax 检测最大值上限
func LuaCheckFuncMax(L *lua.LState) int {
	if L.GetTop() < 1 {
		L.RaiseError("need 1 args")
		return 0
	}
	max := float64(L.CheckNumber(1))

	checker := func(in interface{}) (err error) {
		return checkNumber(in, func(val float64) error {
			if val > max {
				return fmt.Errorf("max value limit %f,got %f", max, val)
			}
			return nil
		})
	}

	list := LuaHelperGetCheckList(L)
	*list = append(*list, checker)
	return 0
}

// LuaCheckFuncMin 检测最小值下限
func LuaCheckFuncMin(L *lua.LState) int {
	if L.GetTop() < 1 {
		L.RaiseError("need 1 args")
		return 0
	}
	min := float64(L.CheckNumber(1))

	checker := func(in interface{}) (err error) {
		return checkNumber(in, func(val float64) error {
			if val < min {
				return fmt.Errorf("min value limit %f,got %f", min, val)
			}
			return nil
		})
	}

	list := LuaHelperGetCheckList(L)
	*list = append(*list, checker)
	return 0
}

// LuaCheckFuncRange 检测数值有效范围
func LuaCheckFuncRange(L *lua.LState) int {
	if L.GetTop() != 2 {
		L.RaiseError("need 2 args")
		return 0
	}
	min := float64(L.CheckNumber(1))
	max := float64(L.CheckNumber(2))

	if min > max {
		min, max = max, min
	}

	checker := func(in interface{}) (err error) {
		return checkNumber(in, func(val float64) error {
			if val < min || val > max {
				return fmt.Errorf("value range(%f,%f),got %f", min, max, val)
			}
			return nil
		})
	}

	list := LuaHelperGetCheckList(L)
	*list = append(*list, checker)
	return 0
}

func innerIntCheck(min, max int64) ValueCheck {
	if min > max {
		min, max = max, min
	}
	return func(val interface{}) (err error) {
		return checkNumber(val, func(val float64) error {
			if int64(val) < min || int64(val) > max {
				return fmt.Errorf("value range(%d,%d),got %f", min, max, val)
			}
			return nil
		})
	}
}

func innerUintCheck(max uint64) ValueCheck {
	return func(val interface{}) (err error) {
		return checkNumber(val, func(val float64) error {
			if uint64(val) > max {
				return fmt.Errorf("value range(0,%d),got %d", max, uint64(val))
			}
			return nil
		})
	}
}

func innerFload32Check() ValueCheck {
	return func(val interface{}) (err error) {
		return checkNumber(val, func(val float64) error {
			if val > math.MaxFloat32 {
				return fmt.Errorf("value out of range float32,got %f", val)
			}
			return nil
		})
	}
}

func RegisterCheck(name string, f lua.LGFunction) {
	checkLState.SetGlobal(name, checkLState.NewFunction(f))
}

func RegisterDefaultChecker() {
	RegisterCheck("range", LuaCheckFuncRange)
	RegisterCheck("values", LuaCheckFuncOption)
	RegisterCheck("max", LuaCheckFuncMax)
	RegisterCheck("min", LuaCheckFuncMin)
	//min
	// size(5,1) // 最大5个 最少1个
	// atleast(5)
}

func ParseCheker(check string) (chekers []ValueCheck, err error) {
	if len(check) < 1 {
		return
	}
	// 并发保护
	checkLock.Lock()
	list := checkLState.NewUserData()
	list.Value = &chekers
	checkLState.SetGlobal(checkListName, list)
	err = checkLState.DoString(check)
	checkLock.Unlock()
	if err != nil {
		return nil, err
	}
	return chekers, err
}
