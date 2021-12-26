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
package plugin

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aggronmagi/wctl/builder/buildpb"
	"github.com/aggronmagi/wctl/utils"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
)

func init() {
	pflag.BoolVar(&utils.Flag.Debug, "debug", utils.Flag.Debug, "是否打印调试信息")
	pflag.BoolVar(&utils.Flag.ShowDetail, "debug-detail", utils.Flag.ShowDetail, "是否打印详细调试信息")
}

func Debug() bool {
	return utils.Debug()
}

func ShowDetail() bool {
	return utils.Flag.ShowDetail
}

// MainRoot 插件主函数
func MainRoot(gen func(rq *buildpb.BuildRQ) (rs *buildpb.BuildRS, err error)) {
	pflag.Parse()

	if ShowDetail() {
		log.Println("start plugin")
	}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
		return
	}
	if ShowDetail() {
		log.Println("start plugin 1")
	}
	req := &buildpb.BuildRQ{}
	err = proto.Unmarshal(data, req)
	if err != nil {
		log.Fatal(err)
		return
	}
	if ShowDetail() {
		log.Println("recv", req)
	}
	// 生成配置
	res, err := gen(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	if res == nil {
		res = &buildpb.BuildRS{}
	}
	data, err = proto.Marshal(res)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Fatal(err)
		return
	}
	if ShowDetail() {
		log.Println("plugin success")
	}
}

// MainOneByOne 单个接口
func MainOneByOne(gf func(prog *buildpb.FileDesc, depend map[string]*buildpb.FileDesc) (out []*buildpb.BuildOutput, err error)) {
	MainRoot(func(rq *buildpb.BuildRQ) (rs *buildpb.BuildRS, err error) {
		rs = &buildpb.BuildRS{}
		for _, file := range rq.Files {
			fdesc, ok := rq.Programs[file]
			if !ok {
				err = fmt.Errorf("%s not exists", file)
				return
			}
			one, err := gf(fdesc, rq.Programs)
			if err != nil {
				return rs, err
			}
			if one == nil {
				continue
			}
			rs.Result = append(rs.Result, one...)
		}
		return
	})
	return
}
