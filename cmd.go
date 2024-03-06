package main

import (
	"fmt"
	"os"

	"github.com/ironiridis/enbypub/enbypub"
)

func main() {
	T, err := enbypub.LoadTextFromFile("example.md")
	fmt.Fprintf(os.Stderr, "err=%v\n", err)
	T.Emit(os.Stdout)
}
