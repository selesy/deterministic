package deterministic_test

import (
	"fmt"

	"github.com/selesy/deterministic"
)

func ExampleRandFunc() {
	randRead := deterministic.RandFunc()

	buf := make([]byte, 8)
	if _, err := randRead(buf); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("%x\n", buf)

	buf = make([]byte, 4)
	if _, err := randRead(buf); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("%x\n", buf)
	// Output:
	// deadbeef00010203
	// 04050607
}
