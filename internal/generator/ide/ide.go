// Package ide contém os geradores de configuração de IDE para projetos C++.
//
// Cada IDE suportada possui seu próprio arquivo de implementação:
//   - vscode.go  — Visual Studio Code (.vscode/)
//   - clion.go   — CLion (configurações adicionais ao CMakePresets.json)
//   - nvim.go    — Neovim (.nvim.lua)
//
// Todas as funções de geração recebem um *Data com as informações necessárias
// e o caminho raiz do projeto onde os arquivos serão criados.
package ide

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ─────────────────────────────────────────────────────────────────────────────
// Data — estrutura de dados compartilhada entre todos os geradores de IDE
// ─────────────────────────────────────────────────────────────────────────────

// Data contém as informações necessárias para gerar as configurações
// de qualquer IDE suportada pelo cpp-gen.
//
// É um subconjunto do generator.TemplateData, exposto como tipo próprio
// para evitar dependência circular entre os pacotes generator e ide.
type Data struct {
	// ProjectName é o nome original do projeto (ex: "meu-projeto").
	// Usado em labels, títulos de tasks e nomes de executável.
	ProjectName string

	// NameUpper é o nome em UPPER_SNAKE_CASE (ex: "MEU_PROJETO").
	// Usado em variáveis de ambiente e defines CMake.
	NameUpper string

	// IsExecutable indica se o projeto gera um binário executável.
	// Quando true, as configurações de launch/debug apontam para o binário.
	IsExecutable bool

	// UseVCPKG indica se o projeto usa VCPKG como gerenciador de pacotes.
	// Quando true, os presets de debug/build usam os presets vcpkg-*.
	UseVCPKG bool
}

// ─────────────────────────────────────────────────────────────────────────────
// Seletores de preset CMake
// ─────────────────────────────────────────────────────────────────────────────

// ConfigurePreset retorna o nome do preset CMake de configure a ser usado
// como padrão nas configurações de build das IDEs.
//
// Se VCPKG estiver habilitado, retorna o preset "vcpkg-debug".
// Caso contrário, retorna "debug".
func (d *Data) ConfigurePreset() string {
	if d.UseVCPKG {
		return "vcpkg-debug"
	}
	return "debug"
}

// BuildPreset retorna o nome do preset CMake de build a ser usado
// nas tasks de build das IDEs.
func (d *Data) BuildPreset() string {
	if d.UseVCPKG {
		return "build-vcpkg-debug"
	}
	return "build-debug"
}

// TestPreset retorna o nome do preset CMake de test a ser usado
// nas tasks de teste das IDEs.
func (d *Data) TestPreset() string {
	if d.UseVCPKG {
		return "test-vcpkg-debug"
	}
	return "test-debug"
}

// BinaryPath retorna o caminho relativo ao executável gerado pelo build,
// baseado no preset de debug ativo.
//
// Usado em configurações de launch/debug para apontar para o binário correto.
// O binário é colocado em build/<preset>/bin/<nome> conforme definido no CMakeLists.txt.
func (d *Data) BinaryPath() string {
	preset := d.ConfigurePreset()
	return fmt.Sprintf("${workspaceFolder}/build/%s/bin/%s", preset, d.ProjectName)
}

// ─────────────────────────────────────────────────────────────────────────────
// Funções de geração de alto nível
// ─────────────────────────────────────────────────────────────────────────────

// GenerateVSCode gera as configurações do Visual Studio Code para o projeto.
// Cria o diretório .vscode/ com os arquivos:
//   - tasks.json       — tarefas de build, clean e test
//   - launch.json      — configurações de debug (lldb e cppdbg)
//   - settings.json    — configurações do workspace (clangd, cmake tools)
//   - extensions.json  — extensões recomendadas para o projeto
func GenerateVSCode(root string, data *Data, verbose bool) error {
	return generateVSCode(root, data, verbose)
}

// GenerateCLion gera as configurações adicionais para CLion.
// O CLion lê nativamente o CMakePresets.json (já gerado pelo cmake.go),
// portanto este gerador cria apenas um arquivo .idea/externalTools.xml
// com integrações adicionais de ferramentas.
func GenerateCLion(root string, data *Data, verbose bool) error {
	return generateCLion(root, data, verbose)
}

// GenerateNvim gera as configurações de projeto para Neovim.
// Cria um arquivo .nvim.lua na raiz do projeto com configurações
// de LSP, keymaps de projeto e integração com o DAP para debug.
func GenerateNvim(root string, data *Data, verbose bool) error {
	return generateNvim(root, data, verbose)
}

// GenerateZed gera as configurações do editor Zed para o projeto.
// Cria o diretório .zed/ com os arquivos:
//   - settings.json  — configurações do workspace (clangd LSP, clang-format,
//     editor, inlay hints e debug adapter)
//   - tasks.json     — tarefas de configure, build, clean, test e run
func GenerateZed(root string, data *Data, verbose bool) error {
	return generateZed(root, data, verbose)
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilitários internos compartilhados entre os geradores de IDE
// ─────────────────────────────────────────────────────────────────────────────

// writeIDEFile cria (ou sobrescreve) um arquivo no caminho especificado.
// Cria todos os diretórios pai necessários com permissão 0755.
// Se verbose for true, imprime o caminho do arquivo criado.
func writeIDEFile(path, content string, verbose bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("criar diretório %q: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("escrever arquivo %q: %w", path, err)
	}

	if verbose {
		fmt.Printf("    + %s\n", path)
	}

	return nil
}

// renderIDETemplate processa um template Go (text/template) com os dados
// fornecidos e retorna o resultado como string.
func renderIDETemplate(name, tmpl string, data any) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse do template %q: %w", name, err)
	}

	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("execução do template %q: %w", name, err)
	}

	return sb.String(), nil
}

// writeIDETemplate combina renderIDETemplate e writeIDEFile:
// processa o template e grava o resultado no arquivo indicado.
func writeIDETemplate(path, name, tmpl string, data any, verbose bool) error {
	content, err := renderIDETemplate(name, tmpl, data)
	if err != nil {
		return err
	}
	return writeIDEFile(path, content, verbose)
}
