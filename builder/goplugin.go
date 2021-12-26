/*
   Copyright © 2020 aggronmagi <czy463@163.com>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package builder

import (
	"fmt"
	"plugin"
)

// LoadGoPluginGenerater 加载插件,并生效
func LoadGoPluginGenerater(name string) (err error) {
	// strings.Split(name, "-")
	so, err := plugin.Open(name)
	if err != nil {
		return err
	}
	symbol, err := so.Lookup("NewGenerator")
	if err != nil {
		return err
	}
	ng, ok := symbol.(func() Generater)
	if !ok {
		return fmt.Errorf("NewGenerator Is Not func()Generator")
	}
	iface := ng()
	if _, ok := factory[iface.Union()]; ok {
		fmt.Println("WARN 替换插件:", iface.Union(), name)
	}
	// 保存生成器
	factory[iface.Union()] = iface
	// 生效插件
	addUse(iface)
	return
}
