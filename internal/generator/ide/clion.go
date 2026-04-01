// Package ide contém os geradores de configuração de IDE para projetos C++.
package ide

import (
	"fmt"
	"path/filepath"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateCLion — ponto de entrada
// ─────────────────────────────────────────────────────────────────────────────

// generateCLion gera as configurações adicionais para o CLion da JetBrains.
//
// O CLion possui suporte nativo ao CMakePresets.json (já gerado pelo cmake.go),
// portanto este gerador complementa com:
//
//   - .idea/runConfigurations/  — configurações de run/debug exportáveis
//   - .idea/cmake.xml           — configuração de perfis CMake do CLion
//   - .idea/inspectionProfiles/ — perfil de inspeção de código customizado
//   - .idea/.gitignore          — ignora arquivos gerados pelo CLion no .idea/
//
// NOTA: O arquivo CMakePresets.json gerado pelo cmake.go é o principal artefato
// para a integração CLion. Este gerador adiciona conveniências opcionais.
//
// Compatibilidade: CLion 2022.1+ (suporte a CMakePresets.json).
// Referência: https://www.jetbrains.com/help/clion/cmake-presets.html
func generateCLion(root string, data *Data, verbose bool) error {
	ideaDir := filepath.Join(root, ".idea")

	steps := []struct {
		name     string
		relPath  string
		tmplName string
		tmpl     string
	}{
		{
			name:     ".idea/.gitignore",
			relPath:  filepath.Join(ideaDir, ".gitignore"),
			tmplName: "clion_idea_gitignore",
			tmpl:     tmplCLionIdeaGitignore,
		},
		{
			name:     ".idea/cmake.xml",
			relPath:  filepath.Join(ideaDir, "cmake.xml"),
			tmplName: "clion_cmake_xml",
			tmpl:     tmplCLionCMakeXML,
		},
		{
			name:     ".idea/inspectionProfiles/Project_Default.xml",
			relPath:  filepath.Join(ideaDir, "inspectionProfiles", "Project_Default.xml"),
			tmplName: "clion_inspection_profile",
			tmpl:     tmplCLionInspectionProfile,
		},
		{
			name:     ".idea/runConfigurations/Debug.xml",
			relPath:  filepath.Join(ideaDir, "runConfigurations", "Debug.xml"),
			tmplName: "clion_run_debug",
			tmpl:     tmplCLionRunDebug,
		},
	}

	for _, s := range steps {
		if err := writeIDETemplate(s.relPath, s.tmplName, s.tmpl, data, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", s.name, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// generateNvim — ponto de entrada
// ─────────────────────────────────────────────────────────────────────────────

// generateNvim gera as configurações de projeto específicas para Neovim.
//
// Cria um arquivo .nvim.lua na raiz do projeto, carregado automaticamente
// por plugins de configuração local (ex: nvim-local-lua, exrc.nvim, neoconf.nvim).
//
// O arquivo configura:
//   - Integração com o servidor LSP Clangd via nvim-lspconfig
//   - Keymaps de projeto (build, test, run)
//   - Integração com nvim-dap para debug (DAP — Debug Adapter Protocol)
//   - Configurações do clangd específicas para este projeto
//
// Pré-requisitos no Neovim do usuário:
//   - nvim-lspconfig   (LSP client)
//   - nvim-dap         (debug adapter, opcional)
//   - nvim-dap-ui      (interface de debug, opcional)
//   - telescope.nvim   (fuzzy finder, opcional)
//
// Referência: https://github.com/neovim/nvim-lspconfig/blob/master/doc/configs.md#clangd
func generateNvim(root string, data *Data, verbose bool) error {
	nvimLuaPath := filepath.Join(root, ".nvim.lua")

	if err := writeIDETemplate(nvimLuaPath, "nvim_lua", tmplNvimLua, data, verbose); err != nil {
		return fmt.Errorf("gerar .nvim.lua: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Templates CLion
// ─────────────────────────────────────────────────────────────────────────────

// tmplCLionIdeaGitignore é o template para .idea/.gitignore.
//
// O diretório .idea/ contém tanto arquivos gerados (que não devem ser versionados)
// quanto arquivos de configuração do projeto (que devem ser versionados).
// Este .gitignore granular versiona apenas os arquivos relevantes para a equipe.
const tmplCLionIdeaGitignore = `# =============================================================================
# .idea/.gitignore — Controle do que versionar dentro do .idea/
# =============================================================================
# CLion armazena configurações de projeto e de usuário em .idea/.
# Esta regra garante que apenas configurações de PROJETO sejam versionadas,
# enquanto configurações PESSOAIS e arquivos GERADOS são ignorados.
# =============================================================================

# ── Arquivos gerados — NUNCA versionar ────────────────────────────────────────

# Cache e índices do CLion (gerados automaticamente, muito grandes)
/workspace.xml
/tasks.xml
/usage.statistics.xml
/shelf/

# Configurações pessoais do usuário (variam por desenvolvedor)
/dataSources/
/dataSources.local.xml
/dynamic.xml
/httpRequests/
/uiDesigner.xml
/.name

# Histórico local (specific to the user's machine)
/dictionaries/
/localHistory/

# Arquivos de cache de compilação e análise
*.iws

# ── Arquivos de projeto — VERSIONAR ──────────────────────────────────────────
# Os arquivos abaixo devem ser versionados para compartilhar configurações
# de projeto com a equipe (run configurations, inspections, cmake profiles).
#
# Não ignore:
#   cmake.xml                  — perfis CMake do projeto
#   inspectionProfiles/        — perfil de inspeção do projeto
#   runConfigurations/         — configurações de run/debug compartilhadas
#   .gitignore                 — este arquivo
`

// tmplCLionCMakeXML é o template para .idea/cmake.xml.
//
// Configura os perfis CMake do CLion mapeados para os CMakePresets.json do projeto.
// Permite que todos os desenvolvedores da equipe usem os mesmos perfis de build
// ao abrir o projeto no CLion sem precisar configurar manualmente.
//
// Referência: https://www.jetbrains.com/help/clion/cmake-profile.html
const tmplCLionCMakeXML = `<?xml version="1.0" encoding="UTF-8"?>
<!--
  .idea/cmake.xml — Perfis CMake para CLion
  ============================================================================
  Define os perfis de build do CLion mapeados para os presets do CMakePresets.json.
  Estes perfis são compartilhados com toda a equipe via controle de versão.

  Compatibilidade: CLion 2022.1+ com suporte a CMakePresets.json.
  Referência: https://www.jetbrains.com/help/clion/cmake-presets.html
  ============================================================================
-->
<project version="4">
  <component name="CMakeSharedSettings">
    <configurations>

      <!--
        Perfil Debug — mapeado para o preset 'debug' ou 'vcpkg-debug'.
        Usado para desenvolvimento com símbolos de debug completos.
      -->
      <configuration
        PROFILE_NAME="Debug"
        CONFIG_NAME="Debug"
        ENABLED="true"
        GENERATION_OPTIONS="--preset {{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}"
        GENERATION_DIR="$PROJECT_DIR$/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}"
        BUILD_OPTIONS="--preset {{if .UseVCPKG}}build-vcpkg-debug{{else}}build-debug{{end}}"
      />

      <!--
        Perfil Release — mapeado para o preset 'release' ou 'vcpkg-release'.
        Usado para builds otimizados de produção.
      -->
      <configuration
        PROFILE_NAME="Release"
        CONFIG_NAME="Release"
        ENABLED="true"
        GENERATION_OPTIONS="--preset {{if .UseVCPKG}}vcpkg-release{{else}}release{{end}}"
        GENERATION_DIR="$PROJECT_DIR$/build/{{if .UseVCPKG}}vcpkg-release{{else}}release{{end}}"
        BUILD_OPTIONS="--preset {{if .UseVCPKG}}build-vcpkg-release{{else}}build-release{{end}}"
      />

      <!--
        Perfil RelWithDebInfo — mapeado para o preset 'release-with-debug'.
        Útil para profiling com símbolos de debug em código otimizado.
      -->
      <configuration
        PROFILE_NAME="RelWithDebInfo"
        CONFIG_NAME="RelWithDebInfo"
        ENABLED="true"
        GENERATION_OPTIONS="--preset release-with-debug"
        GENERATION_DIR="$PROJECT_DIR$/build/release-with-debug"
        BUILD_OPTIONS="--preset build-release"
      />

      <!--
        Perfil Sanitizers — mapeado para o preset 'sanitize'.
        Habilita AddressSanitizer e UBSanitizer para detecção de bugs em runtime.
      -->
      <configuration
        PROFILE_NAME="Sanitizers"
        CONFIG_NAME="Debug"
        ENABLED="true"
        GENERATION_OPTIONS="--preset sanitize"
        GENERATION_DIR="$PROJECT_DIR$/build/sanitize"
        BUILD_OPTIONS="--preset build-sanitize"
      />

    </configurations>
  </component>
</project>
`

// tmplCLionInspectionProfile é o template para .idea/inspectionProfiles/Project_Default.xml.
//
// Define o perfil de inspeção de código padrão do projeto no CLion.
// Habilita inspeções relevantes para C++ moderno e desabilita as ruidosas.
//
// Referência: https://www.jetbrains.com/help/clion/code-inspection.html
const tmplCLionInspectionProfile = `<?xml version="1.0" encoding="UTF-8"?>
<!--
  .idea/inspectionProfiles/Project_Default.xml
  ============================================================================
  Perfil de inspeção de código padrão para {{ .ProjectName }}.
  Define quais análises estáticas o CLion executará no projeto.

  Para editar via UI: Settings → Editor → Inspections
  ============================================================================
-->
<component name="InspectionProjectProfileManager">
  <profile version="1.0">
    <option name="myName" value="Project Default" />

    <!-- ── C++ — Boas Práticas ──────────────────────────────────────────────── -->

    <!-- Detecta uso de funções C obsoletas em C++ (printf, malloc, etc.) -->
    <inspection_tool class="CppDeprecatedAPIInspection" enabled="true" level="WARNING" />

    <!-- Sugere uso de nullptr em vez de NULL ou 0 para ponteiros -->
    <inspection_tool class="CppNullptrInspection" enabled="true" level="WARNING" />

    <!-- Detecta variáveis declaradas mas não utilizadas -->
    <inspection_tool class="CppUnusedIncludeDirective" enabled="true" level="WARNING" />

    <!-- Alerta sobre conversões implícitas potencialmente perigosas -->
    <inspection_tool class="CppImplicitConversionInspection" enabled="true" level="WARNING" />

    <!-- ── C++ — Segurança ─────────────────────────────────────────────────── -->

    <!-- Detecta possível acesso a ponteiro nulo -->
    <inspection_tool class="CppNullDereferenceInspection" enabled="true" level="ERROR" />

    <!-- Detecta vazamentos de memória (new sem delete correspondente) -->
    <inspection_tool class="CppMemoryLeakInspection" enabled="true" level="WARNING" />

    <!-- Alerta sobre comparação de signed e unsigned -->
    <inspection_tool class="CppSignedUnsignedComparison" enabled="true" level="WARNING" />

    <!-- ── C++ — Modernização ─────────────────────────────────────────────── -->

    <!-- Sugere uso de range-based for em vez de iteradores explícitos -->
    <inspection_tool class="CppRangeBasedForInspection" enabled="true" level="WEAK WARNING" />

    <!-- Sugere uso de auto para simplificar declarações de tipo complexas -->
    <inspection_tool class="CppAutoInspection" enabled="true" level="INFORMATION" />

    <!-- Detecta copy constructors que deveriam ser move constructors -->
    <inspection_tool class="CppRedundantCastInspection" enabled="true" level="WEAK WARNING" />

    <!-- ── Desabilitados — muito ruidosos para uso diário ─────────────────── -->

    <!-- Inspeção de estilo muito opinativa — use o clang-format no lugar -->
    <inspection_tool class="CppInconsistentNamingInspection" enabled="false" level="WEAK WARNING" />

  </profile>
</component>
`

// tmplCLionRunDebug é o template para .idea/runConfigurations/Debug.xml.
//
// Define uma configuração de Run/Debug compartilhada para o projeto.
// Permite que todos os desenvolvedores da equipe usem a mesma configuração
// de debug ao clicar em "Debug" no CLion sem precisar configurar manualmente.
//
// Referência: https://www.jetbrains.com/help/clion/run-debug-configuration.html
const tmplCLionRunDebug = `<component name="ProjectRunConfigurationManager">
  <!--
    .idea/runConfigurations/Debug.xml
    ============================================================================
    Configuração de Run/Debug compartilhada para {{ .ProjectName }}.
    Mapeada para o perfil CMake "Debug" definido em cmake.xml.

    Para adicionar mais configurações: Run → Edit Configurations...
    ============================================================================
  -->
  <configuration
    default="false"
    name="{{.ProjectName}} [Debug]"
    type="CMakeRunConfiguration"
    factoryName="Application"
    REDIRECT_INPUT="false"
    ELEVATE="false"
    USE_EXTERNAL_CONSOLE="false"
    PASS_PARENT_ENVS_2="true"
    PROJECT_NAME="{{.ProjectName}}"
    TARGET_NAME="{{.ProjectName}}"
    CONFIG_NAME="Debug"
    RUN_TARGET_PROJECT_NAME="{{.ProjectName}}"
    RUN_TARGET_NAME="{{.ProjectName}}"
  >
    <method v="2">
      <!-- Compila automaticamente antes de depurar -->
      <option name="com.jetbrains.cidr.execution.CidrBuildBeforeRunTaskProvider$BuildBeforeRunTask" enabled="true" />
    </method>
  </configuration>
</component>
`

// ─────────────────────────────────────────────────────────────────────────────
// Template Neovim
// ─────────────────────────────────────────────────────────────────────────────

// tmplNvimLua é o template para o arquivo .nvim.lua na raiz do projeto.
//
// Este arquivo é carregado automaticamente por plugins de configuração local
// do Neovim. Ele é específico do projeto e complementa o init.lua global do usuário.
//
// Plugins que carregam .nvim.lua automaticamente:
//   - klen/nvim-config-local       (mais popular)
//   - MunifTanjim/exrc.nvim        (baseado em 'exrc' do Vim)
//   - neoconf.nvim (Folke)         (via .neoconf.json + configuração Lua)
//
// Para carregar manualmente sem plugin:
//
//	-- No init.lua do usuário:
//	vim.opt.exrc = true   -- habilita .nvim.lua automático (risco de segurança!)
//
// IMPORTANTE: .nvim.lua pode executar código arbitrário. Adicione ao .gitignore
// caso contenha configurações sensíveis ou caminhos de máquina específicos.
// Este arquivo gerado é seguro para versionar pois contém apenas configurações
// de projeto sem informações sensíveis.
//
// Referência: https://github.com/neovim/nvim-lspconfig/blob/master/doc/configs.md#clangd
const tmplNvimLua = `-- =============================================================================
-- .nvim.lua — Configuração de projeto Neovim para {{ .ProjectName }}
-- =============================================================================
-- Este arquivo é carregado automaticamente por plugins de exrc/local-config.
-- Ele COMPLEMENTA (não substitui) o init.lua global do usuário.
--
-- Plugins suportados para carregamento automático:
--   - klen/nvim-config-local    → adicione ao lazy.nvim/packer
--   - MunifTanjim/exrc.nvim     → alternativa segura ao exrc nativo
--
-- Para carregamento manual em sessão única:
--   :luafile .nvim.lua
-- =============================================================================

-- Proteção: só executa se o Neovim for suficientemente recente (0.9+).
if vim.fn.has("nvim-0.9") == 0 then
  vim.notify(
    "[{{ .ProjectName }}] .nvim.lua requer Neovim 0.9+. Configuração ignorada.",
    vim.log.levels.WARN
  )
  return
end

-- =============================================================================
-- Variáveis de projeto
-- =============================================================================

-- Diretório raiz do projeto (onde este .nvim.lua está localizado).
local project_root = vim.fn.fnamemodify(
  vim.fn.resolve(vim.fn.expand("<sfile>:p")), ":h"
)

-- Preset CMake padrão para este projeto.
-- Altere para "vcpkg-debug" se usar VCPKG.
local cmake_preset_debug   = "{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}"
local cmake_preset_release = "{{if .UseVCPKG}}vcpkg-release{{else}}release{{end}}"
local cmake_build_debug    = "{{if .UseVCPKG}}build-vcpkg-debug{{else}}build-debug{{end}}"

-- Caminho para o compile_commands.json gerado pelo CMake.
-- Usado pelo Clangd para análise precisa do código.
local compile_commands_dir = project_root
  .. "/build/"
  .. cmake_preset_debug

-- =============================================================================
-- Configuração do Clangd LSP
-- =============================================================================
-- Configura o servidor Clangd com as opções específicas deste projeto.
-- Requer o plugin nvim-lspconfig: https://github.com/neovim/nvim-lspconfig
-- =============================================================================

local ok_lspconfig, lspconfig = pcall(require, "lspconfig")
if ok_lspconfig then
  lspconfig.clangd.setup({
    -- Argumentos passados ao servidor Clangd ao iniciar.
    cmd = {
      "clangd",

      -- Aponta para o compile_commands.json do preset de debug.
      "--compile-commands-dir=" .. compile_commands_dir,

      -- Workers para indexação em background (ajuste conforme seu hardware).
      "--j=4",

      -- Inclui símbolos de todos os escopos no autocompletar.
      "--all-scopes-completion",

      -- Sugere headers automaticamente ao completar símbolos desconhecidos.
      "--header-insertion=iwyu",

      -- Habilita diagnósticos do clang-tidy inline.
      "--clang-tidy",

      -- Nível de log: error | warning | info | verbose
      "--log=error",

      -- Habilita inlay hints (tipos de variáveis auto, nomes de parâmetros).
      "--inlay-hints",

      -- Usa o cache em disco para acelerar reinicializações do LSP.
      "--enable-config",
    },

    -- Raiz do projeto: onde o Clangd procura o .clangd e compile_commands.json.
    root_dir = lspconfig.util.root_pattern(
      "CMakeLists.txt",
      "CMakePresets.json",
      ".clangd",
      ".git"
    ),

    -- Capacidades do cliente LSP (autocompletar, snippets, etc.).
    -- Integra com nvim-cmp se disponível.
    capabilities = (function()
      local ok_cmp, cmp_lsp = pcall(require, "cmp_nvim_lsp")
      if ok_cmp then
        return cmp_lsp.default_capabilities()
      end
      return vim.lsp.protocol.make_client_capabilities()
    end)(),

    -- Configurações de inicialização específicas do Clangd.
    init_options = {
      usePlaceholders    = true,   -- Insere placeholders em completions de função
      completeUnimported = true,   -- Completa símbolos de headers não importados
      clangdFileStatus   = true,   -- Exibe status de indexação na statusline
    },

    -- Keymaps e configurações aplicadas quando o LSP se conecta ao buffer.
    on_attach = function(client, bufnr)
      -- Garante que este on_attach só configure buffers C++.
      local ft = vim.bo[bufnr].filetype
      if ft ~= "cpp" and ft ~= "c" then
        return
      end

      -- Desabilita formatação do LSP (usa clang-format diretamente).
      client.server_capabilities.documentFormattingProvider = false

      -- ── Keymaps LSP específicos do projeto ────────────────────────────────

      local opts = { buffer = bufnr, silent = true }

      -- Navegação de código
      vim.keymap.set("n", "gd", vim.lsp.buf.definition,       vim.tbl_extend("force", opts, { desc = "LSP: Go to definition" }))
      vim.keymap.set("n", "gD", vim.lsp.buf.declaration,      vim.tbl_extend("force", opts, { desc = "LSP: Go to declaration" }))
      vim.keymap.set("n", "gi", vim.lsp.buf.implementation,   vim.tbl_extend("force", opts, { desc = "LSP: Go to implementation" }))
      vim.keymap.set("n", "gr", vim.lsp.buf.references,       vim.tbl_extend("force", opts, { desc = "LSP: Find references" }))
      vim.keymap.set("n", "gt", vim.lsp.buf.type_definition,  vim.tbl_extend("force", opts, { desc = "LSP: Go to type definition" }))

      -- Documentação e assinaturas
      vim.keymap.set("n", "K",     vim.lsp.buf.hover,            vim.tbl_extend("force", opts, { desc = "LSP: Hover documentation" }))
      vim.keymap.set("n", "<C-k>", vim.lsp.buf.signature_help,   vim.tbl_extend("force", opts, { desc = "LSP: Signature help" }))

      -- Ações de código
      vim.keymap.set("n", "<leader>rn", vim.lsp.buf.rename,         vim.tbl_extend("force", opts, { desc = "LSP: Rename symbol" }))
      vim.keymap.set("n", "<leader>ca", vim.lsp.buf.code_action,    vim.tbl_extend("force", opts, { desc = "LSP: Code action" }))
      vim.keymap.set("n", "<leader>f",  function()                  -- Formata com clang-format
        vim.lsp.buf.format({ async = true })
      end, vim.tbl_extend("force", opts, { desc = "LSP: Format file" }))

      -- Clangd: alterna entre header (.hpp) e implementação (.cpp)
      vim.keymap.set("n", "<leader>h", "<cmd>ClangdSwitchSourceHeader<CR>",
        vim.tbl_extend("force", opts, { desc = "Clangd: Switch header/source" }))

      -- Diagnósticos
      vim.keymap.set("n", "[d", vim.diagnostic.goto_prev, vim.tbl_extend("force", opts, { desc = "Diagnóstico anterior" }))
      vim.keymap.set("n", "]d", vim.diagnostic.goto_next, vim.tbl_extend("force", opts, { desc = "Próximo diagnóstico" }))
      vim.keymap.set("n", "<leader>d", vim.diagnostic.open_float, vim.tbl_extend("force", opts, { desc = "Ver diagnóstico" }))

      -- Inlay hints (Neovim 0.10+)
      if vim.lsp.inlay_hint then
        vim.keymap.set("n", "<leader>ih", function()
          vim.lsp.inlay_hint.enable(not vim.lsp.inlay_hint.is_enabled({ bufnr = bufnr }), { bufnr = bufnr })
        end, vim.tbl_extend("force", opts, { desc = "Toggle inlay hints" }))

        -- Habilita inlay hints por padrão neste projeto.
        vim.lsp.inlay_hint.enable(true, { bufnr = bufnr })
      end

      vim.notify(
        string.format("[{{ .ProjectName }}] Clangd conectado ao buffer %d", bufnr),
        vim.log.levels.DEBUG
      )
    end,
  })

  vim.notify("[{{ .ProjectName }}] Clangd LSP configurado.", vim.log.levels.INFO)
else
  vim.notify(
    "[{{ .ProjectName }}] nvim-lspconfig não encontrado. Instale: https://github.com/neovim/nvim-lspconfig",
    vim.log.levels.WARN
  )
end

-- =============================================================================
-- Keymaps de projeto (CMake Build / Test / Run)
-- =============================================================================
-- Keymaps específicos deste projeto para compilar, testar e executar.
-- Mapeados com <localleader> para não conflitar com os keymaps globais.
-- =============================================================================

-- Usa <localleader>b como prefixo de build (\b por padrão, configure o seu).
local function cmake_term(cmd)
  -- Abre um terminal flutuante ou split com o comando CMake.
  vim.cmd("botright 15split")
  vim.fn.termopen(cmd, { cwd = project_root })
  vim.cmd("startinsert")
end

-- <localleader>bc — CMake Configure (Debug)
vim.keymap.set("n", "<localleader>bc", function()
  cmake_term("cmake --preset " .. cmake_preset_debug)
end, { desc = "[{{ .ProjectName }}] CMake: Configure (Debug)" })

-- <localleader>bb — CMake Build (Debug)
vim.keymap.set("n", "<localleader>bb", function()
  cmake_term("cmake --build --preset " .. cmake_build_debug)
end, { desc = "[{{ .ProjectName }}] CMake: Build (Debug)" })

-- <localleader>br — CMake Build (Release)
vim.keymap.set("n", "<localleader>br", function()
  cmake_term(
    "cmake --preset " .. cmake_preset_release
    .. " && cmake --build --preset build-"
    .. ({{if .UseVCPKG}}"vcpkg-"{{else}}""{{end}} .. "release")
  )
end, { desc = "[{{ .ProjectName }}] CMake: Build (Release)" })

-- <localleader>bt — CMake Run Tests (CTest)
vim.keymap.set("n", "<localleader>bt", function()
  cmake_term(
    "cmake --build --preset " .. cmake_build_debug
    .. " && ctest --preset {{if .UseVCPKG}}test-vcpkg-debug{{else}}test-debug{{end}} --output-on-failure"
  )
end, { desc = "[{{ .ProjectName }}] CMake: Run Tests" })

{{- if .IsExecutable}}
-- <localleader>bx — Executar o binário compilado (Debug)
vim.keymap.set("n", "<localleader>bx", function()
  cmake_term(
    "cmake --build --preset " .. cmake_build_debug
    .. " && ./build/" .. cmake_preset_debug .. "/bin/{{ .ProjectName }}"
  )
end, { desc = "[{{ .ProjectName }}] CMake: Build & Run" })
{{- end}}

-- <localleader>bk — Apagar diretório build/
vim.keymap.set("n", "<localleader>bk", function()
  local confirm = vim.fn.input("Apagar build/? [s/N] ")
  if confirm:lower() == "s" or confirm:lower() == "sim" then
    vim.fn.delete(project_root .. "/build", "rf")
    vim.notify("[{{ .ProjectName }}] Diretório build/ apagado.", vim.log.levels.INFO)
  end
end, { desc = "[{{ .ProjectName }}] CMake: Delete build/" })

-- =============================================================================
-- Configuração do nvim-dap (Debug Adapter Protocol)
-- =============================================================================
-- Configura o nvim-dap para depurar com LLDB ou GDB.
-- Requer: mfussenegger/nvim-dap + nvim-dap-ui (opcional)
-- Documentação: https://github.com/mfussenegger/nvim-dap
-- =============================================================================

local ok_dap, dap = pcall(require, "dap")
if ok_dap then

  -- ── Adapter LLDB ────────────────────────────────────────────────────────────
  -- Tenta configurar o adapter LLDB (recomendado para Linux/macOS com Clang).
  -- Se lldb-vscode / lldb-dap não estiver instalado, tenta o GDB como fallback.
  local lldb_executable = vim.fn.exepath("lldb-dap")
    or vim.fn.exepath("lldb-vscode")

  if lldb_executable and lldb_executable ~= "" then
    dap.adapters.lldb = {
      type    = "executable",
      command = lldb_executable,
      name    = "lldb",
    }

    -- Configuração de debug para C++ com LLDB
    dap.configurations.cpp = dap.configurations.cpp or {}
    table.insert(dap.configurations.cpp, {
      name    = "[{{ .ProjectName }}] Debug (LLDB)",
      type    = "lldb",
      request = "launch",
      program = project_root
        .. "/build/" .. cmake_preset_debug
        .. "/bin/{{ .ProjectName }}",
      cwd            = project_root,
      stopOnEntry    = false,
      args           = {},
      runInTerminal  = false,
      -- Inicializa pretty-printers do LLDB para exibir STL legível.
      initCommands   = {
        "settings set target.max-string-summary-length 256",
      },
      -- preLaunchTask: compile antes de depurar (requer nvim-dap-tasks ou similar).
    })

    {{- if not .IsExecutable}}
    -- Configuração para depurar os testes da biblioteca
    table.insert(dap.configurations.cpp, {
      name    = "[{{ .ProjectName }}] Debug Tests (LLDB)",
      type    = "lldb",
      request = "launch",
      program = project_root
        .. "/build/" .. cmake_preset_debug
        .. "/tests/{{ .ProjectName }}_tests",
      cwd           = project_root,
      stopOnEntry   = false,
      args          = {},
      runInTerminal = false,
    })
    {{- end}}

  else
    -- ── Fallback: GDB ─────────────────────────────────────────────────────────
    local gdb_executable = vim.fn.exepath("gdb")
    if gdb_executable and gdb_executable ~= "" then
      dap.adapters.gdb = {
        type    = "executable",
        command = gdb_executable,
        args    = { "--interpreter=dap", "--eval-command", "set print pretty on" },
        name    = "gdb",
      }

      dap.configurations.cpp = dap.configurations.cpp or {}
      table.insert(dap.configurations.cpp, {
        name    = "[{{ .ProjectName }}] Debug (GDB)",
        type    = "gdb",
        request = "launch",
        program = project_root
          .. "/build/" .. cmake_preset_debug
          .. "/bin/{{ .ProjectName }}",
        cwd         = project_root,
        stopAtBeginningOfMainSubprogram = false,
      })
    end
  end

  -- ── Keymaps de debug ────────────────────────────────────────────────────────

  vim.keymap.set("n", "<F5>",       dap.continue,          { desc = "DAP: Continue / Start" })
  vim.keymap.set("n", "<F10>",      dap.step_over,         { desc = "DAP: Step Over" })
  vim.keymap.set("n", "<F11>",      dap.step_into,         { desc = "DAP: Step Into" })
  vim.keymap.set("n", "<F12>",      dap.step_out,          { desc = "DAP: Step Out" })
  vim.keymap.set("n", "<leader>db", dap.toggle_breakpoint, { desc = "DAP: Toggle Breakpoint" })
  vim.keymap.set("n", "<leader>dr", dap.repl.open,         { desc = "DAP: Open REPL" })
  vim.keymap.set("n", "<leader>dl", dap.run_last,          { desc = "DAP: Run Last" })

  -- Breakpoint condicional
  vim.keymap.set("n", "<leader>dB", function()
    dap.set_breakpoint(vim.fn.input("Condição do breakpoint: "))
  end, { desc = "DAP: Conditional Breakpoint" })

  -- ── nvim-dap-ui ─────────────────────────────────────────────────────────────
  local ok_dapui, dapui = pcall(require, "dapui")
  if ok_dapui then
    -- Abre/fecha a UI automaticamente ao iniciar/encerrar uma sessão de debug.
    dap.listeners.after.event_initialized["dapui_config"]  = dapui.open
    dap.listeners.before.event_terminated["dapui_config"]  = dapui.close
    dap.listeners.before.event_exited["dapui_config"]      = dapui.close

    vim.keymap.set("n", "<leader>du", dapui.toggle, { desc = "DAP: Toggle UI" })
  end

  vim.notify("[{{ .ProjectName }}] nvim-dap configurado.", vim.log.levels.INFO)
else
  vim.notify(
    "[{{ .ProjectName }}] nvim-dap não encontrado — debug indisponível.\n"
    .. "Instale: https://github.com/mfussenegger/nvim-dap",
    vim.log.levels.DEBUG
  )
end

-- =============================================================================
-- Formatação com clang-format
-- =============================================================================

-- Formata o buffer atual com clang-format ao salvar (somente em buffers C++).
local fmt_group = vim.api.nvim_create_augroup("{{ .ProjectName }}ClangFormat", { clear = true })
vim.api.nvim_create_autocmd("BufWritePre", {
  group   = fmt_group,
  pattern = { "*.cpp", "*.hpp", "*.cxx", "*.hxx", "*.cc", "*.h" },
  callback = function()
    -- Verifica se clang-format está disponível antes de tentar formatar.
    if vim.fn.exepath("clang-format") ~= "" then
      vim.lsp.buf.format({ async = false, timeout_ms = 2000 })
    end
  end,
  desc = "[{{ .ProjectName }}] Auto-format C++ ao salvar",
})

-- =============================================================================
-- Configurações de editor para este projeto
-- =============================================================================

-- Cria autocomando para aplicar configurações específicas de C++ apenas neste projeto.
local editor_group = vim.api.nvim_create_augroup("{{ .ProjectName }}Editor", { clear = true })
vim.api.nvim_create_autocmd({ "BufEnter", "BufWinEnter" }, {
  group   = editor_group,
  pattern = { "*.cpp", "*.hpp", "*.h", "*.cxx", "*.cc" },
  callback = function(ev)
    -- Só aplica em buffers dentro deste projeto.
    local buf_path = vim.api.nvim_buf_get_name(ev.buf)
    if not buf_path:find(project_root, 1, true) then
      return
    end

    local opt = vim.bo[ev.buf]
    opt.tabstop     = 4       -- Tab visual = 4 espaços
    opt.shiftwidth  = 4       -- Indentação = 4 espaços
    opt.expandtab   = true    -- Usa espaços em vez de tabs
    opt.textwidth   = 100     -- Largura de texto = 100 colunas (igual ao .clang-format)
    opt.colorcolumn = nil     -- Não sobrescreve colorcolumn global

    -- Habilita quebra de linha visual (sem modificar o arquivo).
    vim.wo.wrap      = false
    vim.wo.linebreak = true
  end,
  desc = "[{{ .ProjectName }}] Configurações de editor para C++",
})

-- =============================================================================
-- Mensagem de boas-vindas
-- =============================================================================

vim.schedule(function()
  vim.notify(
    "📦 {{ .ProjectName }} — Workspace Neovim carregado\n"
    .. "  Compilar:  <localleader>bb\n"
    .. "  Testar:    <localleader>bt\n"
    .. "  Debug:     F5\n"
    .. "  LSP:       gd / K / <leader>ca",
    vim.log.levels.INFO
  )
end)
`
