package main

import (
	"fmt"

	"go.trulyao.dev/seer"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %s, Code: %d\n", err, err.(*seer.Seer).Code())
		fmt.Print(err.(*seer.Seer).ErrorWithStackTrace())
	}
}

func run() error {
	var val int

	fmt.Println("Enter an odd number: ")

	_, err := fmt.Scan(&val)
	if err != nil {
		return seer.Wrap("collectInput", err).WithCode(500)
	}

	if val <= 0 {
		return seer.New("validateInput", fmt.Sprintf("negative number or zero: %d", val)).
			WithCode(400)
	}

	if val%2 == 0 {
		return seer.New("validateInput", fmt.Sprintf("even number: %d", val)).WithCode(400)
	}

	return nil
}
