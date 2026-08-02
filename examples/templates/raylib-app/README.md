# {{.Name}}

{{if .Description}}{{.Description}}{{else}}Projeto raylib gerado pelo cpp-gen.{{end}}

- **Padrão C++:** C++{{.Standard}}
- **Dependências:** {{if .UseVCPKG}}VCPKG (vcpkg.json){{else if .UseFetchContent}}CMake FetchContent{{else}}pacote raylib do sistema{{end}}

## Build
{{if .UseVCPKG}}
Requer a variável de ambiente `VCPKG_ROOT` apontando para uma instalação do VCPKG.

```bash
cmake --preset vcpkg-debug
cmake --build --preset build-vcpkg-debug
./build/vcpkg-debug/bin/{{.Name}}
```
{{else if .UseFetchContent}}
Nenhuma dependência externa é necessária — o CMake baixa e compila o raylib
automaticamente na primeira configuração (pode demorar um pouco).

```bash
cmake --preset debug
cmake --build --preset build-debug
./build/debug/bin/{{.Name}}
```
{{else}}
Instale o raylib pelo gerenciador de pacotes do seu sistema antes de configurar:

```bash
# macOS
brew install raylib

# Debian/Ubuntu
sudo apt install libraylib-dev
```

```bash
cmake --preset debug
cmake --build --preset build-debug
./build/debug/bin/{{.Name}}
```
{{end}}
## Sobre este template

Este é um template de exemplo para o sistema de templates externos do
[cpp-gen](https://github.com/matpdev/cpp-gen). Ele demonstra:

- Substituição de variáveis do projeto em conteúdo e nomes de arquivo
  (`{{"{{"}}.Name{{"}}"}}`, `{{"{{"}}.NameSnake{{"}}"}}`, etc.).
- Inclusão condicional de arquivos inteiros conforme as escolhas do usuário
  (`vcpkg.json` e `vcpkg-configuration.json` só existem quando VCPKG é
  selecionado).
- CMake que se adapta ao gerenciador de pacotes escolhido
  (`--pkg vcpkg`, `--pkg fetchcontent` ou `--pkg none`).

Para usar:

```bash
cpp-gen new meu-jogo --template ./examples/templates/raylib-app --pkg vcpkg
# ou
cpp-gen new meu-jogo --template ./examples/templates/raylib-app --pkg fetchcontent
```
