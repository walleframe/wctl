package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.uber.org/multierr"
)

// Type xlsx支持的类型接口
type Type interface {
	// 解析数据
	Parse(src string) (replace string, val interface{}, err error)
	// 类型自带的检测
	Checkers() []ValueCheck
	// 内置类型名
	Name() string
}

type ArrayType interface {
	Element() Type
}

type MapType interface {
	Key() Type
	Value() Type
}

// String 布尔值类型
type String struct{}

var _ Type = (*String)(nil)

func (*String) Parse(src string) (replace string, val interface{}, err error) {
	return src, src, nil
}

// 类型自带的检测
func (*String) Checkers() []ValueCheck {
	return nil
}

// 内置类型名
func (*String) Name() string {
	return "string"
}

func ParseBool(str string) (bool, error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "": // 空字符串,默认false
		return false, nil
	}
	return false, &strconv.NumError{Func: "ParseBool", Num: str, Err: strconv.ErrSyntax}
}

// FormatBool returns "true" or "false" according to the value of b.
func FormatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Boolean 布尔值类型
type Boolean struct{}

var _ Type = (*Boolean)(nil)

func (*Boolean) Parse(src string) (replace string, val interface{}, err error) {
	v, err := ParseBool(src)
	if err != nil {
		return "", nil, err
	}
	return FormatBool(v), v, nil
}

// 类型自带的检测
func (*Boolean) Checkers() []ValueCheck {
	return nil
}

// 内置类型名
func (*Boolean) Name() string {
	return "bool"
}

// Int 数值类型
type Int struct {
	typ      string
	checkers []ValueCheck
}

var _ Type = (*Int)(nil)

func NewInt(name string, cheker ...ValueCheck) *Int {
	return &Int{typ: name, checkers: cheker}
}

func (*Int) Parse(src string) (replace string, val interface{}, err error) {
	if src == "" {
		src = "0"
	}
	v, err := strconv.ParseInt(src, 10, 64)
	if err != nil {
		return "", nil, err
	}
	return src, v, nil
}

// 类型自带的检测
func (t *Int) Checkers() []ValueCheck {
	return t.checkers
}

// 内置类型名
func (t *Int) Name() string {
	return t.typ
}

// Uint 数值类型
type Uint struct {
	typ      string
	checkers []ValueCheck
}

var _ Type = (*Uint)(nil)

func NewUint(name string, cheker ...ValueCheck) *Uint {
	return &Uint{typ: name, checkers: cheker}
}

func (*Uint) Parse(src string) (replace string, val interface{}, err error) {
	if src == "" {
		src = "0"
	}
	v, err := strconv.ParseUint(src, 10, 64)
	if err != nil {
		return "", nil, err
	}
	return src, v, nil
}

// 类型自带的检测
func (t *Uint) Checkers() []ValueCheck {
	return t.checkers
}

// 内置类型名
func (t *Uint) Name() string {
	return t.typ
}

// Float 数值类型
type Float struct {
	typ      string
	checkers []ValueCheck
}

var _ Type = (*Float)(nil)

func NewFloat(name string, cheker ...ValueCheck) *Float {
	return &Float{typ: name, checkers: cheker}
}

func (*Float) Parse(src string) (replace string, val interface{}, err error) {
	if src == "" {
		src = "0"
	}
	v, err := strconv.ParseFloat(src, 64)
	if err != nil {
		return "", nil, err
	}
	return src, v, nil
}

// 类型自带的检测
func (t *Float) Checkers() []ValueCheck {
	return t.checkers
}

// 内置类型名
func (t *Float) Name() string {
	return t.typ
}

// Array 类型
type Array struct {
	element Type
}

var _ Type = (*Array)(nil)

func NewArray(def string) Type {
	// 类型匹配
	if !strings.HasPrefix(def, "array<") || !strings.HasSuffix(def, ">") {
		return nil
	}
	elt := strings.TrimSuffix(strings.TrimPrefix("array<", def), ">")
	t, ok := typeCache.basic[elt]
	if !ok {
		return nil
	}
	return &Array{element: t}
}

func (t *Array) Parse(src string) (_ string, _ interface{}, errs error) {
	splits := strings.Split(src, ",")
	vals := make([]interface{}, 0, len(splits))
	reps := make([]string, 0, len(splits))
	for _, item := range splits {
		item = strings.TrimSpace(item)
		rep, v, err := t.element.Parse(item)
		if err != nil {
			errs = multierr.Append(errs, fmt.Errorf("parse value [%s] failed,%w", item, err))
			continue
		}
		vals = append(vals, v)
		reps = append(reps, rep)
	}
	return strings.Join(reps, ","), vals, errs
}

// 类型自带的检测
func (t *Array) Checkers() []ValueCheck {
	return t.element.Checkers()
}

// 内置类型名
func (t *Array) Name() string {
	return fmt.Sprintf("array<%s>", t.element.Name())
}

// Map 类型
type Map struct {
	key   Type
	value Type
}

func NewMap(def string) Type {
	// 类型匹配
	if !strings.HasPrefix(def, "map<") || !strings.HasSuffix(def, ">") {
		return nil
	}
	// 拆分类型
	elt := strings.TrimSuffix(strings.TrimPrefix("map<", def), ">")
	list := strings.Split(elt, ",")
	if len(list) != 2 {
		return nil
	}
	t, ok := typeCache.basic[list[0]]
	if !ok {
		return nil
	}
	v, ok := typeCache.basic[list[1]]
	if !ok {
		return nil
	}
	return &Map{key: t, value: v}
}

