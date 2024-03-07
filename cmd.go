package main

import (
	"fmt"
	"os"
)

func main() {
	T, err := WalkContent()
	fmt.Fprintf(os.Stderr, "err=%v\n", err)
	for k := range T {
		T[k].Emit(os.Stdout)
	}
}
