package main

import "fmt"

func must1[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Errorf("cannot return %T: %w", v, err))
	}
	return v
}

func must(task string, err error) {
	if err != nil {
		panic(fmt.Errorf("cannot %s: %w", task, err))
	}
}
