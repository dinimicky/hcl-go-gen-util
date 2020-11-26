package main

import (
	"fmt"
	"github.com/dinimicky/hcl-go-gen-util/model"
	"reflect"
	"testing"
	"text/template"
)

func Test_buildRootHclGoSrc(t *testing.T) {
	type args struct {
		provider  string
		resources []model.Hcl
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRootHclGoSrc(tt.args.provider, tt.args.resources); got != tt.want {
				t.Errorf("buildRootHclGoSrc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_AWS(t *testing.T) {
	resNames := model.GetAllProviderResourceName("aws")
	rootResources := make([]model.Hcl, len(resNames), len(resNames))
	resMap := make(map[string]string, len(resNames))
	for i, resName := range resNames {
		fmt.Println("resName", resName)
		resList := model.BuildProviderHclResource("aws", resName)
		resMap[resName] = buildSubHclGoSrc(&resName, resList)
		rootResources[i] = resList[0]
	}
}

func Test_buildSubHclGoSrc(t *testing.T) {
	type args struct {
		resName *string
		resList []model.Hcl
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildSubHclGoSrc(tt.args.resName, tt.args.resList); got != tt.want {
				t.Errorf("buildSubHclGoSrc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_render(t *testing.T) {
	type args struct {
		strTmp  string
		params  interface{}
		funcMap template.FuncMap
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := render(tt.args.strTmp, tt.args.params, tt.args.funcMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("render() = %v, want %v", got, tt.want)
			}
		})
	}
}
