package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chainreactors/files"
	"github.com/chainreactors/parsers"
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
	return parsers.Base64Encode(s)
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
		panic(filename + err.Error())
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
		data[strings.TrimSuffix(filepath.Base(file), ".txt")] = string(bs)
	}
	jsonstr, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return encode(jsonstr)
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
	case "nuclei":
		return recuLoadPoc2JsonString("nuclei")
	case "rule":
		return loadRawFiles("rule")
	case "mask":
		return loadYamlFile2JsonString("keywords.yaml")
	default:
		panic("illegal key")
	}
}

func main() {
	var needs []string
	flag.StringVar(&templatePath, "t", ".", "templates repo path")
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
			s.WriteString(fmt.Sprintf("if typ == \"%s\" {\n\t\treturn files.UnFlate(parsers.Base64Decode(\"%s\"))\n\t}", n, parser(n)))
			first = true
		} else {
			s.WriteString(fmt.Sprintf("else if typ==\"%s\"{\n\t\treturn files.UnFlate(parsers.Base64Decode(\"%s\"))\n\t}", n, parser(n)))
		}
	}
	template := `package pkg

import (
	"github.com/chainreactors/files"
	"github.com/chainreactors/parsers"
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
