package ast

import "github.com/aggronmagi/wctl/builder/buildpb"

// GetFileDesc 获取文件描述
func (prog *YTProgram) GetFileDesc() *buildpb.FileDesc {
	prog.buildFileDesc()
	return prog.desc
}

// GetFileDescWithImports 获取文件描述(包含依赖)
func (prog *YTProgram) GetFileDescWithImports() (list []*buildpb.FileDesc) {
	list = append(list, prog.GetFileDesc())
	for _, v := range prog.Imports {
		list = append(list, v.Prog.GetFileDesc())
	}
	return
}

func (prog *YTProgram) buildFileDesc() {
	if prog.desc != nil {
		return
	}
	desc := &buildpb.FileDesc{}

	desc.File = prog.File
	desc.Options = prog.YTOptions.toDesc()
	for _, v := range prog.Imports {
		desc.Imports = append(desc.Imports, v.toDesc())
	}
	for _, v := range prog.EnumDefs {
		desc.Enums = append(desc.Enums, v.toDesc())
	}
	for _, v := range prog.Messages {
		desc.Msgs = append(desc.Msgs, v.toDesc())
	}

	for _, v := range prog.Services {
		desc.Services = append(desc.Services, v.toDesc())
	}

	for _, v := range prog.Projects {
		desc.Projects = append(desc.Projects, v.toDesc())
	}
	desc.Pkg = prog.Pkg.toDesc()

	prog.desc = desc
	return
}

func (pkg *YTPackage) toDesc() (desc *buildpb.PackageDesc) {
	desc = &buildpb.PackageDesc{}
	desc.Doc = pkg.YTDoc.toDesc()
	desc.Package = pkg.Name
	return
}

func (proj *YTProject) toDesc() (desc *buildpb.ProjectDesc) {
	desc = &buildpb.ProjectDesc{}
	desc.Doc = proj.YTDoc.toDesc()
	desc.Name = proj.Name
	for k, v := range proj.Conf {
		desc.Conf[k] = v.toDesc()
	}
	return
}

func (method *YTMethod) toDesc() (desc *buildpb.MethodDesc) {
	desc = &buildpb.MethodDesc{}
	desc.Doc = method.YTDoc.toDesc()
	desc.Name = method.Name
	if method.No != nil {
		desc.MethodID = *method.No.Value
	}
	desc.Options = method.YTOptions.toDesc()
	desc.Request = method.Request.toDesc()
	desc.Reply = method.Reply.toDesc()
	desc.MethodFlag = int32(method.Flag)
	return
}

func (service *YTService) toDesc() (desc *buildpb.ServiceDesc) {
	desc = &buildpb.ServiceDesc{}
	desc.Doc = service.YTDoc.toDesc()
	desc.Name = service.Name
	desc.Options = service.YTOptions.toDesc()
	for _, v := range service.Methods {
		desc.Methods = append(desc.Methods, v.toDesc())
	}
	return
}

func (typ *YTFieldType) toDesc() (desc *buildpb.TypeDesc) {
	desc = &buildpb.TypeDesc{}
	switch {
	case typ.YTBaseType != nil:
		desc.Type = buildpb.FieldType_BaseType
		desc.KeyBase = buildpb.BaseTypeDesc(*typ.YTBaseType)
		desc.Key = typ.YTBaseType.String()
	case typ.YTCustomType != nil:
		desc.Type = buildpb.FieldType_CustomType
		desc.Key = typ.YTCustomType.Name
		desc.Msg = typ.Msg.toDesc()
		desc.ElemCustom = true
	case typ.YTListType != nil:
		desc.Type = buildpb.FieldType_ListType
		if typ.YTListType.YTBaseType != nil {
			desc.ElemCustom = false
			desc.Key = typ.YTListType.YTBaseType.String()
			desc.KeyBase = buildpb.BaseTypeDesc(*typ.YTListType.YTBaseType)
		} else if typ.YTListType.YTCustomType != nil {
			desc.ElemCustom = true
			desc.Key = typ.YTListType.YTCustomType.Name
			desc.Msg = typ.YTListType.YTCustomType.Msg.toDesc()
		}
	case typ.YTMapTypee != nil:
		desc.Type = buildpb.FieldType_MapType
		desc.Key = typ.YTMapTypee.Key.String()
		desc.KeyBase = buildpb.BaseTypeDesc(*typ.YTMapTypee.Key)

		if typ.YTMapTypee.Value.YTBaseType != nil {
			desc.ElemCustom = false
			desc.Value = typ.YTMapTypee.Value.YTBaseType.String()
			desc.ValueBase = buildpb.BaseTypeDesc(*typ.YTMapTypee.Value.YTBaseType)
		} else if typ.YTMapTypee.Value.YTCustomType != nil {
			desc.ElemCustom = true
			desc.Value = typ.YTMapTypee.Value.YTCustomType.Name
			desc.Msg = typ.YTMapTypee.Value.YTCustomType.Msg.toDesc()
		}
	}
	return
}

func (field *YTField) toDesc() (desc *buildpb.Field) {
	desc = &buildpb.Field{}
	desc.Doc = field.YTDoc.toDesc()
	desc.Name = field.Name
	desc.No = int32(field.No)
	desc.Options = field.YTOptions.toDesc()
	desc.Type = field.Type.toDesc()
	return
}

func (msg *YTMessage) toDesc() (desc *buildpb.MsgDesc) {
	if msg == nil {
		return nil
	}
	desc = &buildpb.MsgDesc{}
	desc.Doc = msg.YTDoc.toDesc()
	desc.Name = msg.Name
	desc.Options = msg.YTOptions.toDesc()
	for _, v := range msg.Fields {
		desc.Fields = append(desc.Fields, v.toDesc())
	}
	// 子消息
	for _, sub := range msg.SubMsgs {
		desc.SubMsgs = append(desc.SubMsgs, sub.toDesc())
	}
	return
}

func (enum *YTEnumDef) toDesc() (desc *buildpb.EnumDesc) {
	desc = &buildpb.EnumDesc{}
	desc.Doc = enum.YTDoc.toDesc()
	desc.Name = enum.Name
	desc.Options = enum.YTOptions.toDesc()
	for _, v := range enum.Values {
		val := &buildpb.EnumValue{}
		val.Doc = v.YTDoc.toDesc()
		val.Name = v.Name
		val.Value = v.Value
		desc.Values = append(desc.Values, val)
	}
	return
}

func (imp *YTImport) toDesc() (desc *buildpb.ImportDesc) {
	desc = &buildpb.ImportDesc{}
	desc.Alias = imp.AliasName
	desc.File = imp.File
	desc.Doc = imp.YTDoc.toDesc()
	return
}

func (doc *YTDoc) toDesc() (desc *buildpb.DocDesc) {
	if doc == nil {
		return
	}
	desc = &buildpb.DocDesc{Doc: doc.Doc, TailDoc: doc.TailDoc}
	return
}

func (opts *YTOptions) toDesc() (desc *buildpb.OptionDesc) {
	if opts == nil {
		return
	}
	desc = &buildpb.OptionDesc{}
	desc.Options = make(map[string]*buildpb.OptionValue, len(opts.Opts))
	for _, v := range opts.Opts {
		val := &buildpb.OptionValue{}
		val.Doc = v.YTDoc.toDesc()
		if v.Value != nil {
			if v.Value.Value != nil {
				val.Value = *v.Value.Value
			} else if v.Value.IntVal != nil {
				val.IntValue = *v.Value.IntVal
			}
		}
		desc.Options[v.Key] = val
	}
	return
}
