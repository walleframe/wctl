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
package gen

import (
	"strings"

	"github.com/aggronmagi/wctl/builder/buildpb"
)

func (gen *Generator) Doc(doc *buildpb.DocDesc) {
	if doc == nil || len(doc.Doc) == 0 {
		return
	}

	for _, v := range doc.Doc {
		gen.P(v)
	}
}

func (gen *Generator) LuaDoc(doc *buildpb.DocDesc) {
	if doc == nil || len(doc.Doc) == 0 {
		return
	}

	for _, v := range doc.Doc {
		// if strings.HasPrefix(v, "/*")
		v = "--" + strings.TrimPrefix(v, "//")
		gen.P(v)
	}
}
