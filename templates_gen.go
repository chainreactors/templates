package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/chainreactors/fingers/resources"
	en "github.com/chainreactors/utils/encode"
	"sigs.k8s.io/yaml"
)

var (
	templatePath string
	resultPath   string
	embedMode    bool
)

func deflateCompress(input []byte) []byte {
	return en.MustDeflateCompress(input)
}

func loadYamlFile(filename string) []byte {
	var err error
	file, err := os.Open(path.Join(templatePath, filename))
	if err != nil {
		panic(err.Error())
	}

	bs, _ := ioutil.ReadAll(file)
	return deflateCompress(bs)
}

// getFingersResource 根据 key 返回对应的 fingers/resources 数据
// 这些数据已经是 gzip 压缩的，需要解压后再用 deflate 压缩以保持一致性
func getFingersResource(key string) []byte {
	var data []byte
	switch key {
	case "fingerprinthub_web":
		data = resources.FingerprinthubWebData
	case "nmap_service_probes":
		data = resources.NmapServiceProbesData
	case "nmap_services":
		data = resources.NmapServicesData
	default:
		panic(fmt.Sprintf("unknown fingers resource key: %s", key))
	}

	// 数据是 gzip 压缩的，需要先解压
	decompressed, err := en.GzipDecompress(data)
	if err != nil {
		panic(fmt.Sprintf("failed to decompress %s: %v", key, err))
	}

	return deflateCompress(decompressed)
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if !info.IsDir() {
			*files = append(*files, path)
		}
		return nil
	}
}

func loadRawFiles(dir string) []byte {
	var files []string
	err := filepath.Walk(path.Join(templatePath, dir), visit(&files))
	if err != nil {
		panic(err)
	}
	data := make(map[string]string)
	for _, file := range files {
		bs, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		data[strings.TrimSuffix(filepath.Base(file), ".rule")] = string(bs)
	}
	jsonstr, err := yaml.Marshal(data)
	if err != nil {
		panic(err)
	}

	return deflateCompress(jsonstr)
}

func recuLoadPoc(dir string) []byte {
	var files []string
	err := filepath.Walk(path.Join(templatePath, dir), visit(&files))
	if err != nil {
		panic(err)
	}
	var pocs []interface{}
	for _, file := range files {
		var tmp interface{}
		bs, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		err = yaml.Unmarshal(bs, &tmp)
		if err != nil {
			print(file)
			panic(err)
		}

		if tmp == nil {
			continue
		}

		pocs = append(pocs, tmp)
	}

	jsonstr, err := yaml.Marshal(pocs)
	if err != nil {
		panic(err)
	}

	return deflateCompress(jsonstr)
}

func recuLoadFinger(dir string, isJson bool) []byte {
	var files []string
	err := filepath.Walk(path.Join(templatePath, dir), visit(&files))
	if err != nil {
		panic(err)
	}
	var pocs []interface{}
	for _, file := range files {
		// 使用父目录名作为 tag，而不是文件名
		// 例如: fingers/http/cdn/aliyun.yaml -> tag 为 "cdn"
		parentDir := filepath.Base(filepath.Dir(file))
		var tmp interface{}
		bs, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		err = yaml.Unmarshal(bs, &tmp)
		if err != nil {
			print(file)
			panic(err)
		}

		if tmp == nil {
			continue
		}
		fingers := tmp.([]interface{})
		for i, finger := range fingers {
			f := finger.(map[string]interface{})
			f["link"] = ""
			f["tag"] = []string{parentDir}
			fingers[i] = f
		}
		pocs = append(pocs, fingers...)
	}

	var content []byte
	if isJson {
		content, err = json.Marshal(pocs)
		if err != nil {
			panic(err)
		}
	} else {
		content, err = yaml.Marshal(pocs)
		if err != nil {
			panic(err)
		}
	}

	return deflateCompress(content)
}

func parser(key string) []byte {
	switch key {
	case "socket":
		return recuLoadFinger("fingers/socket", true)
	case "http":
		return recuLoadFinger("fingers/http", true)
	case "fingerprinthub_web":
		return getFingersResource("fingerprinthub_web")
	case "nmap_service_probes":
		return getFingersResource("nmap_service_probes")
	case "nmap_services":
		return getFingersResource("nmap_services")
	case "port":
		return loadYamlFile("port.yaml")
	case "workflow":
		return loadYamlFile("workflows.yaml")
	case "neutron":
		return recuLoadPoc("neutron")
	case "spray_rule":
		return loadRawFiles("spray/rule")
	case "spray_common":
		return loadYamlFile("spray/common.yaml")
	case "spray_dict":
		return loadRawFiles("spray/dict")
	case "extract":
		return loadYamlFile("extract.yaml")
	case "proton_rules":
		return recuLoadPoc("proton_rules")
	case "zombie_common":
		return loadYamlFile("zombie/keywords.yaml")
	case "zombie_default":
		return loadYamlFile("zombie/default.yaml")
	case "zombie_rule":
		return loadRawFiles("zombie/rule")
	case "zombie_template":
		return recuLoadPoc("neutron/login")
	case "found_keys":
		return recuLoadPoc("found/keys")
	case "found_auto":
		return recuLoadPoc("found/auto")
	case "found_filter_ext":
		return loadYamlFile("found/filters/extensions.yaml")
	case "found_filter_dir":
		return loadYamlFile("found/filters/directories.yaml")
	default:
		panic("illegal key")
	}
}

