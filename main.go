// Main entry point for cpp-gen.
//
// cpp-gen is a CLI tool written in Go for generating modern C++ projects
// with CMake, package managers, IDE configurations and
// development tools such as Clangd and Clang-Format.
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
