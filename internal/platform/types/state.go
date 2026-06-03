package types

import "fmt"

type State interface {
	fmt.Stringer
	IsValid() bool
}
