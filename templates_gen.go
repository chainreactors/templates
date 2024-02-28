package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chainreactors/files"
	en "github.com/chainreactors/utils/encode"
	"io"
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
	s := files.Flate(input)
	return en.Base64Encode(s)
}

func loadYamlFile2JsonString(filename string) string {
	var err error
	file, err := os.Open(path.Join(templatePath, filename))
	if err != nil {
		panic(err.Error())
	}

	bs, _ := io.ReadAll(file)
	jsonstr, err := yaml.YAMLToJSON(bs)
	if err != nil {
		panic(filename + " " + err.Error())
	}

	return encode(jsonstr)
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
	jsonstr, err := json.Marshal(data)
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

func recuLoadPoc2JsonString(dir string) string {
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

	jsonstr, err := json.Marshal(pocs)
	if err != nil {
		panic(err)
	}

	return encode(jsonstr)
}

func recuLoadFinger2JsonString(dir string) string {
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

	jsonstr, err := json.Marshal(pocs)
	if err != nil {
		panic(err)
	}

	return encode(jsonstr)
}

func parser(key string) string {
	switch key {
	case "tcp":
		return loadYamlFile2JsonString("fingers/tcpfingers.yaml")
	case "http":
		return recuLoadFinger2JsonString("fingers/http")
	case "port":
		return loadYamlFile2JsonString("port.yaml")
	case "workflow":
		return loadYamlFile2JsonString("workflows.yaml")
	case "neutron":
		return recuLoadPoc2JsonString("neutron")
	case "spray_rule":
		return loadRawFiles("spray/rule")
	case "spray_common":
		return loadYamlFile2JsonString("spray/common.yaml")
	case "spray_default":
		return loadRawFile("spray/dicc.txt")
	case "extract":
		return loadYamlFile2JsonString("extract.yaml")
	case "zombie_common":
		return loadYamlFile2JsonString("zombie/keywords.yaml")
	case "zombie_default":
		return loadYamlFile2JsonString("zombie/default.yaml")
	case "zombie_rule":
		return loadRawFiles("zombie/rule")
	case "zombie_template":
		return recuLoadPoc2JsonString("neutron/login")
	case "fingerprinthub":
		return loadRawFile("spray/web_fingerprint_v3.json")
	default:
		panic("illegal key")
	}
}

func main() {
	var needs []string
	flag.StringVar(&templatePath, "t", ".", "templates repo path")
	flag.StringVar(&resultPath, "o", "templates.go", "result filename")
	need := flag.String("need", "gogo", "tcp|http|port|workflow|neutron")
	flag.Parse()

	if *need == "gogo" {
		needs = []string{"tcp", "http", "port", "workflow", "neutron", "extract"}
	} else if *need == "spray" {
		needs = []string{"http", "spray_rule", "spray_common", "spray_default", "extract", "fingerprinthub"}
	} else if *need == "zombie" {
		needs = []string{"zombie_default", "zombie_common", "zombie_rule", "zombie_template"}
	} else {
		needs = strings.Split(*need, ",")
	}

	var s strings.Builder
	var first bool
	for _, n := range needs {
		if !first {
			s.WriteString(fmt.Sprintf("if typ == \"%s\" {\n\t\treturn files.UnFlate(encode.Base64Decode(\"%s\"))\n\t}", n, parser(n)))
			first = true
		} else {
			s.WriteString(fmt.Sprintf("else if typ==\"%s\"{\n\t\treturn files.UnFlate(encode.Base64Decode(\"%s\"))\n\t}", n, parser(n)))
		}
	}
	template := `package pkg

import (
	"github.com/chainreactors/files"
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
