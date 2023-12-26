package gen

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/walleframe/wctl/xlsx/parser"
)

// -- string Value
type customFlagStringValue struct {
	val    *string
	root   *ExportSupportConfig
	update []func()
}

var _ pflag.Value = (*customFlagStringValue)(nil)

func (s *customFlagStringValue) Set(val string) error {
	*s.val = val
	s.root.setFlag = true
	for _, uf := range s.update {
		uf()
	}
	return nil
}
func (s *customFlagStringValue) Type() string {
	return "string"
}

func (s *customFlagStringValue) String() string { return *s.val }

// -- bool Value
type customFlagBoolValue struct {
	val    *bool
	root   *ExportSupportConfig
	update []func()
}

var _ pflag.Value = (*customFlagBoolValue)(nil)

func (s *customFlagBoolValue) Set(val string) error {
	v, err := parser.ParseBool(val)
	if err != nil {
		return err
	}
	*s.val = v
	s.root.setFlag = true
	for _, uf := range s.update {
		uf()
	}
	return nil
}
func (s *customFlagBoolValue) Type() string {
	return "bool"
}

func (s *customFlagBoolValue) String() string { return parser.FormatBool(*s.val) }

// -- int Value
type customFlagIntValue struct {
	val    *int64
	root   *ExportSupportConfig
	limits []int64
	update []func()
}

var _ pflag.Value = (*customFlagIntValue)(nil)

func (s *customFlagIntValue) Set(val string) error {
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	// 数值限制
	if len(s.limits) > 0 {
		find := false
		for _, c := range s.limits {
			if c == v {
				find = true
			}
		}
		if !find {
			return fmt.Errorf("value %d not in %+v", v, s.limits)
		}
	}
	*s.val = v
	s.root.setFlag = true
	for _, uf := range s.update {
		uf()
	}
	return nil
}
func (s *customFlagIntValue) Type() string {
	return "int"
}

func (s *customFlagIntValue) String() string { return strconv.FormatInt(*s.val, 10) }

func WriteFile(fname string, data []byte) (err error) {
	os.MkdirAll(filepath.Dir(fname), 0755)
	err = os.WriteFile(fname, data, 0644)
	if err != nil {
		return err
	}
	//log.Println("genrate ", fname)
	return nil
}

// -- stringSlice Value
type customFlagStringSliceValue struct {
	value   *[]string
	root    *ExportSupportConfig
	changed bool
	update  []func()
}

func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

func writeAsCSV(vals []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func (s *customFlagStringSliceValue) Set(val string) error {
	v, err := readAsCSV(val)
	if err != nil {
		return err
	}
	if !s.changed {
		*s.value = v
	} else {
		*s.value = append(*s.value, v...)
	}
	s.changed = true
	s.root.setFlag = true
	for _, uf := range s.update {
		uf()
	}
	return nil
}

func (s *customFlagStringSliceValue) Type() string {
	return "stringSlice"
}

func (s *customFlagStringSliceValue) String() string {
	str, _ := writeAsCSV(*s.value)
	return "[" + str + "]"
}

func (s *customFlagStringSliceValue) Append(val string) error {
	*s.value = append(*s.value, val)
	return nil
}

func (s *customFlagStringSliceValue) Replace(val []string) error {
	*s.value = val
	return nil
}

func (s *customFlagStringSliceValue) GetSlice() []string {
	return *s.value
}

func stringSliceConv(sval string) (interface{}, error) {
	sval = sval[1 : len(sval)-1]
	// An empty string would cause a slice with one (empty) string
	if len(sval) == 0 {
		return []string{}, nil
	}
	return readAsCSV(sval)
}
