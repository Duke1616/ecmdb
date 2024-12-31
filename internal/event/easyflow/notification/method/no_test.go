package method

import (
	"fmt"
	"testing"
)

func TestFrom(t *testing.T) {
	var userIDs = []string{"1", "2", "3", "4", "5"}
	var notFinishUsers = []string{"1", "2", "3"}
	for i, u := range userIDs {
		for _, nu := range notFinishUsers {
			if u == nu {
				userIDs = RemoveFromSlice[string](userIDs, i)
			}
		}
	}
	fmt.Printf("case1:expect [4, 5], got %v\n", userIDs)

	var userIDs2 = []string{"1", "2", "3", "4", "5"}
	var notFinishUsers2 = []string{"4", "5"}
	//出现越界访问错误，退出程序
	for i, u := range userIDs2 {
		for _, nu := range notFinishUsers2 {
			if u == nu {
				userIDs2 = RemoveFromSlice[string](userIDs2, i)
			}
		}
	}
	fmt.Printf("case2: expect [1, 2, 3], got %v\n", userIDs2)
}

// 从切片中删除对应Index的项
func RemoveFromSlice[T any](Slice []T, RemoveItemIndex int) []T {
	NewSlice := append(Slice[:RemoveItemIndex], Slice[RemoveItemIndex+1:]...)
	return NewSlice
}
