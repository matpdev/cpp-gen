# {{.Name}}

{{if .Description}}{{.Description}}{{else}}Projeto gerado a partir de um template customizado do cpp-gen.{{end}}

- **Padrão C++:** C++{{.Standard}}
- **Dependências:** {{if .UseVCPKG}}VCPKG (vcpkg.json){{else if .UseFetchContent}}CMake FetchContent{{else}}nenhuma{{end}}

## Build

```bash
cmake -B build -DCMAKE_BUILD_TYPE=Debug
cmake --build build
./build/bin/{{.Name}}
```

---

*Este README é parte do template — edite `files/README.md` no seu template pra mudar o que
aparece aqui em todo projeto gerado a partir dele.*
