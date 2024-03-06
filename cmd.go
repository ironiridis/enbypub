package main

import (
	"fmt"
	"os"

	"github.com/ironiridis/enbypub/enbypub"
)

func main() {
	T, err := enbypub.LoadTextFromFile("example.md")
	fmt.Fprintf(os.Stderr, "err=%v\nT=%+v\n", err, T)
	// html.NewRenderer().Render(os.Stdout, T.Raw, T.Body)
}
