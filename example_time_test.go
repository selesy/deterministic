package deterministic_test

import (
	"fmt"

	"github.com/selesy/deterministic"
)

func ExampleNowFunc() {
	now := deterministic.NowFunc()

	fmt.Println(now())
	fmt.Println(now())
	fmt.Println(now())
	// Output:
	// 2006-01-02 15:04:05 +0000 UTC
	// 2006-01-02 15:04:06 +0000 UTC
	// 2006-01-02 15:04:07 +0000 UTC
}
