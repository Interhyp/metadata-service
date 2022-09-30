package docs

import (
	"fmt"
)

// these do nothing really, but they make tests and their log output way more readable

func Given(s string) {
	fmt.Println("    " + s)
}

func When(s string) {
	fmt.Println("    " + s)
}

func Then(s string) {
	fmt.Println("    " + s)
}

func Description(s string) {
	fmt.Println("    " + s)
}

func Limitation(s string) {
	fmt.Println("LIMITATION: " + s)
}
