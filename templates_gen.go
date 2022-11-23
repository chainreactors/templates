package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/chainreactors/gogo/v2/pkg"
	"sigs.k8s.io/yaml"
	"strings"
)

var (
	templatePath string
	resultPath   string
)

func loadYamlFile2JsonString(filename string) string {
	var err error
	file, err := os.Open(path.Join(templatePath, filename))
	if err != nil {
		panic(err.Error())
	}

	bs, _ := io.ReadAll(file)
	jsonstr, err := yaml.YAMLToJSON(bs)
	if err != nil {
		panic(filename + err.Error())
	}

	return pkg.Encode(jsonstr)
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

func recuLoadYamlFiles2JsonString(dir string, single bool) string {
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

		if single {
			pocs = append(pocs, tmp)
		} else {
			pocs = append(pocs, tmp.([]interface{})...)
		}

	}

	jsonstr, err := json.Marshal(pocs)
	if err != nil {
		panic(err)
	}

	return pkg.Encode(jsonstr)
}

func parser(key string) string {
	switch key {
	case "tcp":
		return loadYamlFile2JsonString("fingers/tcpfingers.yaml")
	case "http":
		return recuLoadYamlFiles2JsonString("fingers/http", false)
	case "port":
		return loadYamlFile2JsonString("port.yaml")
	case "workflow":
		return loadYamlFile2JsonString("workflows.yaml")
	case "nuclei":
		return recuLoadYamlFiles2JsonString("nuclei", true)
	default:
		panic("illegal key")
	}
}

func main() {
	var needs []string
	flag.StringVar(&templatePath, "t", "templates", "templates repo path")
	flag.StringVar(&resultPath, "o", "templates.go", "result filename")
	need := flag.String("need", "all", "tcp|http|port|workflow|nuclei")
	flag.Parse()

	if *need == "all" {
		needs = []string{"tcp", "http", "port", "workflow", "nuclei"}
	} else {
		needs = strings.Split(*need, ",")
	}

	var s strings.Builder
	var first bool
	for _, n := range needs {
		if !first {
			s.WriteString(fmt.Sprintf("if typ == \"%s\" {\n\t\treturn files.UnFlate(base64Decode(\"%s\"))\n\t}", n, parser(n)))
			first = true
		} else {
			s.WriteString(fmt.Sprintf("else if typ==\"%s\"{\n\t\treturn files.UnFlate(base64Decode(\"%s\"))\n\t}", n, parser(n)))
		}
	}
	template := `package pkg

import (
	"encoding/base64"
	"github.com/chainreactors/files"
)

func base64Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

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
