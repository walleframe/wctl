package parser

import "fmt"

// PBSupport 支持go语言接口
type PBSupport interface {
	PBTypeName() (string, error)
}

func (*String) PBTypeName() (string, error) {
	return "string", nil
}

func (*Boolean) PBTypeName() (string, error) {
	return "bool", nil
}

func (t *Int) PBTypeName() (string, error) {
	switch t.typ {
	case "int64":
		return t.typ, nil
	default:
		return "int32", nil
	}
}
func (t *Uint) PBTypeName() (string, error) {
	// int 默认是 int32
	if t.typ == "uint64" {
		return t.typ, nil
	}
	return "uint32", nil
}

func (t *Float) PBTypeName() (string, error) {
	if t.typ == "float32" {
		return "float", nil
	}
	return "double", nil
}

func (t *Array) PBTypeName() (string, error) {
	if gt, ok := t.element.(PBSupport); ok {
		typ, err := gt.PBTypeName()
		if err != nil {
			return "", fmt.Errorf("convert %s element type to protobuf failed,%w", t.Name(), err)
		}
		return fmt.Sprintf("repeated %s", typ), nil
	}
	return "", fmt.Errorf("%s element not support protobuf", t.Name())
}

func (t *Map) PBTypeName() (string, error) {
	var key, value string
	if gt, ok := t.key.(PBSupport); ok {
		typ, err := gt.PBTypeName()
		if err != nil {
			return "", fmt.Errorf("convert %s key type to protobuf failed,%w", t.Name(), err)
		}
		key = typ
	}
	if gt, ok := t.value.(PBSupport); ok {
		typ, err := gt.PBTypeName()
		if err != nil {
			return "", fmt.Errorf("convert %s value type to protobuf failed,%w", t.Name(), err)
		}
		value = typ
	}

	return fmt.Sprintf("map<%s,%s>", key, value), nil
}
