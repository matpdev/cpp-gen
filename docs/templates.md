# Custom Templates

`cpp-gen` can bootstrap a project from **any Git repository or local folder**, not just the
built-in `blank` and `vulkan` templates — in the spirit of `npm create <template>` or
`cargo generate`. This guide shows how to author, test, and publish your own templates, and
how to make them discoverable through `cpp-gen templates`.

---

## Table of Contents

- [Concepts](#concepts)
- [Starting from a scaffold](#starting-from-a-scaffold)
- [Quick demo](#quick-demo)
- [Authoring reference](#authoring-reference)
- [Testing a template locally](#testing-a-template-locally)
- [Making a template discoverable](#making-a-template-discoverable)
- [Worked example: the raylib-app template](#worked-example-the-raylib-app-template)
- [Cheatsheet](#cheatsheet)

---

## Concepts

There are three built-in template kinds, selected via `--template` (flag) or the **Template do
projeto** step of the interactive form:

| `--template` value        | What happens                                                              |
|----------------------------|-----------------------------------------------------------------------------|
| `blank` *(default)*        | cpp-gen's own procedural generator (layout, CMake, packages, IDE configs) |
| `vulkan`                   | Built-in Vulkan template, embedded in the binary                          |
| anything else               | Treated as a **custom template** — see below                              |

Any other `--template` value is resolved, in order, as:

1. **A local folder** — `./my-template`, `~/templates/foo`, `/abs/path`, or any string that
   already exists as a directory.
2. **A GitHub shorthand** — `owner/repo`, `owner/repo/subdir`, `owner/repo#branch`,
   `owner/repo/subdir#tag` (expands to `https://github.com/owner/repo.git`).
3. **A full Git URL** — `https://github.com/owner/repo.git[#ref]` or
   `git@host:owner/repo.git[#ref]`.
4. **A registry alias** — a short name like `raylib-app` looked up via `cpp-gen templates`
   (see [Making a template discoverable](#making-a-template-discoverable)). Only tried when
   the value doesn't already match one of the three forms above.

Whatever the source, cpp-gen:

1. Fetches it (clones the repo into a temp dir, or reads the local folder directly — nothing
   is downloaded for local paths).
2. Walks every file, substituting Go `text/template` placeholders (`{{.Name}}`, `{{.NameSnake}}`,
   …) in **both file contents and file/directory names**.
3. Writes the result into the new project directory.
4. Still runs the shared post-steps: IDE config, `.clangd`/`.clang-format`, Git init — the
   template only replaces the "project skeleton" step (layout, CMake, packages), same as the
   built-in `blank` flow.

---

## Starting from a scaffold

The fastest way to start a real template — not just a demo — is `cpp-gen templates create`,
which generates a minimal, working starter using every convention below (variable
substitution, filename placeholders, a conditional file) instead of writing them by hand:

```bash
cpp-gen templates create ./my-template
```

This produces:

```
my-template/
├── TEMPLATE_GUIDE.md                           ← authoring reference; delete once you're done
├── README.md                                    ← real template file, ships with generated projects
├── CMakeLists.txt                                ← {{.Name}}, {{.Standard}}, {{if .UseVCPKG}}...
├── src/
│   └── {{.NameSnake}}.cpp
└── {{if .UseVCPKG}}vcpkg.json{{end}}
```

It refuses to run against a non-empty directory, so it's always safe to point it at a fresh
folder. Try it immediately, no editing required:

```bash
cpp-gen new demo --template ./my-template --no-interactive --output /tmp/scratch
cmake -B /tmp/scratch/demo/build -S /tmp/scratch/demo && cmake --build /tmp/scratch/demo/build
```

Add `--register <alias>` to also add it to your [local catalog](#making-a-template-discoverable)
in the same step, so it's immediately usable as `--template <alias>` and shows up in `cpp-gen
templates`:

```bash
cpp-gen templates create ./my-template --register my-starter --description "My starter template"
cpp-gen new demo --template my-starter
```

From here, edit the generated files freely — add dependencies, more source files, extra
conditional files for other package-manager branches, etc. Everything in
[Authoring reference](#authoring-reference) below applies to what you just generated.

---

## Quick demo

If you'd rather see the mechanics by hand instead of using the scaffold above, point
`--template` straight at a folder you write yourself — no Git repo, no publishing, nothing to
set up.

```bash
mkdir -p /tmp/hello-template/src

cat > /tmp/hello-template/CMakeLists.txt <<'EOF'
cmake_minimum_required(VERSION 3.20)
project({{.Name}} VERSION {{.Version}} LANGUAGES CXX)
set(CMAKE_CXX_STANDARD {{.Standard}})
add_executable({{.Name}} src/main.cpp)
EOF

cat > "/tmp/hello-template/src/{{.NameSnake}}.cpp" <<'EOF'
#include <cstdio>

int main() {
    std::printf("Hello from {{.NamePascal}}!\n");
    return 0;
}
EOF

cpp-gen new demo --template /tmp/hello-template --no-interactive --std 20
```

Result: `demo/CMakeLists.txt` has `project(demo VERSION 1.0.0 ...)`, and
`demo/src/demo.cpp` (the filename itself was substituted) prints `Hello from Demo!`.

Nothing about this required a manifest file, a build step, or a special repo layout — the
folder *is* the template.

---

## Authoring reference

### Variable substitution

Every file (and every path segment) is run through Go's `text/template`, with the following
fields available — this is the exact `TemplateData` struct cpp-gen uses internally
(`internal/generator/generator.go`), so anything the built-in `blank` template can use, your
template can use too.

**Metadata**

| Field                | Example           | Notes                                    |
|-----------------------|--------------------|-------------------------------------------|
| `{{.Name}}`           | `my-project`       | Exactly as typed                          |
| `{{.NameUpper}}`      | `MY_PROJECT`       | For CMake vars / include guards           |
| `{{.NameSnake}}`      | `my_project`       | For file names / C++ identifiers          |
| `{{.NamePascal}}`     | `MyProject`        | For C++ class/namespace names             |
| `{{.Description}}`    | `A CLI tool`        | May be empty — guard with `{{if .Description}}` |
| `{{.Author}}`         | `Jane Doe`          | May be empty                              |
| `{{.Version}}`        | `1.0.0`             | SemVer                                    |
| `{{.Year}}`           | `2026`              | Current year                              |
| `{{.Standard}}`       | `20`                | C++ standard, numeric string              |

**Package manager** (set via `--pkg` / the form, independent of which template is used)

| Field                    | True when             |
|---------------------------|------------------------|
| `{{.UseVCPKG}}`           | `--pkg vcpkg`          |
| `{{.UseFetchContent}}`    | `--pkg fetchcontent`   |

**Project type**

| Field                | True when                |
|------------------------|----------------------------|
| `{{.IsExecutable}}`    | `--type executable` (default) |
| `{{.IsStaticLib}}`     | `--type static-lib`        |
| `{{.IsHeaderOnly}}`    | `--type header-only`       |

**IDE**

| Field           | True when       |
|------------------|------------------|
| `{{.IsVSCode}}`  | `--ide vscode`   |
| `{{.IsCLion}}`   | `--ide clion`    |
| `{{.IsNvim}}`    | `--ide nvim`     |
| `{{.IsZed}}`     | `--ide zed`      |

**Tooling**

| Field                    | Meaning                                  |
|---------------------------|--------------------------------------------|
| `{{.UseGit}}`             | Initialize a Git repo after generation      |
| `{{.UseClangd}}`          | `--no-clangd` not passed                    |
| `{{.UseClangFormat}}`     | `--no-clang-format` not passed              |
| `{{.ClangFormatStyle}}`   | e.g. `LLVM`, `Google`                        |

> `Layout*` fields also exist (folder layout metadata) but only make sense for cpp-gen's own
> procedural generator — custom templates own their directory layout outright, so you generally
> won't need them.

### Conditionally including a whole file or folder

A file or directory whose **rendered name is empty** is skipped entirely. This is how you make
a file appear only for certain choices — name it with an `{{if}}` that produces nothing when
false:

```
vcpkg-templates/
├── {{if .UseVCPKG}}vcpkg.json{{end}}
├── {{if .UseVCPKG}}vcpkg-configuration.json{{end}}
└── CMakeLists.txt
```

Generate with `--pkg vcpkg` → both files are created. Generate with `--pkg fetchcontent` or
`--pkg none` → both are silently skipped, no empty/broken files left behind. The same trick
works on directories (name the folder `{{if .X}}dir{{end}}`; if `X` is false the whole
subtree is skipped).

### Conditional content

Ordinary `{{if}} / {{else if}} / {{else}} / {{end}}` works anywhere inside a file:

```cmake
{{if .UseVCPKG}}
find_package(raylib CONFIG REQUIRED)
{{else if .UseFetchContent}}
include(FetchContent)
FetchContent_Declare(raylib GIT_REPOSITORY https://github.com/raysan5/raylib.git GIT_TAG 5.5)
FetchContent_MakeAvailable(raylib)
{{else}}
find_package(raylib CONFIG REQUIRED)  # assume it's installed system-wide
{{end}}
```

### Binary files & "not actually a template" files

Every file is attempted as a template; if it fails, it's **copied byte-for-byte instead** —
generation never aborts because of this:

- Files that look binary (a NUL byte in the first 8&nbsp;KB) are copied verbatim, no
  substitution attempted.
- Text files that aren't valid Go template syntax are also copied verbatim. This matters for
  C++: aggregate initialization like `int arr[2]{{1, 2}};` contains a literal `{{`/`}}` pair
  that Go's template parser can't make sense of, so *that specific file* silently loses
  variable substitution and is copied as-is. If a file needs both brace-init and
  `{{.Something}}` substitution, rewrite the init to avoid the double-brace
  (`int arr[2] = {1, 2};` is fine — only `{{` immediately followed by `}}`-style syntax is a
  problem).

Run `cpp-gen new ... --verbose` to see exactly which files were substituted (`+`) vs. copied
as-is (`~`) or skipped by a conditional name (`~ ... omitido`).

### Your `README.md` / `.gitignore` win

If your template ships its own `README.md` or `.gitignore`, cpp-gen's default Git step leaves
them alone instead of overwriting them — only missing files get cpp-gen's generic versions.

---

## Testing a template locally

Point `--template` straight at the folder you're editing — no packaging step required:

```bash
cpp-gen new test-run --template ./my-template --no-interactive --output /tmp/scratch -v
```

Because there's no compile-time embedding involved (unlike the built-in `vulkan` template),
every edit is picked up on the next run — no rebuild of `cpp-gen` itself needed.

Once it works from a local folder, push it to a Git repo and confirm the shorthand resolves the
same way:

```bash
cpp-gen new test-run2 --template your-org/your-template --no-interactive --output /tmp/scratch
```

---

## Making a template discoverable

A template works via `--template <source>` the moment it exists — discovery (`cpp-gen
templates`, the "Procurar templates..." step in the form, and bare-name aliases like
`--template raylib-app`) is a separate, optional layer on top, backed by
`internal/registry`. It merges three sources, highest priority last:

1. **Embedded defaults** — bundled in the `cpp-gen` binary itself
   (`internal/registry/catalog.json`), curated by maintainers over releases. This is what makes
   `raylib-app` work out of the box on every install method (Homebrew, AUR, `go install`, …)
   without needing a local clone of this repo or the GitHub registry to be reachable.
2. **The GitHub registry** — a JSON file fetched over HTTP, shared by everyone.
3. **Your local catalog** — a JSON file you edit yourself, on your machine only.

All three use the identical shape:

```json
{
  "templates": [
    {
      "name": "my-cli-tool",
      "description": "CLI11 + fmt + spdlog starter.",
      "source": "your-org/my-cli-tool-template",
      "tags": ["cli"]
    }
  ]
}
```

- `name` — the alias used in `--template <name>` and matched by `cpp-gen templates <query>`
  (case-insensitive).
- `source` — anything from the [Concepts](#concepts) list: a local path, `owner/repo[/subdir][#ref]`,
  or a full Git URL.
- `tags` — free-form keywords, also matched by search.

### Add a template to your local catalog

The catalog file lives at `os.UserConfigDir()/cpp-gen/templates.json`, which resolves per OS:

| OS      | Path                                                  |
|---------|--------------------------------------------------------|
| macOS   | `~/Library/Application Support/cpp-gen/templates.json` |
| Linux   | `${XDG_CONFIG_HOME:-~/.config}/cpp-gen/templates.json` |
| Windows | `%AppData%\cpp-gen\templates.json`                     |

```bash
mkdir -p "$HOME/Library/Application Support/cpp-gen"   # macOS; adjust per the table above
cat > "$HOME/Library/Application Support/cpp-gen/templates.json" <<'EOF'
{
  "templates": [
    {
      "name": "my-cli-tool",
      "description": "CLI11 + fmt + spdlog starter.",
      "source": "/absolute/path/to/my-cli-tool-template",
      "tags": ["cli"]
    }
  ]
}
EOF

cpp-gen templates                     # confirm it shows up
cpp-gen new my-project --template my-cli-tool
```

A local entry with the same `name` as an embedded one takes priority over it — handy for
testing a fork of `raylib-app` locally without touching the binary:

```bash
cat > "$HOME/Library/Application Support/cpp-gen/templates.json" <<'EOF'
{
  "templates": [
    { "name": "raylib-app", "source": "/path/to/my/raylib-app-fork" }
  ]
}
EOF

cpp-gen new demo --template raylib-app   # now resolves to your fork, not the embedded one
```

### Publish to the shared GitHub registry

Host a `templates.json` (same shape) in a public repo, then point cpp-gen at its raw URL:

```bash
export CPPGEN_REGISTRY_URL=https://raw.githubusercontent.com/your-org/your-registry/main/templates.json
cpp-gen templates
```

`CPPGEN_REGISTRY_URL` also accepts a **local file path** instead of a URL — handy for
reviewing a `templates.json` before you actually publish it:

```bash
CPPGEN_REGISTRY_URL=./templates.json cpp-gen templates
```

Without the env var, cpp-gen uses the project's default community registry URL. That registry
is fetched with a 5-second timeout and fails soft — if it's unreachable, `cpp-gen templates`
and `cpp-gen new` just show fewer results, they never error out because of it.

---

## Worked example: the raylib-app template

`examples/templates/raylib-app/` in this repository is a complete, tested template — every
file in it doubles as a reference for the tricks above:

```
examples/templates/raylib-app/
├── CMakeLists.txt                              ← {{.Name}}, {{.Standard}}, {{if .UseVCPKG}}...
├── CMakePresets.json                           ← conditional vcpkg-* presets
├── README.md                                   ← per-package-manager build instructions
├── cmake/
│   └── Dependencies.cmake                      ← 3-way {{if .UseVCPKG}}/{{else if .UseFetchContent}}/{{else}}
├── src/
│   └── main.cpp                                ← {{.Name}} substituted into window title
├── {{if .UseVCPKG}}vcpkg.json{{end}}           ← whole-file conditional
└── {{if .UseVCPKG}}vcpkg-configuration.json{{end}}
```

Try all three package-manager variants:

```bash
cpp-gen new demo1 --template ./examples/templates/raylib-app --pkg vcpkg        --no-interactive
cpp-gen new demo2 --template ./examples/templates/raylib-app --pkg fetchcontent --no-interactive
cpp-gen new demo3 --template ./examples/templates/raylib-app --pkg none         --no-interactive
```

`demo1/` gets `vcpkg.json` + `vcpkg-configuration.json` and `vcpkg-*` CMake presets; `demo2/`
and `demo3/` don't — same template, same command, different generated files, purely from
`{{if .UseVCPKG}}`.

---

## Cheatsheet

```bash
# Scaffold a new starter template, optionally registering it locally
cpp-gen templates create ./my-template
cpp-gen templates create ./my-template --register my-starter --description "..."

# Use a local template while authoring it
cpp-gen new demo --template ./my-template --no-interactive

# Use a template from GitHub
cpp-gen new demo --template owner/repo
cpp-gen new demo --template owner/repo/subdir#v2.0.0

# Use a registered alias
cpp-gen new demo --template raylib-app

# List / search what's registered
cpp-gen templates
cpp-gen templates raylib

# Point at a registry you're testing before publishing
CPPGEN_REGISTRY_URL=./templates.json cpp-gen templates
```

```
{{.Name}} {{.NameSnake}} {{.NamePascal}} {{.NameUpper}}   # name forms
{{.Description}} {{.Author}} {{.Version}} {{.Year}}        # metadata
{{.Standard}}                                               # "17" | "20" | "23"
{{if .UseVCPKG}} … {{else if .UseFetchContent}} … {{else}} … {{end}}
{{if .UseVCPKG}}filename.json{{end}}                        # whole-file conditional
```
