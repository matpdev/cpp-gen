// Package generator contém toda a lógica de geração de projetos C++ do cpp-gen.
//
// O pacote é organizado em sub-módulos especializados:
//   - cmake.go      — Geração de CMakeLists.txt, CMakePresets.json e helpers cmake/
//   - structure.go  — Criação da estrutura de pastas e arquivos fonte iniciais
//   - git.go        — Inicialização do repositório Git e geração de .gitignore / README
//   - clang.go      — Geração de .clangd e .clang-format
//   - ide/          — Configurações específicas de IDE (VSCode, CLion, Neovim)
//   - packages/     — Integração com VCPKG e FetchContent
//
// O ponto de entrada público é Generator, criado via New() e executado via Generate().
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"

	"cpp-gen/internal/config"
	"cpp-gen/internal/generator/ide"
	"cpp-gen/internal/generator/packages"
)

// ─────────────────────────────────────────────────────────────────────────────
// TemplateData — dados passados para todos os templates de geração
// ─────────────────────────────────────────────────────────────────────────────

// TemplateData centraliza todas as variáveis disponíveis nos templates Go
// (text/template) usados para gerar os arquivos do projeto.
//
// Além dos campos diretos da ProjectConfig, inclui formas derivadas do nome
// (NameUpper, NameSnake, NamePascal) e flags booleanas para simplificar
// a lógica condicional dentro dos templates ({{if .IsVSCode}}, etc.).
type TemplateData struct {
	// ── Metadados ──────────────────────────────────────────────────────────────

	// Name é o nome original do projeto, exatamente como digitado (ex: "meu-projeto").
	Name string

	// NameUpper é o nome em UPPER_SNAKE_CASE, usado em variáveis CMake e include guards.
	// Exemplo: "meu-projeto" → "MEU_PROJETO"
	NameUpper string

	// NameSnake é o nome em snake_case, usado em nomes de arquivo e funções C++.
	// Exemplo: "meu-projeto" → "meu_projeto"
	NameSnake string

	// NamePascal é o nome em PascalCase, usado em nomes de classe e namespace C++.
	// Exemplo: "meu-projeto" → "MeuProjeto"
	NamePascal string

	// Description é a descrição do projeto fornecida pelo usuário.
	Description string

	// Author é o nome do autor ou organização.
	Author string

	// Version é a versão inicial no formato SemVer (ex: "1.0.0").
	Version string

	// Year é o ano atual, usado em cabeçalhos de copyright e README.
	Year string

	// ── Configurações técnicas ─────────────────────────────────────────────────

	// Standard é o padrão C++ como string numérica (ex: "20").
	Standard string

	// ── Flags booleanas de tipo de projeto ────────────────────────────────────
	// Derivadas de config.ProjectType para simplificar os templates.

	IsExecutable bool // true se TypeExecutable
	IsStaticLib  bool // true se TypeStaticLib
	IsHeaderOnly bool // true se TypeHeaderOnly

	// ── Flags booleanas de gerenciador de pacotes ─────────────────────────────

	UseVCPKG        bool // true se PkgVCPKG
	UseFetchContent bool // true se PkgFetchContent

	// ── Flags booleanas de IDE ────────────────────────────────────────────────

	IsVSCode bool // true se IDEVSCode
	IsCLion  bool // true se IDECLion
	IsNvim   bool // true se IDENvim
	IsZed    bool // true se IDEZed

	// ── Flags de ferramentas opcionais ────────────────────────────────────────

	UseGit         bool // inicializar repositório Git
	UseClangd      bool // gerar .clangd
	UseClangFormat bool // gerar .clang-format
}

// ─────────────────────────────────────────────────────────────────────────────
// Generator
// ─────────────────────────────────────────────────────────────────────────────

// Generator é a estrutura principal que coordena a geração de todos os
// artefatos de um projeto C++. Deve ser criado com New() e executado com Generate().
type Generator struct {
	// cfg contém a configuração original fornecida pelo usuário.
	cfg *config.ProjectConfig

	// data são os dados derivados de cfg, prontos para uso nos templates.
	data *TemplateData

	// root é o caminho absoluto do diretório raiz do projeto a ser criado.
	root string

	// verbose ativa a exibição de cada arquivo gerado durante o processo.
	verbose bool

	// steps acumula as linhas de log de cada etapa para exibição final.
	steps []stepResult
}

// stepResult representa o resultado de uma etapa de geração.
type stepResult struct {
	label   string // descrição curta da etapa (ex: "Estrutura de pastas")
	success bool   // true se completou sem erros
	err     error  // erro ocorrido, nil se success == true
}