func toVarName(key string) string {
	parts := strings.Split(key, "_")
	for i, p := range parts {
		if len(p) > 0 {
			if i == 0 {
				parts[i] = strings.ToLower(p[:1]) + p[1:]
			} else {
				parts[i] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
	}
	return strings.Join(parts, "") + "Data"
}

func main() {
	var needs []string
	flag.StringVar(&templatePath, "t", ".", "templates repo path")
	flag.StringVar(&resultPath, "o", "templates.go", "result filename")
	need := flag.String("need", "gogo", "socket|http|port|workflow|neutron")
	flag.BoolVar(&embedMode, "embed", false, "use go:embed for binary data (requires Go 1.16+)")
	flag.Parse()

	if *need == "gogo" {
		needs = []string{
			"socket", "http",
			"fingerprinthub_web",
			"nmap_service_probes", "nmap_services",
			"port", "workflow", "neutron", "extract",
		}
	} else if *need == "spray" {
		needs = []string{"spray_rule", "spray_common", "spray_dict", "extract", "proton_rules", "port"}
	} else if *need == "zombie" {
		needs = []string{"zombie_default", "zombie_common", "zombie_rule", "zombie_template", "port", "socket", "http"}
	} else if *need == "found" {
		needs = []string{"found_keys", "found_auto", "found_filter_ext", "found_filter_dir"}
	} else {
		needs = strings.Split(*need, ",")
	}

	if embedMode {
		generateEmbed(needs)
	} else {
		generateLegacy(needs)
	}
}

func generateLegacy(needs []string) {
	var s strings.Builder
	var first bool
	for _, n := range needs {
		b64 := en.Base64Encode(parser(n))
		if !first {
			s.WriteString(fmt.Sprintf("if typ == \"%s\" {\n\t\treturn encode.MustDeflateDeCompress(encode.Base64Decode(\"%s\"))\n\t}", n, b64))
			first = true
		} else {
			s.WriteString(fmt.Sprintf("else if typ==\"%s\"{\n\t\treturn encode.MustDeflateDeCompress(encode.Base64Decode(\"%s\"))\n\t}", n, b64))
		}
	}
	tmpl := `//go:build !emptytemplates
// +build !emptytemplates

package pkg

import (
	"github.com/chainreactors/utils/encode"
)

func loadEmbeddedConfig(typ string) []byte {
	%s
	return nil
}
`
	tmpl = fmt.Sprintf(tmpl, s.String())
	writeOutput(resultPath, tmpl)
}

func generateEmbed(needs []string) {
	dataDir := filepath.Join(filepath.Dir(resultPath), "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("create data dir: %v", err))
	}

	var embedDecls strings.Builder
	var loadBody strings.Builder
	first := false

	for _, n := range needs {
		data := parser(n)

		binFile := n + ".bin"
		binPath := filepath.Join(dataDir, binFile)
		if err := os.WriteFile(binPath, data, 0644); err != nil {
			panic(fmt.Sprintf("write %s: %v", binPath, err))
		}
		fmt.Printf("  embed: %s (%d bytes)\n", binFile, len(data))

		varName := toVarName(n)
		embedDecls.WriteString(fmt.Sprintf("//go:embed data/%s\nvar %s []byte\n\n", binFile, varName))

		if !first {
			loadBody.WriteString(fmt.Sprintf("if typ == \"%s\" {\n\t\treturn encode.MustDeflateDeCompress(%s)\n\t}", n, varName))
			first = true
		} else {
			loadBody.WriteString(fmt.Sprintf("else if typ == \"%s\" {\n\t\treturn encode.MustDeflateDeCompress(%s)\n\t}", n, varName))
		}
	}

	tmpl := fmt.Sprintf(`//go:build !emptytemplates
// +build !emptytemplates

package pkg

import (
	_ "embed"

	"github.com/chainreactors/utils/encode"
)

%s
func loadEmbeddedConfig(typ string) []byte {
	%s
	return nil
}
`, embedDecls.String(), loadBody.String())

	writeOutput(resultPath, tmpl)
}

func writeOutput(path, content string) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	f.WriteString(content)
	f.Sync()
	f.Close()
	println("generate templates.go successfully")
}
