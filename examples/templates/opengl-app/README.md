# {{.Name}}

{{if .Description}}{{.Description}}{{else}}Projeto OpenGL básico gerado pelo cpp-gen.{{end}}

- **Padrão C++:** C++{{.Standard}}
- **Dependências:** GLFW (janela/contexto) + GLEW (carregador de funções), via {{if .UseVCPKG}}VCPKG (vcpkg.json){{else if .UseFetchContent}}CMake FetchContent{{else}}pacotes do sistema{{end}}

Renderiza um triângulo colorido usando OpenGL 3.3 core profile — pipeline programável
completo (VAO, VBO, vertex/fragment shader) como ponto de partida pra qualquer coisa mais
elaborada.

## Build
{{if .UseVCPKG}}
Requer a variável de ambiente `VCPKG_ROOT` apontando para uma instalação do VCPKG.

```bash
cmake --preset vcpkg-debug
cmake --build --preset build-vcpkg-debug
./build/vcpkg-debug/bin/{{.Name}}
```
{{else if .UseFetchContent}}
Nenhuma dependência externa é necessária — o CMake baixa e compila GLFW e GLEW
automaticamente na primeira configuração (pode demorar um pouco).

```bash
cmake --preset debug
cmake --build --preset build-debug
./build/debug/bin/{{.Name}}
```
{{else}}
Instale GLFW e GLEW pelo gerenciador de pacotes do seu sistema antes de configurar:

```bash
# macOS
brew install glfw glew

# Debian/Ubuntu
sudo apt install libglfw3-dev libglew-dev
```

```bash
cmake --preset debug
cmake --build --preset build-debug
./build/debug/bin/{{.Name}}
```
{{end}}
## Sobre este template

Este é um dos templates de exemplo para o sistema de templates externos do
[cpp-gen](https://github.com/matpdev/cpp-gen). Já vem registrado no catálogo embutido do
binário — funciona direto:

```bash
cpp-gen new meu-app --template opengl-app --pkg vcpkg
# ou
cpp-gen new meu-app --template opengl-app --pkg fetchcontent
```

Veja `docs/templates.md` no repositório do cpp-gen para a referência completa de como este
template foi construído.
