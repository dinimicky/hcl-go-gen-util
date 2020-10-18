package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/dinimicky/hcl-go-gen-util/model"
	"github.com/dinimicky/hcl-go-gen-util/util"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	err      error
	pkgInfo  *build.Package
	provider = flag.String(
		"provider", "tencentcloud",
		fmt.Sprintf("generate golang struct for specific provider, SupportedProvider: %v ", model.SupportedProvider),
	)
	isShowComputed = flag.Bool("show-computed", false, "show computed attribute in golang struct if true")
	resName        = flag.String(
		"res", "tencentcloud_instance",
		fmt.Sprintf("supported res name %v", model.GetAllProviderResourceName(*provider)),
	)
	importpath = flag.String("importpath", "", "set import path")
)

const (
	importSrcTemplate = `{{$path:=.ImportPath}}
import (
	{{range  .Resources}}  "{{$path}}/{{.}}"
{{end}})`
)

func render(strTmp string, params interface{}, funcMap template.FuncMap) []byte {
	//利用模板库，生成代码文件
	t, err := template.New("").Funcs(funcMap).Parse(strTmp)
	if err != nil {
		log.Fatal(err)
	}
	buff := bytes.NewBufferString("")
	err = t.Execute(buff, params)
	if err != nil {
		log.Fatal(err)
	}

	return buff.Bytes()
}

func buildSubHclGoSrc(resName *string, resList []model.Hcl) string {
	structs := make([]string, len(resList))
	for i, hcl := range resList {

		structs[i] = hcl.GoString(*isShowComputed)

	}

	src := strings.Join(structs, "\n")
	src = fmt.Sprintf("package %v \n %v", *resName, src)
	return src
}

func buildRootHclGoSrc(provider string, resources []model.Hcl) string {

	structMember := make([]string, len(resources))
	for i, r := range resources {
		structMember[i] = fmt.Sprintf("%v []%v.%v %v", r.GoType(), util.Camel2Case(r.GoType()), r.GoType(), r.HclTag())
	}
	structBody := strings.Join(structMember, "\n")
	return fmt.Sprintf("type %v struct { %v \n }", util.Case2Camel(provider), structBody)

}
func main() {
	flag.Parse()
	resNames := model.GetAllProviderResourceName(*provider)
	rootResources := make([]model.Hcl, len(resNames), len(resNames))
	resMap := make(map[string]string, len(resNames))
	for i, resName := range resNames {
		resList := model.BuildProviderHclResource(*provider, resName)
		resMap[resName] = buildSubHclGoSrc(&resName, resList)
		rootResources[i] = resList[0]
	}

	pkgName := os.Getenv("GOPACKAGE")
	pkgInfo, err = build.ImportDir(".", 0)
	//pkgStr, err := json.Marshal(pkgInfo)

	if err != nil {
		log.Fatal(err)
	}
	if pkgName == "" {
		pkgName = pkgInfo.Name
	}

	//save to file
	for resName, src := range resMap {

		outputName := strings.ToLower(fmt.Sprintf("%v.go", resName))
		err = os.MkdirAll(resName, 0744)
		err = ioutil.WriteFile(filepath.Join(resName, outputName), []byte(src), 0644)
	}

	if err != nil {
		log.Fatalf("writing output: %s", err)
	}

	//root
	//rootSrc := buildRootHclGoSrc(*provider, rootResources)
	//type ImportStruct struct {
	//	ImportPath string
	//	Resources  []string
	//}
	//params := &ImportStruct{
	//	ImportPath: *importpath,
	//	Resources:  resNames,
	//}
	//importSrc := string(render(importSrcTemplate, params, nil))
	//rootSrc = fmt.Sprintf("package %v \n %v \n %v ", pkgName, importSrc, rootSrc)
	//outputName2 := strings.ToLower(fmt.Sprintf("%v.go", *provider))

	//err = ioutil.WriteFile(outputName2, []byte(rootSrc), 0644)
	//if err != nil {
	//	log.Fatalf("writing output: %s", err)
	//}
}
