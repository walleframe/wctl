// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n Option -o options.go"
// Version: 0.0.2

package gen

var _ = generateOptions()

type Options struct {
	// 缩进
	Indent string
	// Go格式化
	GoFmt bool
	// key 是否大写
	KeyTitle bool
}

// 缩进
func WithIndent(v string) Option {
	return func(cc *Options) Option {
		previous := cc.Indent
		cc.Indent = v
		return WithIndent(previous)
	}
}

// Go格式化
func WithGoFmt(v bool) Option {
	return func(cc *Options) Option {
		previous := cc.GoFmt
		cc.GoFmt = v
		return WithGoFmt(previous)
	}
}

// key 是否大写
func WithKeyTitle(v bool) Option {
	return func(cc *Options) Option {
		previous := cc.KeyTitle
		cc.KeyTitle = v
		return WithKeyTitle(previous)
	}
}

// SetOption modify options
func (cc *Options) SetOption(opt Option) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *Options) ApplyOption(opts ...Option) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *Options) GetSetOption(opt Option) Option {
	return opt(cc)
}

// Option option define
type Option func(cc *Options) Option

// NewOptions create options instance.
func NewOptions(opts ...Option) *Options {
	cc := newDefaultOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogOptions != nil {
		watchDogOptions(cc)
	}
	return cc
}

// InstallOptionsWatchDog install watch dog
func InstallOptionsWatchDog(dog func(cc *Options)) {
	watchDogOptions = dog
}

var watchDogOptions func(cc *Options)

// newDefaultOptions new option with default value
func newDefaultOptions() *Options {
	cc := &Options{
		Indent:   "\t",
		GoFmt:    false,
		KeyTitle: true,
	}
	return cc
}
