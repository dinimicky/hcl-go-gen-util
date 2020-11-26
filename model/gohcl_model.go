package model

import (
	"bytes"
	"fmt"
	"github.com/dinimicky/hcl-go-gen-util/util"
	"github.com/hashicorp/go-hclog"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud"
	"github.com/terraform-providers/terraform-provider-aws/aws"
	"go/format"
	"log"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	_ "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud"
	_ "github.com/terraform-providers/terraform-provider-aws/aws"
)

func getMapKeys(m map[string]*schema.Provider) []string {
	keys := make([]string, len(m))
	j := 0
	for k := range m {
		keys[j] = k
		j++
	}
	return keys
}

var (
	logger = hclog.L()

	ResourceIdSchema, cloudProviderMap = NewHclSchema("id", &schema.Schema{Type: schema.TypeString, Computed: true}), map[string]*schema.Provider{
		"tencentcloud": tencentcloud.Provider().(*schema.Provider),
		"huaweicloud":  huaweicloud.Provider().(*schema.Provider),
		"aws":          aws.Provider().(*schema.Provider), //aws version v2.70.0
	}
	SupportedProvider = getMapKeys(cloudProviderMap)
)

type Hcl interface {
	GoString(isDisplayComputed bool) string
	GoType() string
	HclTag() string
}

type hclResource struct {
	ResourceName string
	HclBlkName   string
	LabelNames   []string
	HclLabelTag  string
	HclSchemas   []Hcl
}

type hclSchema struct {
	TypeName  string
	ValueType schema.ValueType
	Optional  bool
	Required  bool
	Computed  bool
	Elem      Hcl
}

func NewHclSchema(typeName string, sa *schema.Schema) Hcl {
	hs := &hclSchema{
		TypeName:  typeName,
		ValueType: sa.Type,
		Optional:  sa.Optional,
		Required:  sa.Required,
		Computed:  sa.Computed,
	}
	switch sa.Elem.(type) {
	case *schema.Resource:
		hs.Elem = NewHclResource(typeName, typeName, sa.Elem.(*schema.Resource), nil)
	case *schema.Schema:
		hs.Elem = NewHclSchema(typeName, sa.Elem.(*schema.Schema))
	case schema.ValueType:
		hs.Elem = NewHclSchema(typeName, &schema.Schema{
			Type:     sa.Elem.(schema.ValueType),
			Required: true,
		})
	case nil:
	default:
		panic(fmt.Errorf("Unsupported Elem type %T, typeName %v", sa.Elem, typeName))
	}

	return hs
}

func NewHclResource(resName, hclBlkName string, res *schema.Resource, extraHcl Hcl, label ...string) Hcl {
	saList := make([]Hcl, len(res.Schema))

	i := 0
	for k, v := range res.Schema {
		hs := NewHclSchema(k, v)
		if ptrHs, ok := hs.(*hclSchema); ok {
			saList[i] = ptrHs
		}
		i++
	}

	if extraHcl != nil {
		saList = append(saList, extraHcl)
	}

	return &hclResource{
		ResourceName: resName,
		HclBlkName:   hclBlkName,
		LabelNames:   label,
		HclLabelTag:  fmt.Sprintf("`hcl:\"%v,label\"`", resName),
		HclSchemas:   saList,
	}
}

func (hs *hclSchema) GoString(isDisplayComputed bool) string {
	if hs.Optional || hs.Required || (isDisplayComputed && hs.Computed) {
		return fmt.Sprintf("%v %v %v", util.Case2Camel(hs.TypeName), hs.GoType(), hs.HclTag())
	}

	return ""
}

func (hs *hclSchema) GoType() string {
	switch hs.ValueType {
	case schema.TypeBool:
		if hs.Optional || hs.Computed {
			return "*bool"
		}
		return "bool"
	case schema.TypeInt:
		if hs.Optional || hs.Computed {
			return "*int"
		}
		return "int"
	case schema.TypeFloat:
		if hs.Optional || hs.Computed {
			return "*float32"
		}
		return "float32"
	case schema.TypeString:
		if hs.Optional || hs.Computed {
			return "*string"
		}
		return "string"
	case schema.TypeList, schema.TypeSet:
		return fmt.Sprintf("[]%v", hs.Elem.GoType())
	case schema.TypeMap:
		if hs.Elem == nil {
			return fmt.Sprintf("map[string]string")
		}
		return fmt.Sprintf("map[string]%v", hs.Elem.GoType())
	default:
		return ""
	}
}

