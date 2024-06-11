package hash

import (
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	test := Teacher{
		Name: "张三",
		Age:  21,
	}
	fmt.Println(Hash(test))
}

type Teacher struct {
	Name string
	Age  int32
}