func (t *Map) Parse(src string) (_ string, _ interface{}, errs error) {
	splits := strings.Split(src, ";")
	vals := make(map[interface{}]interface{})
	buf := strings.Builder{}
	buf.Grow(len(src))
	for k, item := range splits {
		item = strings.TrimSpace(item)
		kv := strings.Split(item, ":")
		repK, vk, err := t.key.Parse(strings.TrimSpace(kv[0]))
		if err != nil {
			errs = multierr.Append(errs, fmt.Errorf("parse %s key [%s] failed,%w", t.Name(), kv[0], err))
			continue
		}
		repV, ve, err := t.value.Parse(strings.TrimSpace(kv[1]))
		if err != nil {
			errs = multierr.Append(errs, fmt.Errorf("parse %s value [%s] failed,%w", t.Name(), kv[1], err))
			continue
		}
		vals[vk] = ve
		//
		if k > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(repK)
		buf.WriteByte(':')
		buf.WriteString(repV)
	}
	return buf.String(), vals, errs
}

// 类型自带的检测
func (t *Map) Checkers() []ValueCheck {
	return []ValueCheck{func(val interface{}) (errs error) {
		maps := val.(map[interface{}]interface{})
		for k, v := range maps {
			for _, kf := range t.key.Checkers() {
				err := kf(k)
				if err != nil {
					errs = multierr.Append(errs, fmt.Errorf("%w. %s key check failed", err, t.Name()))
				}
			}
			for _, vf := range t.value.Checkers() {
				err := vf(v)
				if err != nil {
					errs = multierr.Append(errs, fmt.Errorf("%w. %s value check failed", err, t.Name()))
				}
			}
		}
		return
	}}
}

// 内置类型名
func (t *Map) Name() string {
	return fmt.Sprintf("map<%s,%s>", t.key.Name(), t.value.Name())
}

// Vector3 自定义类型
type Vector3 struct{}

// 实际数据格式
type Vec3 []float64

var _ Type = (*Vector3)(nil)

func (*Vector3) Parse(src string) (replace string, val interface{}, err error) {
	// 可选的{}包装
	src = strings.TrimSuffix(strings.TrimPrefix(src, "{"), "}")
	if src == "" {
		return "0,0,0", Vec3{0, 0, 0}, nil
	}
	lists := strings.Split(src, ",")
	if len(lists) != 3 {
		err = fmt.Errorf("vec3 must have 3 values,%s", src)
		return
	}
	var vec3 Vec3
	for i := 0; i < 3; i++ {
		n, err := strconv.ParseFloat(lists[i], 64)
		if err != nil {
			return "", nil, fmt.Errorf("vec3 %d value %s parse failed,%w", i, lists[i], err)
		}
		vec3[0] = n
	}
	return src, replace, nil
}

// 类型自带的检测
func (*Vector3) Checkers() []ValueCheck {
	return nil
}

// 内置类型名
func (*Vector3) Name() string {
	return "vec3"
}

var typeCache = &struct {
	basic map[string]Type
	autos []func(def string) Type
}{
	basic: map[string]Type{},
	autos: []func(def string) Type{},
}

// RegisterType 注册基础类型
func RegisterType(typ Type) {
	typeCache.basic[typ.Name()] = typ
}

// RegisterAutoType 注册自动解析的类型
func RegisterAutoType(check func(def string) Type) {
	typeCache.autos = append(typeCache.autos, check)
}

func RegisterDefaultType() {
	RegisterType(&String{})
	RegisterType(&Boolean{})
	RegisterType(NewInt("int", innerIntCheck(math.MinInt32, math.MaxInt32)))
	RegisterType(NewInt("int8", innerIntCheck(math.MinInt8, math.MaxInt8)))
	RegisterType(NewInt("int16", innerIntCheck(math.MinInt16, math.MaxInt16)))
	RegisterType(NewInt("int32", innerIntCheck(math.MinInt32, math.MaxInt32)))
	RegisterType(NewInt("int64"))
	RegisterType(NewUint("uint", innerUintCheck(math.MaxUint32)))
	RegisterType(NewUint("uint8", innerUintCheck(math.MaxUint8)))
	RegisterType(NewUint("uint16", innerUintCheck(math.MaxUint16)))
	RegisterType(NewUint("uint32", innerUintCheck(math.MaxUint32)))
	RegisterType(NewUint("uint64"))
	RegisterType(NewFloat("float32", innerFload32Check()))
	RegisterType(NewFloat("float64"))
	// container type
	RegisterAutoType(NewArray)
	RegisterAutoType(NewMap)
	// 自定义类型
	RegisterType(&Vector3{})
}

// ParseType 解析类型定义
func ParseType(def string) (Type, error) {
	if typ, ok := typeCache.basic[def]; ok {
		return typ, nil
	}
	for _, check := range typeCache.autos {
		if typ := check(def); typ != nil {
			return typ, nil
		}
	}
	return nil, fmt.Errorf("unkown type %s,check please", def)
}
