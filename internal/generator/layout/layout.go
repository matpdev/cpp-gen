// Package layout define as especificações de estrutura de pastas para projetos C++.
//
// Constantes de kind exportadas para uso por outros pacotes sem importar config:
//
//	layout.KindSeparate, layout.KindMerged, layout.KindFlat, layout.KindModular, layout.KindTwoRoot
//
// Cada layout (padrão de organização) determina:
//   - Quais diretórios criar
//   - Onde ficam os arquivos fonte, headers e testes
//   - Quais diretórios de include o CMake usa em target_include_directories()
//   - O prefixo usado nas diretivas #include dos arquivos C++ gerados
//
// Layouts disponíveis:
//
//	separate  — include/<nome>/ + src/ (clássico CMake)
//	merged    — <nome>/ com headers e sources juntos (Pitchfork / P1204R0)
//	flat      — tudo em src/ sem separação
//	modular   — libs/<nome>/ para multi-módulos (Pitchfork libs/)
//	two-root  — include/ + src/ sem subdiretório de namespace
package layout

import (
	"fmt"
	"path/filepath"
	"strings"

	"cpp-gen/internal/config"
)

// ─────────────────────────────────────────────────────────────────────────────
// Constantes de kind exportadas
// ─────────────────────────────────────────────────────────────────────────────
// Aliases das constantes config.FolderLayout expostos diretamente pelo pacote
// layout para que outros pacotes (ex: structure.go) possam comparar spec.Kind
// sem precisar importar o pacote config separadamente.

const (
	KindSeparate = config.LayoutSeparate
	KindMerged   = config.LayoutMerged
	KindFlat     = config.LayoutFlat
	KindModular  = config.LayoutModular
	KindTwoRoot  = config.LayoutTwoRoot
)

// ─────────────────────────────────────────────────────────────────────────────
// Spec — especificação completa de um layout
// ─────────────────────────────────────────────────────────────────────────────

// Spec descreve todos os aspectos parametrizáveis de um layout de pastas C++.
// É calculada uma vez por projeto em Resolve() e usada por todos os geradores
// (structure.go, cmake.go) para saber onde criar arquivos e como configurar o CMake.
type Spec struct {
	// Kind é o identificador do layout (ex: "separate", "merged").
	Kind config.FolderLayout

	// ── Diretórios a criar ────────────────────────────────────────────────────

	// Dirs lista todos os diretórios a criar (relativos à raiz do projeto).
	// Inclui todos os intermediários necessários — os geradores usam MkdirAll.
	Dirs []string

	// ── Caminhos dos arquivos gerados ─────────────────────────────────────────
	// Todos os caminhos são relativos à raiz do projeto.

	// MainCPP é o caminho para o arquivo de entrada de executáveis (main.cpp).
	MainCPP string

	// LibCPP é o caminho para o arquivo de implementação de bibliotecas.
	LibCPP string

	// PublicHPP é o caminho para o header público principal da biblioteca.
	PublicHPP string

	// TestCPP é o caminho para o arquivo de testes principal.
	TestCPP string

	// ── Configuração CMake ────────────────────────────────────────────────────

	// CMakeSubdir é o subdiretório passado para add_subdirectory() no
	// CMakeLists.txt raiz para o target principal (ex: "src", "mylib", "libs/mylib").
	CMakeSubdir string

	// CMakeTestsDir é o subdiretório dos testes (quase sempre "tests").
	CMakeTestsDir string

	// CMakeIncludeBlock é o bloco pré-formatado de target_include_directories()
	// para o target principal, pronto para inserção direta no template CMake.
	// Exemplo:
	//     PUBLIC
	//         $<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/include>
	//         $<INSTALL_INTERFACE:include>
	//     PRIVATE
	//         ${CMAKE_CURRENT_SOURCE_DIR}
	CMakeIncludeBlock string

	// CMakeTestIncludeBlock é o bloco de include dirs para o target de testes.
	CMakeTestIncludeBlock string

	// CMakeModularLibDir é o diretório da biblioteca no layout modular.
	// Preenchido apenas quando Kind == LayoutModular; vazio nos demais.
	CMakeModularLibDir string

	// ── Include C++ ───────────────────────────────────────────────────────────

	// IncludePrefix é o prefixo de diretório usado nas diretivas #include
	// dos arquivos C++ gerados.
	//
	// Exemplos:
	//   "mylib/"   → #include "mylib/mylib.hpp"   (separate, merged, modular)
	//   ""         → #include "mylib.hpp"          (flat, two-root)
	IncludePrefix string

	// ── Notas documentais ─────────────────────────────────────────────────────

	// LayoutNote é um comentário inserido no CMakeLists.txt gerado explicando
	// brevemente a convenção de layout adotada e onde encontrar a documentação.
	LayoutNote string
}

