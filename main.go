package tencentcloud_instance

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
	err      error
	pkgInfo  *build.Package
	provider = flag.String(
		"provider", "tencentcloud",
		fmt.Sprintf("generate golang struct for specific provider, SupportedProvider: %v ", model.SupportedProvider),
	)
	isShowComputed = flag.Bool("show-computed", false, "show computed attribute in golang struct if true")
	resName        = flag.String(
		"res", "tencentcloud_instance", fmt.Sprintf(
			"supported res name %v",
			model.GetAllProviderResourceName(*provider),
		),
	)
)

func main() {
	flag.Parse()

	resList := model.BuildProviderHclResource(*provider, *resName)
	root := model.BuildProviderRootResource(*provider, []model.Hcl{resList[0]})

	resList = append(resList, root)
	structs := make([]string, len(resList)+1)
	for i, hcl := range resList {

		structs[i] = hcl.GoString(*isShowComputed)

	}

	src := strings.Join(structs, "\n")

	pkgName := os.Getenv("GOPACKAGE")
	pkgInfo, err = build.ImportDir(".", 0)
	if err != nil {
		log.Fatal(err)
	}
	if pkgName == "" {
		pkgName = pkgInfo.Name
	}
	src = fmt.Sprintf("package %v \n %v", pkgName, src)

	//save to file
	outputName := strings.ToLower(fmt.Sprintf("%v_%v.go", *provider, *resName))

	err = ioutil.WriteFile(outputName, []byte(src), 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}
