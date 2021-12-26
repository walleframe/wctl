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
package utils

// 通用配置
type walleCommonFlag struct {
	// 调试标记
	Debug bool
	// 显示详情
	ShowDetail bool
}

// Flag 导出全局通用标记
var Flag = &walleCommonFlag{
	Debug:      false,
	ShowDetail: false,
}

// Debug 是否打印调试信息
func Debug() bool {
	return Flag.Debug || Flag.ShowDetail
}

func ShowDetail() bool {
	return Flag.ShowDetail
}
