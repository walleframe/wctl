package buildpb

import (
	"errors"
	"strings"

	"github.com/aggronmagi/wctl/utils"
)

func (x *OptionDesc) getOpt(opt string) (val *OptionValue) {
	if x == nil || x.Options == nil || len(x.Options) < 1 {
		return nil
	}
	val, _ = x.Options[opt]
	return
}

func (x *OptionDesc) HasOption(opt string) (ok bool) {
	return x.getOpt(opt) != nil
}
func (x *OptionDesc) GetStringCheck(opt string) (val string, ok bool) {
	v := x.getOpt(opt)
	if v == nil {
		return
	}

	if v.IntValue > 0 {
		return
	}

	if v.Value != "" {
		val = v.Value
		ok = true
	}

	return
}
func (x *OptionDesc) GetString(opt, def string) string {
	v := x.getOpt(opt)
	if v == nil {
		return def
	}

	return v.Value
}
func (x *OptionDesc) GetStringSlice(opt, sep string, def ...string) (slice []string, ok bool) {
	v := x.getOpt(opt)
	if v == nil {
		slice = def
		return
	}
	if v.IntValue > 0 {
		slice = def
		return
	}

	slice = strings.Split(v.Value, sep)
	ok = len(slice) > 0
	return
}
func (x *OptionDesc) GetIntCheck(opt string) (val int64, ok bool) {
	v := x.getOpt(opt)
	if v == nil {
		return
	}
	if v.Value != "" {
		return
	}
	ok = true
	val = v.IntValue
	return
}
func (x *OptionDesc) GetInt64(opt string, def int64) (val int64) {
	v := x.getOpt(opt)
	if v == nil {
		return def
	}
	return v.IntValue
}

func (x *FileDesc) HasOption(opt string) (ok bool) {
	return x.Options.HasOption(opt)
}
func (x *FileDesc) GetStringCheck(opt string) (val string, ok bool) {
	return x.Options.GetStringCheck(opt)
}
func (x *FileDesc) GetString(opt, def string) string {
	return x.Options.GetString(opt, def)
}
func (x *FileDesc) GetStringSlice(opt, sep string, def ...string) (slice []string, ok bool) {
	return x.Options.GetStringSlice(opt, sep, def...)
}
func (x *FileDesc) GetIntCheck(opt string) (val int64, ok bool) {
	return x.Options.GetIntCheck(opt)
}
func (x *FileDesc) GetInt64(opt string, def int64) (val int64) {
	return x.Options.GetInt64(opt, def)
}

func (x *MsgDesc) HasOption(opt string) (ok bool) {
	return x.Options.HasOption(opt)
}
func (x *MsgDesc) GetStringCheck(opt string) (val string, ok bool) {
	return x.Options.GetStringCheck(opt)
}
func (x *MsgDesc) GetString(opt, def string) string {
	return x.Options.GetString(opt, def)
}
func (x *MsgDesc) GetStringSlice(opt, sep string, def ...string) (slice []string, ok bool) {
	return x.Options.GetStringSlice(opt, sep, def...)
}
func (x *MsgDesc) GetIntCheck(opt string) (val int64, ok bool) {
	return x.Options.GetIntCheck(opt)
}
func (x *MsgDesc) GetInt64(opt string, def int64) (val int64) {
	return x.Options.GetInt64(opt, def)
}

func (x *Field) HasOption(opt string) (ok bool) {
	return x.Options.HasOption(opt)
}
func (x *Field) GetStringCheck(opt string) (val string, ok bool) {
	return x.Options.GetStringCheck(opt)
}
func (x *Field) GetString(opt, def string) string {
	return x.Options.GetString(opt, def)
}
func (x *Field) GetStringSlice(opt, sep string, def ...string) (slice []string, ok bool) {
	return x.Options.GetStringSlice(opt, sep, def...)
}
func (x *Field) GetIntCheck(opt string) (val int64, ok bool) {
	return x.Options.GetIntCheck(opt)
}
func (x *Field) GetInt64(opt string, def int64) (val int64) {
	return x.Options.GetInt64(opt, def)
}