func (hr *hclSchema) HclTag() string {
	if ehr, ok := hr.Elem.(*hclResource); hr.Elem != nil && ok {
		return ehr.HclTag()
	}
	if hr.Optional {
		return fmt.Sprintf("`hcl:\"%v,optional\"`", hr.TypeName)
	}
	return fmt.Sprintf("`hcl:\"%v\"`", hr.TypeName)

}
func (hr *hclResource) GoString(isDisplayComputed bool) string {
	const strTmp = `type {{ Case2Camel .ResourceName}} struct {
{{$tag := .HclLabelTag}}
{{range  .LabelNames}}{{.}} string {{$tag}} 
{{end}}
{{range .HclSchemas}} {{   .GoString IsDisplayComputed }}
{{end}}
}`
	t, err := template.New("").Funcs(
		template.FuncMap{
			"Case2Camel":        util.Case2Camel,
			"IsDisplayComputed": func() bool { return isDisplayComputed },
		},
	).Parse(strTmp)
	if err != nil {
		log.Fatal(err)
	}
	return render(t, hr)
}

func (hr *hclResource) GoType() string {
	return util.Case2Camel(hr.ResourceName)
}

func render(t *template.Template, params interface{}) string {
	//利用模板库，生成代码文件

	buff := bytes.NewBufferString("")
	err := t.Execute(buff, params)
	if err != nil {
		log.Fatal(err)
	}
	//格式化
	src, err := format.Source(buff.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	return string(src)
}

func (hr *hclResource) HclTag() string {
	return fmt.Sprintf("`hcl:\"%v,block\"`", hr.HclBlkName)
}

func collectHclResources(hcl Hcl, m *map[string]Hcl) {
	if hr, ok := hcl.(*hclResource); ok {
		(*m)[hr.ResourceName] = hr
		for _, hs := range hr.HclSchemas {
			collectHclResources(hs, m)
		}
	}
	if hs, ok := hcl.(*hclSchema); ok {
		collectHclResources(hs.Elem, m)
	}
}

func collect(resName string, resource *schema.Resource) []Hcl {

	hclres := NewHclResource(resName, "resource", resource, ResourceIdSchema, "HclResLabelType", "HclResLabelName")
	m := make(map[string]Hcl)
	collectHclResources(hclres, &m)
	hcls := make([]Hcl, 0)
	for _, h := range m {
		hcls = append(hcls, h)
	}
	return hcls

}

func BuildProviderRootResource(provider string, subResoources []Hcl) Hcl {
	hss := make([]Hcl, len(subResoources))

	for i, res := range subResoources {
		hs := &hclSchema{
			TypeName:  res.GoType(),
			ValueType: schema.TypeList,
			Optional:  true,
			Elem:      res,
		}
		hss[i] = hs
	}

	return &hclResource{
		ResourceName: provider,
		HclBlkName:   "resources",
		HclSchemas:   hss,
	}
}

func BuildProviderHclResource(provider string, resName string) []Hcl {
	providerResources, ok := cloudProviderMap[provider]
	if !ok {
		panic(fmt.Errorf("provider %v not supported, supported provider is %v", provider, SupportedProvider))
	}
	if _, ok := providerResources.ResourcesMap[resName]; !ok {
		panic(
			fmt.Errorf(
				"resource name %v is wrong, supported resources are %v", resName,
				GetAllProviderResourceName(provider),
			),
		)
	}
	return collect(resName, providerResources.ResourcesMap[resName])

}

func GetAllProviderResourceName(provider string) []string {
	providerResources, ok := cloudProviderMap[provider]
	if !ok {
		panic(fmt.Errorf("provider %v not supported, supported provider is %v", provider, SupportedProvider))
	}

	res := make([]string, len(providerResources.ResourcesMap))
	i := 0
	for k := range providerResources.ResourcesMap {
		res[i] = k
		i++
	}

	return res
}
