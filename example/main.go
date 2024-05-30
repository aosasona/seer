package main

import (
	"fmt"

	"go.trulyao.dev/seer"
)

func main() {
	if err := run(); err != nil {
		fmt.Print(err)
	}
}

func run() error {
	var val int

	fmt.Println("Enter an odd number: ")

	_, err := fmt.Scan(&val)
	if err != nil {
		return seer.Wrap("collectInput", err)
	}

	if val <= 0 {
		return seer.New("validateInput", fmt.Sprintf("negative number or zero: %d", val))
	}

	if val%2 == 0 {
		return seer.New("validateInput", fmt.Sprintf("even number: %d", val))
	}

	return nil
}
