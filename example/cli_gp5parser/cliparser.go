package main

import (
	"fmt"

	"github.com/knasan/parsegp"
)

func main() {

	p, err := parsegp.NewParser("../testfiles/gp5/ready_or_not_2.gp5")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Version:", p.Version)
}