func (x *MethodDesc) HasOption(opt string) (ok bool) {
	return x.Options.HasOption(opt)
}
func (x *MethodDesc) GetStringCheck(opt string) (val string, ok bool) {
	return x.Options.GetStringCheck(opt)
}
func (x *MethodDesc) GetString(opt, def string) string {
	return x.Options.GetString(opt, def)
}
func (x *MethodDesc) GetStringSlice(opt, sep string, def ...string) (slice []string, ok bool) {
	return x.Options.GetStringSlice(opt, sep, def...)
}
func (x *MethodDesc) GetIntCheck(opt string) (val int64, ok bool) {
	return x.Options.GetIntCheck(opt)
}
func (x *MethodDesc) GetInt64(opt string, def int64) (val int64) {
	return x.Options.GetInt64(opt, def)
}

func (x *ServiceDesc) HasOption(opt string) (ok bool) {
	return x.Options.HasOption(opt)
}
func (x *ServiceDesc) GetStringCheck(opt string) (val string, ok bool) {
	return x.Options.GetStringCheck(opt)
}
func (x *ServiceDesc) GetString(opt, def string) string {
	return x.Options.GetString(opt, def)
}
func (x *ServiceDesc) GetStringSlice(opt, sep string, def ...string) (slice []string, ok bool) {
	return x.Options.GetStringSlice(opt, sep, def...)
}
func (x *ServiceDesc) GetIntCheck(opt string) (val int64, ok bool) {
	return x.Options.GetIntCheck(opt)
}
func (x *ServiceDesc) GetInt64(opt string, def int64) (val int64) {
	return x.Options.GetInt64(opt, def)
}

func (x *Field) GoType() (v string, err error) {
	switch x.Type.Type {
	case FieldType_BaseType:
		if x.Type.KeyBase == BaseTypeDesc_Binary {
			v = "[]byte"
			return
		}
		v = x.Type.Key
	case FieldType_CustomType:
		v = "*" + custom(x.Type.Key)
	case FieldType_ListType:
		if x.Type.ElemCustom {
			v = "[]*" + custom(x.Type.Key)
		} else {
			if x.Type.KeyBase == BaseTypeDesc_Binary {
				v = "[][]byte"
			} else {
				v = "[]" + x.Type.Key
			}
		}
	case FieldType_MapType:
		// map 的可以不能使用[]byte
		if x.Type.KeyBase == BaseTypeDesc_Binary {
			err = errors.New("map key can't set binary type.")
			return
		}
		if x.Type.ElemCustom {
			v = "map[" + x.Type.Key + "]*" + custom(x.Type.Value)
		} else {
			if x.Type.ValueBase == BaseTypeDesc_Binary {
				v = "map[" + x.Type.Key + "][]byte"
			} else {
				v = "map[" + x.Type.Key + "]" + x.Type.Value
			}
		}
	default:
		err = errors.New("invalid go type")
	}
	return
}

func (x *Field) IsList() bool {
	return x.Type.Type == FieldType_ListType
}

func (x *Field) IsMap() bool {
	return x.Type.Type == FieldType_MapType
}

func (x *Field) IsCustom() bool {
	return x.Type.Type == FieldType_CustomType
}

func (x *Field) IsBasicType() bool {
	return x.Type.Type == FieldType_BaseType
}

func (x *Field) ContainCustom() bool {
	return x.Type.Msg != nil
}

func (x Field) CustomMsg() *MsgDesc {
	return x.Type.Msg
}

func custom(typ string) (v string) {
	list := strings.Split(typ, ".")
	switch len(list) {
	case 2:
		v = list[0] + "." + utils.Title(list[1])
	case 1:
		fallthrough
	default:
		v = utils.Title(typ)
	}
	return
}

func (x *MsgDesc) ContainCustom() bool {
	for _, v := range x.Fields {
		if v.ContainCustom() {
			return true
		}
	}
	return false
}
