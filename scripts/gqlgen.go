// +build ignore

// This script runs gqlgen via the repo cmd rather than via the installed binary.
// Abides by the changes as described in https://github.com/99designs/gqlgen/issues/415.
// This maybe a temporary measure until the maintainers streamline this in the future.

package main

import "github.com/99designs/gqlgen/cmd"

func main() {
	cmd.Execute()
}
