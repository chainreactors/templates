package main

import (
	"encoding/json"
	"flag"
	"fmt"
	en "github.com/chainreactors/utils/encode"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

var (
	templatePath string
	resultPath   string
)

func encode(input []byte) string {
	s := en.MustDeflateCompress(input)
	return en.Base64Encode(s)
}

func loadYamlFile(filename string) string {
	var err error
	file, err := os.Open(path.Join(templatePath, filename))
	if err != nil {
		panic(err.Error())
	}

	bs, _ := ioutil.ReadAll(file)
	return encode(bs)
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

func loadRawFiles(dir string) string {
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

	return encode(jsonstr)
}

func loadRawFile(dir string) string {
	content, err := os.ReadFile(path.Join(templatePath, dir))
	if err != nil {
		panic(err)
	}
	return encode(content)
}

func recuLoadPoc(dir string) string {
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

	return encode(jsonstr)
}

func recuLoadFinger(dir string, isJson bool) string {
	var files []string
	err := filepath.Walk(path.Join(templatePath, dir), visit(&files))
	if err != nil {
		panic(err)
	}
	var pocs []interface{}
	for _, file := range files {
		filename := filepath.Base(file)
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
			f["tag"] = []string{strings.TrimSuffix(filename, ".yaml")}
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

	return encode(content)
}

func parser(key string) string {
	switch key {
	case "socket":
		return recuLoadFinger("fingers/socket", true)
	case "http":
		return recuLoadFinger("fingers/http", true)
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
	case "zombie_common":
		return loadYamlFile("zombie/keywords.yaml")
	case "zombie_default":
		return loadYamlFile("zombie/default.yaml")
	case "zombie_rule":
		return loadRawFiles("zombie/rule")
	case "zombie_template":
		return recuLoadPoc("neutron/login")
	default:
		panic("illegal key")
	}
}

func main() {
	var needs []string
	flag.StringVar(&templatePath, "t", ".", "templates repo path")
	flag.StringVar(&resultPath, "o", "templates.go", "result filename")
	need := flag.String("need", "gogo", "socket|http|port|workflow|neutron")
	flag.Parse()

	if *need == "gogo" {
		needs = []string{"socket", "http", "port", "workflow", "neutron", "extract"}
	} else if *need == "spray" {
		needs = []string{"spray_rule", "spray_common", "spray_dict", "extract", "port"}
	} else if *need == "zombie" {
		needs = []string{"zombie_default", "zombie_common", "zombie_rule", "zombie_template", "port", "socket", "http"}
	} else {
		needs = strings.Split(*need, ",")
	}

	var s strings.Builder
	var first bool
	for _, n := range needs {
		if !first {
			s.WriteString(fmt.Sprintf("if typ == \"%s\" {\n\t\treturn encode.MustDeflateDeCompress(encode.Base64Decode(\"%s\"))\n\t}", n, parser(n)))
			first = true
		} else {
			s.WriteString(fmt.Sprintf("else if typ==\"%s\"{\n\t\treturn encode.MustDeflateDeCompress(encode.Base64Decode(\"%s\"))\n\t}", n, parser(n)))
		}
	}
	template := `package pkg

import (
	"github.com/chainreactors/utils/encode"
)


var RandomDir = "/g8kZMwp4oeKsL2in"

func LoadConfig(typ string)[]byte  {
	%s
	return []byte{}
}
`
	template = fmt.Sprintf(template, s.String())
	f, err := os.OpenFile(resultPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	f.WriteString(template)
	f.Sync()
	f.Close()
	println("generate templates.go successfully")
}
