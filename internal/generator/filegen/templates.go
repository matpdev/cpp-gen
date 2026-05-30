package filegen

import (
	"bytes"
	"text/template"
)

// FileType represents the kind of C++ artifact to generate.
type FileType string

const (
	FileTypeClass     FileType = "class"
	FileTypeStruct    FileType = "struct"
	FileTypeFree      FileType = "free"      // free functions (hpp + cpp)
	FileTypeInterface FileType = "interface" // abstract base class (hpp only)
)

// FileTemplateData holds the data passed to all file generation templates.
type FileTemplateData struct {
	Header       string // resultado de BuildHeader (já formatado)
	Name         string // nome do arquivo/classe em PascalCase, ex: "FooBar"
	NameSnake    string // snake_case, ex: "foo_bar"
	NameUpper    string // UPPER_SNAKE, ex: "FOO_BAR"
	Namespace    string // namespace name, empty if disabled
	HasNamespace bool
	IncludePath  string // caminho relativo para o #include no .cpp, ex: "myproject/foo_bar.hpp"
	ProjectName  string
}

const tmplClassHpp = `{{- if .Header}}{{.Header}}

{{end -}}
#pragma once
{{if .HasNamespace}}
namespace {{.Namespace}} {
{{end}}
class {{.Name}} {
public:
    {{.Name}}();
    ~{{.Name}}();

private:
};
{{- if .HasNamespace}}

} // namespace {{.Namespace}}
{{- end}}`

const tmplClassCpp = `{{- if .Header}}{{.Header}}

{{end -}}
#include "{{.IncludePath}}"
{{if .HasNamespace}}
namespace {{.Namespace}} {
{{end}}
{{.Name}}::{{.Name}}() = default;
{{.Name}}::~{{.Name}}() = default;
{{- if .HasNamespace}}

} // namespace {{.Namespace}}
{{- end}}`

const tmplStructHpp = `{{- if .Header}}{{.Header}}

{{end -}}
#pragma once
{{if .HasNamespace}}
namespace {{.Namespace}} {
{{end}}
struct {{.Name}} {
};
{{- if .HasNamespace}}

} // namespace {{.Namespace}}
{{- end}}`

const tmplFreeHpp = `{{- if .Header}}{{.Header}}

{{end -}}
#pragma once
{{if .HasNamespace}}
namespace {{.Namespace}} {
{{end}}
// TODO: declare free functions here
{{- if .HasNamespace}}

} // namespace {{.Namespace}}
{{- end}}`

const tmplFreeCpp = `{{- if .Header}}{{.Header}}

{{end -}}
#include "{{.IncludePath}}"
{{if .HasNamespace}}
namespace {{.Namespace}} {
{{end}}
// TODO: implement free functions here
{{- if .HasNamespace}}

} // namespace {{.Namespace}}
{{- end}}`

const tmplInterfaceHpp = `{{- if .Header}}{{.Header}}

{{end -}}
#pragma once
{{if .HasNamespace}}
namespace {{.Namespace}} {
{{end}}
class {{.Name}} {
public:
    virtual ~{{.Name}}() = default;

    // TODO: declare pure virtual methods here
};
{{- if .HasNamespace}}

} // namespace {{.Namespace}}
{{- end}}`

// renderFileTemplate executes a template string with the provided data.
func renderFileTemplate(tmpl string, data FileTemplateData) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