// ─────────────────────────────────────────────────────────────────────────────
// Construtor
// ─────────────────────────────────────────────────────────────────────────────

// New cria um novo Generator a partir da ProjectConfig e da flag de verbose.
//
// Deriva automaticamente o TemplateData (formas do nome, flags booleanas, etc.)
// e calcula o caminho raiz do projeto a ser gerado.
func New(cfg *config.ProjectConfig, verbose bool) *Generator {
	data := buildTemplateData(cfg)
	root := cfg.ProjectPath()

	return &Generator{
		cfg:     cfg,
		data:    data,
		root:    root,
		verbose: verbose,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Generate — orquestrador principal
// ─────────────────────────────────────────────────────────────────────────────

// Generate executa todas as etapas de geração do projeto na ordem correta:
//
//  1. Cria a estrutura de pastas e arquivos fonte
//  2. Gera os arquivos CMake (CMakeLists.txt, presets, helpers)
//  3. Configura o gerenciador de pacotes (VCPKG ou FetchContent)
//  4. Gera as configurações da IDE escolhida
//  5. Gera .clangd e/ou .clang-format
//  6. Inicializa o repositório Git e gera .gitignore / README
//
// Ao final, imprime um relatório de todas as etapas executadas.
// Se qualquer etapa crítica falhar, a geração é interrompida imediatamente.
func (g *Generator) Generate() error {
	fmt.Printf("\n  Gerando projeto %q em %q...\n\n", g.cfg.Name, g.root)

	// As etapas são executadas em sequência; cada uma registra seu resultado.
	pipeline := []struct {
		label string
		fn    func() error
	}{
		{"Estrutura de pastas e arquivos fonte", g.runStructure},
		{"Arquivos CMake", g.runCMake},
		{"Gerenciador de pacotes", g.runPackages},
		{"Configuração da IDE", g.runIDE},
		{"Ferramentas Clang", g.runClang},
		{"Git e metadados do repositório", g.runGit},
	}

	for _, step := range pipeline {
		err := step.fn()
		g.steps = append(g.steps, stepResult{
			label:   step.label,
			success: err == nil,
			err:     err,
		})

		if err != nil {
			g.printStepReport()
			return fmt.Errorf("falha em %q: %w", step.label, err)
		}
	}

	g.printStepReport()
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Etapas do pipeline
// ─────────────────────────────────────────────────────────────────────────────

// runStructure cria a hierarquia de diretórios e os arquivos fonte C++ iniciais.
func (g *Generator) runStructure() error {
	return generateStructure(g.root, g.data, g.verbose)
}

// runCMake gera todos os arquivos CMake do projeto:
//   - CMakeLists.txt raiz
//   - src/CMakeLists.txt
//   - tests/CMakeLists.txt
//   - cmake/CompilerWarnings.cmake
//   - CMakePresets.json
func (g *Generator) runCMake() error {
	return generateCMake(g.root, g.data, g.verbose)
}

// runPackages configura o gerenciador de pacotes escolhido.
// Se nenhum foi selecionado (PkgNone), a etapa é pulada silenciosamente.
func (g *Generator) runPackages() error {
	switch g.cfg.PackageManager {
	case config.PkgVCPKG:
		return packages.GenerateVCPKG(g.root, g.verbose)
	case config.PkgFetchContent:
		return packages.GenerateFetchContent(g.root, g.verbose)
	default:
		// Nenhum gerenciador selecionado — nada a fazer.
		return nil
	}
}

// runIDE gera as configurações específicas da IDE escolhida.
// Se IDENone foi selecionado, a etapa é pulada silenciosamente.
func (g *Generator) runIDE() error {
	ideData := &ide.Data{
		ProjectName:  g.data.Name,
		NameUpper:    g.data.NameUpper,
		IsExecutable: g.data.IsExecutable,
		UseVCPKG:     g.data.UseVCPKG,
	}

	switch g.cfg.IDE {
	case config.IDEVSCode:
		return ide.GenerateVSCode(g.root, ideData, g.verbose)
	case config.IDECLion:
		return ide.GenerateCLion(g.root, ideData, g.verbose)
	case config.IDENvim:
		return ide.GenerateNvim(g.root, ideData, g.verbose)
	case config.IDEZed:
		return ide.GenerateZed(g.root, ideData, g.verbose)
	default:
		return nil
	}
}

// runClang gera os arquivos de configuração das ferramentas Clang:
//   - .clangd  (se UseClangd == true)
//   - .clang-format (se UseClangFormat == true)
func (g *Generator) runClang() error {
	return generateClang(g.root, g.data, g.verbose)
}

// runGit inicializa o repositório Git, gera .gitignore e README.md.
// Se UseGit == false, apenas o README é criado (sem git init).
func (g *Generator) runGit() error {
	return generateGit(g.root, g.data, g.verbose)
}

// ─────────────────────────────────────────────────────────────────────────────
// Relatório de etapas
// ─────────────────────────────────────────────────────────────────────────────

// printStepReport imprime na saída padrão um resumo de todas as etapas
// executadas, indicando sucesso ou falha com ícones visuais.
func (g *Generator) printStepReport() {
	checkOK := "  ✓"
	checkFail := "  ✗"

	for _, s := range g.steps {
		if s.success {
			fmt.Printf("%s  %s\n", checkOK, s.label)
		} else {
			fmt.Printf("%s  %s — %v\n", checkFail, s.label, s.err)
		}
	}
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────────────────────
// buildTemplateData — derivação dos dados do template
// ─────────────────────────────────────────────────────────────────────────────

// buildTemplateData converte uma ProjectConfig em TemplateData, calculando
// todas as formas derivadas do nome e as flags booleanas necessárias
// para a lógica condicional dos templates.
func buildTemplateData(cfg *config.ProjectConfig) *TemplateData {
	return &TemplateData{
		// Formas do nome
		Name:       cfg.Name,
		NameUpper:  toUpperSnake(cfg.Name),
		NameSnake:  toSnakeCase(cfg.Name),
		NamePascal: toPascalCase(cfg.Name),

		// Metadados
		Description: cfg.Description,
		Author:      cfg.Author,
		Version:     cfg.Version,
		Year:        fmt.Sprintf("%d", time.Now().Year()),

		// Técnicos
		Standard: string(cfg.Standard),

		// Tipo de projeto
		IsExecutable: cfg.ProjectType == config.TypeExecutable,
		IsStaticLib:  cfg.ProjectType == config.TypeStaticLib,
		IsHeaderOnly: cfg.ProjectType == config.TypeHeaderOnly,

		// Gerenciadores de pacotes
		UseVCPKG:        cfg.PackageManager == config.PkgVCPKG,
		UseFetchContent: cfg.PackageManager == config.PkgFetchContent,

		// IDEs
		IsVSCode: cfg.IDE == config.IDEVSCode,
		IsCLion:  cfg.IDE == config.IDECLion,
		IsNvim:   cfg.IDE == config.IDENvim,
		IsZed:    cfg.IDE == config.IDEZed,

		// Ferramentas opcionais
		UseGit:         cfg.UseGit,
		UseClangd:      cfg.UseClangd,
		UseClangFormat: cfg.UseClangFormat,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilitários de transformação de nomes
// ─────────────────────────────────────────────────────────────────────────────

// toUpperSnake converte um nome de projeto para UPPER_SNAKE_CASE.
//
// Exemplos:
//
//	"meu-projeto"  → "MEU_PROJETO"
//	"my.lib.core"  → "MY_LIB_CORE"
func toUpperSnake(name string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	return strings.ToUpper(replacer.Replace(name))
}

// toSnakeCase converte um nome de projeto para snake_case.
//
// Exemplos:
//
//	"meu-projeto" → "meu_projeto"
//	"My-Lib"      → "my_lib"
func toSnakeCase(name string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	return strings.ToLower(replacer.Replace(name))
}

// toPascalCase converte um nome de projeto para PascalCase.
// Delimitadores reconhecidos: hífen, underscore, ponto e espaço.
//
// Exemplos:
//
//	"meu-projeto"  → "MeuProjeto"
//	"my_lib_core"  → "MyLibCore"
func toPascalCase(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == '.' || unicode.IsSpace(r)
	})

	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}

	return strings.Join(parts, "")
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilitários de I/O compartilhados entre os sub-geradores
// ─────────────────────────────────────────────────────────────────────────────

// writeFile cria (ou sobrescreve) um arquivo no caminho dado com o conteúdo
// fornecido. Cria todos os diretórios pai necessários automaticamente.
// Se verbose for true, imprime o caminho do arquivo criado.
func writeFile(path, content string, verbose bool) error {
	// Garante que o diretório pai existe
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

// renderTemplate processa um template Go (text/template) com os dados fornecidos
// e retorna o resultado como string. Retorna erro se o template for inválido
// ou se os dados não satisfizerem os campos referenciados.
func renderTemplate(name, tmpl string, data any) (string, error) {
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

// writeTemplate é uma combinação de renderTemplate + writeFile:
// processa o template e grava o resultado no arquivo indicado.
func writeTemplate(path, name, tmpl string, data any, verbose bool) error {
	content, err := renderTemplate(name, tmpl, data)
	if err != nil {
		return err
	}
	return writeFile(path, content, verbose)
}
