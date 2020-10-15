package main

import (
	"flag"
	"fmt"
	"go/build"
	"hcl-go-gen-util/model"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	pkgInfo  *build.Package
	provider = flag.String(
		"provider", "tencentcloud",
		fmt.Sprintf("generate golang struct for specific provider, SupportedProvider: %v ", model.SupportedProvider),
	)
	isShowComputed = flag.Bool("show-computed", false, "show computed attribute in golang struct if true")
)

func main() {
	flag.Parse()

	root, subResoources := model.ProviderHclResource(*provider)
	structs := make([]string, len(subResoources)+1)
	for i, subres := range subResoources {
		structs[i] = subres.GoString(*isShowComputed)
	}
	structs[len(subResoources)] = root.GoString(*isShowComputed)

	src := strings.Join(structs, "\n")

	pkgName := os.Getenv("GOPACKAGE")
	if pkgName == "" {
		pkgName = pkgInfo.Name
	}
	src = fmt.Sprintf("package %v \n %v", pkgName, src)

	//save to file
	outputName := strings.ToLower(fmt.Sprintf("%s.go", provider))

	err := ioutil.WriteFile(outputName, []byte(src), 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}
