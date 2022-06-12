package builtin

var Template = `// Code generated by dbtest; DO NOT EDIT.
// github.com/RussellLuo/dbtest

package {{$.DstPkgName}}_test

import (
	"fmt"
	"testing"
	"time"
	"github.com/protogodev/dbtest/builtin"
	"github.com/protogodev/dbtest/spec"

	{{- range $.Imports}}
	{{.ImportString}}
	{{- end}}
)

var (
	testee *builtin.Testee
	sut    {{$.SrcPkgName}}.{{$.InterfaceName}}
)

func TestMain(m *testing.M) {
	t, err := {{.Testee}}
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	testee = t.Complete()
	if err := testee.Validate(); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	sut = testee.SUT.({{$.SrcPkgName}}.{{$.InterfaceName}})

	// os.Exit() does not respect deferred functions
	code := m.Run()

	_ = testee.DB.Close()
	os.Exit(code)
}

{{- range $.Tests}}
{{- $method := interfaceMethod .Name}}

func Test{{.Name}}(t *testing.T) {
	f := builtin.NewFixture(t, testee.DB, map[string]spec.Rows{
		{{- range $tableName, $rows := .Fixture }}
		"{{$tableName}}": {
			{{- range $rows}}
			{{.LiteralString}},
			{{- end}} {{/* range $rows */}}
		},
		{{- end}} {{/* range $tableName, $rows := .Fixture */}}
	})
	f.SetUp()
	defer f.TearDown()

	// in contains all the input parameters (except ctx) of {{.Name}}.
	type in struct {
		{{- range $method.Params}}
		{{title .Name}} {{.TypeString}} ` + "`dbtest:\"{{.Name}}\"`" + `
		{{- end}}
	}
	
	// out contains all the output parameters of {{.Name}}.
	type out struct {
		{{- range $method.Returns}}
		{{title .Name}} {{.TypeString}} ` + "`dbtest:\"{{.Name}}\"`" + `
		{{- end}}
	}

	tests := []struct {
		name     string
		in       map[string]interface{}
		wantOut  map[string]interface{}
		wantData []spec.DataAssertion
	}{
	    {{- range .Subtests}}
		{
			name: "{{.Name}}",
			in: {{goString .In}},
			wantOut: {{goString .WantOut}},
			wantData: []spec.DataAssertion{
			    {{- range .WantData}}
				{
					Query: "{{.Query}}",
					Result: spec.Rows{
						{{- range .Result}}
						{{.LiteralString}},
						{{- end}} {{/* range .Result */}}
					},
				},
			    {{- end}} {{/* range .WantData */}}
			},
		},
	    {{- end}} {{/* range .Subtests */}}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in in
			if err := testee.Codec.Decode(tt.in, &in); err != nil {
				t.Errorf("err when decoding In: %v", err)
			}

			var gotOut out
			{{fmtArgCSV $method.ReturnArgValueList "gotOut.>Name"}} = sut.{{.Name}}({{fmtArgCSV $method.CallArgList "in.>Name"}})

			encodedOut, err := testee.Codec.Encode(gotOut)
			if err != nil {
				t.Errorf("err when encoding Out: %v", err)
			}

			// Using "%+v" instead of "%#v" as a workaround for https://github.com/go-yaml/yaml/issues/139.
			if fmt.Sprintf("%+v", encodedOut) != fmt.Sprintf("%+v", tt.wantOut) {
				t.Errorf("Out: Got (%+v) != Want (%+v)", encodedOut, tt.wantOut)
			}

			for _, want := range tt.wantData {
				gotResult := f.Query(want.Query)
				if !gotResult.Equal(want.Result) {
					t.Errorf("Result: Got (%#v) != Want (%#v)", gotResult, want.Result)
				}
			}

			if len(tt.wantData) > 0 {
				// This is an unsafe test, reset the fixture.
				f.Reset()
			}
		})
	}
}
{{- end}} {{/* range $.Tests */}}
`
