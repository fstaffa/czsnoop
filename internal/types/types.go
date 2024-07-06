package types

import (
	"fmt"
	"strconv"
)

type Ico string

func CreateIco(ico string) (Ico, error) {
	if len(ico) != 8 {
		return "", fmt.Errorf("Ico must be 8 characters long")
	}
	if _, err := strconv.Atoi(ico); err != nil {
		return "", fmt.Errorf("Ico must be a number")
	}
	return Ico(ico), nil
}

type Result[T any] struct {
	Result T
	Err    error
}