// ─────────────────────────────────────────────────────────────────────────────
// Resolve — fábrica principal
// ─────────────────────────────────────────────────────────────────────────────

// Resolve calcula e retorna o Spec completo para a combinação de nome de projeto,
// tipo de projeto e layout escolhido.
//
// Todos os caminhos de arquivo retornados são relativos à raiz do projeto.
// Os blocos CMake são pré-formatados e prontos para inserção nos templates.
//
// Parâmetros:
//
//	name        — nome do projeto (ex: "minha-lib")
//	nameSnake   — nome em snake_case   (ex: "minha_lib")
//	layout      — constante FolderLayout escolhida
//	projectType — tipo de artefato (executable, static-lib, header-only)
func Resolve(
	name string,
	nameSnake string,
	layout config.FolderLayout,
	projectType config.ProjectType,
) *Spec {
	switch layout {
	case config.LayoutMerged:
		return resolveMerged(name, nameSnake, projectType)
	case config.LayoutFlat:
		return resolveFlat(name, nameSnake, projectType)
	case config.LayoutModular:
		return resolveModular(name, nameSnake, projectType)
	case config.LayoutTwoRoot:
		return resolveTwoRoot(name, nameSnake, projectType)
	default: // config.LayoutSeparate — padrão e fallback
		return resolveSeparate(name, nameSnake, projectType)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Layouts individuais
// ─────────────────────────────────────────────────────────────────────────────

// resolveSeparate retorna a spec do layout clássico com separação include/<nome>/ e src/.
//
// Estrutura gerada:
//
//	include/<nome>/  ← headers públicos com prefixo de namespace
//	src/             ← implementações e headers privados
//	tests/           ← testes unitários e de integração
//	cmake/           ← módulos CMake auxiliares
//	docs/            ← documentação
//
// Exemplo de include: #include "<nome>/file.hpp"
// CMake: PUBLIC include, PRIVATE src
func resolveSeparate(name, nameSnake string, pt config.ProjectType) *Spec {
	isHeaderOnly := pt == config.TypeHeaderOnly

	dirs := []string{
		filepath.Join("include", name),
		"cmake",
		"tests",
		"docs",
	}
	if !isHeaderOnly {
		dirs = append(dirs, "src")
	}

	mainCPP := filepath.Join("src", "main.cpp")
	libCPP := filepath.Join("src", nameSnake+".cpp")
	publicHPP := filepath.Join("include", name, nameSnake+".hpp")
	testCPP := filepath.Join("tests", "test_main.cpp")

	includeBlock := buildIncludeBlock(
		[]string{
			"$<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/include>",
			"$<INSTALL_INTERFACE:include>",
		},
		[]string{
			"${CMAKE_CURRENT_SOURCE_DIR}",
		},
		pt,
	)

	testIncludeBlock := indent(4, "${PROJECT_SOURCE_DIR}/include")

	return &Spec{
		Kind:                  config.LayoutSeparate,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           "src",
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		IncludePrefix:         name + "/",
		LayoutNote: fmt.Sprintf(
			"# Layout: Separate (clássico)\n"+
				"# Headers públicos em include/%s/, implementações em src/.\n"+
				"# Referência: padrão amplamente adotado em projetos CMake modernos.",
			name,
		),
	}
}

// resolveMerged retorna a spec do layout Pitchfork / P1204R0 (merged placement).
//
// Headers e implementações ficam no mesmo diretório <nome>/, eliminando a
// necessidade de navegar entre include/ e src/. Unit tests ficam como arquivos
// irmãos (*.<nome>.test.cpp). Tests de integração ficam em tests/.
//
// Estrutura gerada:
//
//	<nome>/           ← headers (.hpp) e fontes (.cpp) juntos
//	<nome>/*.test.cpp ← unit tests ao lado dos módulos (gerado se biblioteca)
//	tests/            ← testes de integração e funcionais
//	cmake/            ← módulos CMake auxiliares
//	docs/             ← documentação
//
// Exemplo de include: #include "<nome>/file.hpp"
// CMake: PUBLIC ${PROJECT_SOURCE_DIR} (raiz como base de includes)
func resolveMerged(name, nameSnake string, pt config.ProjectType) *Spec {
	dirs := []string{
		name,
		"tests",
		"cmake",
		"docs",
	}

	mainCPP := filepath.Join(name, "main.cpp")
	libCPP := filepath.Join(name, nameSnake+".cpp")
	publicHPP := filepath.Join(name, nameSnake+".hpp")
	testCPP := filepath.Join("tests", "driver.cpp")

	// No merged layout, o CMake subdir é o próprio diretório do projeto
	cmakeSubdir := name

	includeBlock := buildIncludeBlock(
		[]string{
			"$<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}>",
			"$<INSTALL_INTERFACE:.>",
		},
		nil,
		pt,
	)

	testIncludeBlock := indent(4, "${PROJECT_SOURCE_DIR}")

	return &Spec{
		Kind:                  config.LayoutMerged,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           cmakeSubdir,
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		IncludePrefix:         name + "/",
		LayoutNote: fmt.Sprintf(
			"# Layout: Merged (Pitchfork / P1204R0)\n"+
				"# Headers e implementações juntos em %s/. Unit tests como arquivos irmãos.\n"+
				"# Referência: https://vector-of-bool.github.io/pitchfork (SG15 P1204R0)",
			name,
		),
	}
}

// resolveFlat retorna a spec do layout plano (tudo em src/).
//
// Headers e implementações ficam juntos em src/ sem nenhuma separação.
// É o layout mais simples, indicado para executáveis e projetos pequenos
// onde a distinção entre API pública e privada não é relevante.
//
// Estrutura gerada:
//
//	src/   ← headers (.hpp) e fontes (.cpp) no mesmo lugar
//	tests/ ← testes
//	cmake/ ← módulos CMake auxiliares
//	docs/  ← documentação
//
// Exemplo de include: #include "file.hpp" (sem prefixo de namespace)
// CMake: PRIVATE src
func resolveFlat(name, nameSnake string, pt config.ProjectType) *Spec {
	dirs := []string{
		"src",
		"tests",
		"cmake",
		"docs",
	}

	mainCPP := filepath.Join("src", "main.cpp")
	libCPP := filepath.Join("src", nameSnake+".cpp")
	publicHPP := filepath.Join("src", nameSnake+".hpp")
	testCPP := filepath.Join("tests", "test_main.cpp")

	// Flat: sem PUBLIC includes — tudo é privado ao target.
	// Para header-only, INTERFACE include aponta para src/.
	var includeBlock string
	if pt == config.TypeHeaderOnly {
		includeBlock = buildIncludeBlock(
			[]string{
				"$<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/src>",
				"$<INSTALL_INTERFACE:src>",
			},
			nil,
			pt,
		)
	} else {
		includeBlock = buildIncludeBlock(
			nil,
			[]string{"${CMAKE_CURRENT_SOURCE_DIR}"},
			pt,
		)
	}

	testIncludeBlock := indent(4, "${PROJECT_SOURCE_DIR}/src")

	return &Spec{
		Kind:                  config.LayoutFlat,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           "src",
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		IncludePrefix:         "",
		LayoutNote: "# Layout: Flat\n" +
			"# Headers e implementações juntos em src/. Ideal para executáveis e projetos simples.\n" +
			"# Referência: convenção common para projetos pequenos e scripts.",
	}
}

// resolveModular retorna a spec do layout multi-módulo Pitchfork com libs/.
//
// Cada biblioteca/módulo vive em seu próprio subdiretório dentro de libs/.
// Executáveis ficam em apps/. Este layout facilita a extração futura de módulos
// em projetos separados e é adequado para projetos grandes com múltiplos componentes.
//
// Estrutura gerada:
//
//	libs/<nome>/include/<nome>/  ← headers públicos
//	libs/<nome>/src/             ← implementações
//	apps/                        ← executáveis (entry points)
//	tests/                       ← testes
//	cmake/                       ← módulos CMake auxiliares
//	docs/                        ← documentação
//
// Exemplo de include: #include "<nome>/file.hpp"
// CMake: PUBLIC libs/<nome>/include, PRIVATE libs/<nome>/src
func resolveModular(name, nameSnake string, pt config.ProjectType) *Spec {
	libDir := filepath.Join("libs", name)

	dirs := []string{
		filepath.Join(libDir, "include", name),
		filepath.Join(libDir, "src"),
		"apps",
		"tests",
		"cmake",
		"docs",
	}

	// Para modular, o executável fica em apps/ e a lib em libs/<nome>/
	mainCPP := filepath.Join("apps", "main.cpp")
	libCPP := filepath.Join(libDir, "src", nameSnake+".cpp")
	publicHPP := filepath.Join(libDir, "include", name, nameSnake+".hpp")
	testCPP := filepath.Join("tests", "test_main.cpp")

	// O CMakeSubdir aponta para a biblioteca. O executável apps/main.cpp
	// é adicionado diretamente no root CMakeLists.txt (ver cmake.go).
	cmakeSubdir := libDir

	// No layout modular, o CMakeLists.txt em libs/<nome>/ SEMPRE define um target
	// de biblioteca estática — mesmo quando o projeto é um executável.
	// O executável (apps/main.cpp) é definido no root CMakeLists.txt e linka
	// à lib. Portanto forçamos TypeStaticLib para gerar PUBLIC/PRIVATE corretos.
	includeBlock := buildIncludeBlock(
		[]string{
			"$<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>",
			"$<INSTALL_INTERFACE:include>",
		},
		[]string{
			"${CMAKE_CURRENT_SOURCE_DIR}/src",
		},
		config.TypeStaticLib, // sempre lib, independente do tipo do projeto
	)

	testIncludeBlock := indent(4,
		fmt.Sprintf("${PROJECT_SOURCE_DIR}/%s/include", libDir),
	)

	return &Spec{
		Kind:                  config.LayoutModular,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           cmakeSubdir,
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		CMakeModularLibDir:    libDir,
		IncludePrefix:         name + "/",
		LayoutNote: fmt.Sprintf(
			"# Layout: Modular (Pitchfork libs/)\n"+
				"# Biblioteca em %s/include/%s/ e %s/src/. Executáveis em apps/.\n"+
				"# Referência: https://vector-of-bool.github.io/pitchfork (libs/ submodules)",
			libDir, name, libDir,
		),
	}
}

// resolveTwoRoot retorna a spec do layout two-root: include/ + src/ sem namespace subdir.
//
// Semelhante ao Separate, mas os headers ficam diretamente em include/*.hpp
// sem criar um subdiretório de namespace (include/<nome>/). Comum em projetos
// menores ou bibliotecas com um único header principal.
//
// Estrutura gerada:
//
//	include/  ← headers públicos (sem subdir de namespace)
//	src/      ← implementações
//	tests/    ← testes
//	cmake/    ← módulos CMake auxiliares
//	docs/     ← documentação
//
// Exemplo de include: #include "file.hpp" (sem prefixo de namespace)
// CMake: PUBLIC include, PRIVATE src
func resolveTwoRoot(name, nameSnake string, pt config.ProjectType) *Spec {
	isHeaderOnly := pt == config.TypeHeaderOnly

	dirs := []string{
		"include",
		"cmake",
		"tests",
		"docs",
	}
	if !isHeaderOnly {
		dirs = append(dirs, "src")
	}

	mainCPP := filepath.Join("src", "main.cpp")
	libCPP := filepath.Join("src", nameSnake+".cpp")
	publicHPP := filepath.Join("include", nameSnake+".hpp")
	testCPP := filepath.Join("tests", "test_main.cpp")

	includeBlock := buildIncludeBlock(
		[]string{
			"$<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/include>",
			"$<INSTALL_INTERFACE:include>",
		},
		[]string{
			"${CMAKE_CURRENT_SOURCE_DIR}",
		},
		pt,
	)

	testIncludeBlock := indent(4, "${PROJECT_SOURCE_DIR}/include")

	return &Spec{
		Kind:                  config.LayoutTwoRoot,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           "src",
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		IncludePrefix:         "",
		LayoutNote: "# Layout: Two-Root\n" +
			"# Headers públicos em include/ (sem subdir de namespace), implementações em src/.\n" +
			"# Referência: convenção simples para bibliotecas com interface mínima.",
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers de construção de blocos CMake
// ─────────────────────────────────────────────────────────────────────────────

// buildIncludeBlock constrói o bloco de argumentos para target_include_directories()
// com base nos diretórios públicos e privados fornecidos, adaptado ao tipo de projeto.
//
// Para projetos header-only (INTERFACE), todo o bloco usa INTERFACE em vez de
// PUBLIC/PRIVATE, pois targets INTERFACE não podem ter propriedades PRIVATE.
//
// Exemplo de saída para biblioteca estática:
//
//	PUBLIC
//	    $<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/include>
//	    $<INSTALL_INTERFACE:include>
//	PRIVATE
//	    ${CMAKE_CURRENT_SOURCE_DIR}
func buildIncludeBlock(public, private []string, pt config.ProjectType) string {
	var sb strings.Builder

	if pt == config.TypeHeaderOnly {
		// Header-only usa INTERFACE para PUBLIC e ignora PRIVATE
		if len(public) > 0 {
			sb.WriteString("    INTERFACE\n")
			for _, p := range public {
				sb.WriteString("        " + p + "\n")
			}
		}
		return strings.TrimRight(sb.String(), "\n")
	}

	// Executável: usa apenas PRIVATE (sem API pública instalável)
	if pt == config.TypeExecutable {
		allPrivate := append(public, private...)
		if len(allPrivate) > 0 {
			sb.WriteString("    PRIVATE\n")
			for _, p := range allPrivate {
				sb.WriteString("        " + p + "\n")
			}
		}
		return strings.TrimRight(sb.String(), "\n")
	}

	// Biblioteca estática: PUBLIC para a API, PRIVATE para implementação
	if len(public) > 0 {
		sb.WriteString("    PUBLIC\n")
		for _, p := range public {
			sb.WriteString("        " + p + "\n")
		}
	}
	if len(private) > 0 {
		sb.WriteString("    PRIVATE\n")
		for _, p := range private {
			sb.WriteString("        " + p + "\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// indent retorna a string com n espaços de indentação à esquerda.
// Usado para formatar diretórios de include no bloco de tests CMakeLists.txt.
func indent(n int, s string) string {
	return strings.Repeat(" ", n) + s
}

// ─────────────────────────────────────────────────────────────────────────────
// Relatório legível do layout (para diagnóstico e --verbose)
// ─────────────────────────────────────────────────────────────────────────────

// Summary retorna uma string multi-linha com um resumo legível do Spec,
// útil para exibição em modo verbose ou logs de diagnóstico.
func (s *Spec) Summary() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Layout:        %s\n", s.Kind))
	sb.WriteString(fmt.Sprintf("CMake subdir:  %s\n", s.CMakeSubdir))
	sb.WriteString(fmt.Sprintf("Include prefix: %q\n", s.IncludePrefix))
	sb.WriteString("Directories:\n")
	for _, d := range s.Dirs {
		sb.WriteString(fmt.Sprintf("  %s/\n", d))
	}

	return sb.String()
}
