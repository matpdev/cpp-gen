# Guia deste template (apague este arquivo quando terminar)

Este arquivo é só para você, autor do template — ele **não** é copiado verbatim como os
outros: se você rodar `cpp-gen new` apontando pra este template sem apagá-lo, ele também vai
passar pela substituição de variáveis e acabar indo parar no projeto gerado. Apague-o (ou
troque por um README próprio do seu template) quando terminar de customizar.

## Como funciona

Todo arquivo (e todo nome de arquivo/pasta) deste template passa por `text/template` do Go
quando alguém roda:

```bash
cpp-gen new meu-projeto --template <pasta-deste-template>
```

Os arquivos gerados por `cpp-gen templates create` já usam essas variáveis — abra
`CMakeLists.txt`, `README.md` e `src/{{.NameSnake}}.cpp` pra ver exemplos reais.

## Variáveis disponíveis

| Variável                | Exemplo         |
|---------------------------|-------------------|
| `{{.Name}}`               | `meu-projeto`      |
| `{{.NameSnake}}`          | `meu_projeto`      |
| `{{.NamePascal}}`         | `MeuProjeto`        |
| `{{.NameUpper}}`          | `MEU_PROJETO`       |
| `{{.Description}}`        | descrição do projeto (pode ser vazia) |
| `{{.Author}}`              | autor (pode ser vazio) |
| `{{.Version}}`             | `1.0.0`              |
| `{{.Standard}}`            | `17` / `20` / `23`   |
| `{{.UseVCPKG}}`            | `true` se `--pkg vcpkg` |
| `{{.UseFetchContent}}`     | `true` se `--pkg fetchcontent` |
| `{{.IsExecutable}}` / `{{.IsStaticLib}}` / `{{.IsHeaderOnly}}` | conforme `--type` |
| `{{.IsVSCode}}` / `{{.IsCLion}}` / `{{.IsNvim}}` / `{{.IsZed}}` | conforme `--ide` |

Lista completa em `docs/templates.md` no repositório do cpp-gen.

## Truques úteis

**Nome de arquivo condicional** — um arquivo cujo nome renderiza pra string vazia é
omitido inteiro. Este template já usa isso em
`{{"{{"}}if .UseVCPKG{{"}}"}}vcpkg.json{{"{{"}}end{{"}}"}}` — só existe quando alguém gera o
projeto com `--pkg vcpkg`.

**Conteúdo condicional** — `{{"{{"}}if .UseVCPKG{{"}}"}} ... {{"{{"}}else{{"}}"}} ... {{"{{"}}end{{"}}"}}`
funciona em qualquer lugar dentro de um arquivo.

**Arquivo não é um template válido?** Ele é copiado como está, sem quebrar a geração — útil
pra binários, ou C++ com `int a[2]{{"{{"}}1, 2{{"}}"}};` (brace-init colide com a sintaxe do
template).

## Testando

Sem precisar publicar nada:

```bash
cpp-gen new demo --template <pasta-deste-template> --no-interactive --output /tmp/scratch -v
```

## Deixando descobrível

```bash
cpp-gen templates create <pasta> --register meu-template   # já feito, se você usou --register
cpp-gen templates                                            # confirma que aparece
cpp-gen new meu-projeto --template meu-template               # usa pelo alias
```
