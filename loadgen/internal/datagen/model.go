package datagen

import "loadgen/internal/common"

type Employee struct {
	common.Name
	Salary   int
	Position int
	Email    string
}
