# templates

chainreactors 工具链的通用模板仓库，为 [gogo](https://github.com/chainreactors/gogo)、[spray](https://github.com/chainreactors/spray)、[zombie](https://github.com/chainreactors/zombie)、[found](https://github.com/chainreactors/found) 提供统一的配置与规则。

## 目录结构

```
├── templates_gen.go            # 模板生成器，将 YAML/规则打包为嵌入式二进制数据
├── templates_gen_test.go
├── port.yaml                   # 端口与服务映射配置
├── workflows.yaml              # gogo workflow 编排
├── extract.yaml                # 通用正则提取器（js/url 爬取）
│
├── fingers/                    # 指纹识别规则
│   ├── http/                   # HTTP 指纹（13 个分类）
│   │   ├── cdn.yaml            # CDN
│   │   ├── cloud.yaml          # 云服务
│   │   ├── cms.yaml            # CMS
│   │   ├── component.yaml      # 组件
│   │   ├── device.yaml         # 设备
│   │   ├── language.yaml       # 语言/框架
│   │   ├── mail.yaml           # 邮件
│   │   ├── oa.yaml             # OA 办公
│   │   ├── spray.yaml          # spray 专用
│   │   ├── supply.yaml         # 供应链
│   │   ├── waf.yaml            # WAF
│   │   └── ...
│   └── socket/                 # TCP 指纹
│       └── tcpfingers.yaml
│
├── neutron/                    # neutron 漏洞检测 POC（83 个）
│   ├── login/                  # 默认口令/弱口令检测（40+）
│   ├── spring/                 # Spring 系列漏洞
│   ├── weblogic/               # WebLogic 漏洞
│   ├── vmware/                 # VMware/vCenter 漏洞
│   └── ...                     # bigip, gitlab, grafana, tomcat 等
│
├── spray/                      # spray 目录扫描工具配置
│   ├── common.yaml             # 通用配置
│   ├── dict/                   # 字典文件（14 个，admin/cgi/java/swagger 等）
│   ├── rule/                   # 变换规则（authbypass/extbypass/filebak）
│   └── proton/                 # 响应内容提取规则（45 个）
│       ├── cloud/              # 云 AK/SK 泄露（aliyun/aws/tencent 等）
│       ├── crawl/              # URL/JS 提取
│       ├── credential/         # 凭据泄露（password/jwt/rsa-key 等）
│       ├── info-leak/          # 信息泄露（email/phone/ip/idcard 等）
│       └── token/              # Token/API Key 泄露
│
├── found/                      # found 敏感文件/凭据发现工具配置
│   ├── auto/                   # 按 OS 分类的敏感路径（linux/windows/darwin）
│   ├── filters/                # 扫描过滤规则（扩展名/目录排除）
│   ├── keys/                   # API Key/凭据检测规则（159 个）
│   └── spray/                  # 响应内容提取规则（36 个）
│
├── services/                   # zombie 服务交互模板（75 个）
│   ├── ftp/                    # FTP 敏感文件读取
│   ├── ldap/                   # LDAP 用户枚举
│   ├── mssql/                  # MSSQL 提权/信息收集/持久化（19 个）
│   ├── mysql/                  # MySQL UDF/文件读取/哈希提取
│   ├── oracle/                 # Oracle 提权/调度执行
│   ├── postgresql/             # PostgreSQL RCE/大对象读写
│   ├── redis/                  # Redis 写 webshell/SSH key/crontab
│   ├── smb/                    # SMB 共享枚举
│   ├── ssh/                    # SSH 凭据/信息收集/Docker 逃逸
│   └── ...                     # memcached, mongodb
│
└── zombie/                     # zombie 爆破工具配置
    ├── default.yaml            # 默认配置
    ├── keywords.yaml           # 关键词字典
    ├── rule/                   # 密码变换规则（rockyou/weakpass）
    └── loot/                   # 后渗透数据提取规则（10 个）
        ├── loot-bank-card.yaml
        ├── loot-cloud-credential.yaml
        ├── loot-connection-string.yaml
        └── ...                 # email, id-card, jwt, password-hash, phone, private-key, internal-ip
```

## 工具对应关系

| 工具 | 使用的模板 | 构建参数 |
|------|-----------|---------|
| gogo | `fingers/`, `neutron/`, `port.yaml`, `workflows.yaml`, `extract.yaml` | `-need gogo` |
| spray | `spray/`, `extract.yaml`, `found/keys/`, `port.yaml` | `-need spray` |
| zombie | `zombie/`, `neutron/login/`, `services/`, `port.yaml`, `fingers/` | `-need zombie` |
| found | `found/` | `-need found` |

## 使用

模板通过 `templates_gen.go` 打包为压缩后的二进制数据，嵌入到对应工具中。

支持两种模式：
- **legacy 模式**（默认）：生成 base64 编码的 Go 源文件
- **embed 模式**（`-embed`）：生成 `go:embed` 指令 + `.bin` 文件

参数：
- `-t` 模板目录路径（默认 `.`）
- `-o` 输出文件名（默认 `templates.go`）
- `-need` 指定构建目标（`gogo`/`spray`/`zombie`/`found`，或逗号分隔的 key 列表）
- `-embed` 使用 `go:embed` 模式

示例（gogo）：

```go
//go:generate go run templates/templates_gen.go -t templates -o pkg/templates.go -need gogo
package main

import "github.com/chainreactors/gogo/v2/cmd"

func main() {
	cmd.Gogo()
}
```

示例（spray，embed 模式）：

```go
//go:generate go run templates/templates_gen.go -t templates -o pkg/templates.go -need spray -embed
```
