
gogo的通用配置, 包括端口, 指纹, poc, workflow

## 目录结构

```
│  templates_gen.go         # template生成器
│  port.yaml                # 端口配置文件
│  workflows.yaml           # workflow配置文件
├─fingers                   # 指纹目录
│  │  tcpfingers.yaml       # tcp指纹
│  │
│  └─http                   # http指纹 因为http指纹较多, 将其分成多个子文件方便管理
│          cloud.yaml       # 云相关的框架
│          cms.yaml         # 各类cms
│          component.yaml   # 内嵌的各种组件
│          device.yaml      # 设备相关
│          mail.yaml        # 邮件相关
│          oa.yaml          # 办公相关
│          other.yaml       # 暂未分类
│          waf.yaml         # waf相关
│
└─nuclei                    # nuclei的poc
    ├─bigip                 # 按框架名分类, 方便管理
    ├─cloud
    ├─component
    ├─device
    ......

```

## 使用

为了方便打包, 大部分情况下, 会将这些配置文件转为json后压缩, 生成为templates.go文件, 进行加载.

因此提供了, `templates_gen.go`

仅有两个参数. `-o` 指定输出的文件名, `-t` templates所在的目录, 默认"."

例如在gogo中, 就在入口文件添加go generate在编译时将templates打包到二进制文件中.

```go
//go:generate go run templates/templates_gen.go -t templates -o pkg/templates.go -need gogo
package main

import "github.com/chainreactors/gogo/v2/cmd"

func main() {
	cmd.Gogo()
}

```

