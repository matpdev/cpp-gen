// Ponto de entrada principal do cpp-gen.
//
// cpp-gen é uma ferramenta CLI escrita em Go para geração de projetos C++
// modernos com CMake, gerenciadores de pacotes, configurações de IDE e
// ferramentas de desenvolvimento como Clangd e Clang-Format.
package main

import (
	"fmt"
	"os"

	"cpp-gen/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
