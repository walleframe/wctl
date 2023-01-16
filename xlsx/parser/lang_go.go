package parser

import "fmt"

// GoSupport 支持go语言接口
type GoSupport interface {
	GoTypeName() (string, error)
}

func (*String) GoTypeName() (string, error) {
	return "string", nil
}

func (*Boolean) GoTypeName() (string, error) {
	return "bool", nil
}

func (t *Int) GoTypeName() (string, error) {
	// int 默认是 int32
	if t.typ == "int" {
		return "int32", nil
	}
	return t.typ, nil
}
func (t *Uint) GoTypeName() (string, error) {
	// int 默认是 int32
	if t.typ == "uint" {
		return "uint32", nil
	}
	return t.typ, nil
}

func (t *Float) GoTypeName() (string, error) {
	return t.typ, nil
}

func (t *Array) GoTypeName() (string, error) {
	if gt, ok := t.element.(GoSupport); ok {
		typ, err := gt.GoTypeName()
		if err != nil {
			return "", fmt.Errorf("convert %s element type to golang failed,%w", t.Name(), err)
		}
		return fmt.Sprintf("[]%s", typ), nil
	}
	return "", fmt.Errorf("%s element not support golang", t.Name())
}

func (t *Map) GoTypeName() (string, error) {
	var key, value string
	if gt, ok := t.key.(GoSupport); ok {
		typ, err := gt.GoTypeName()
		if err != nil {
			return "", fmt.Errorf("convert %s key type to golang failed,%w", t.Name(), err)
		}
		key = typ
	}
	if gt, ok := t.value.(GoSupport); ok {
		typ, err := gt.GoTypeName()
		if err != nil {
			return "", fmt.Errorf("convert %s value type to golang failed,%w", t.Name(), err)
		}
		value = typ
	}

	return fmt.Sprintf("map[%s]%s", key, value), nil
}

// 暂不解析此字段 
func (*Vector3) GoTypeName() string {
	return "string"
}
