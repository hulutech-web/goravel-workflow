package workflow

import "testing"

func TestUniqueSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 1}
	seen := make(map[int]bool)
	result := []int{}

	for _, value := range slice {
		if _, ok := seen[value]; !ok {
			seen[value] = true
			result = append(result, value)
		}
	}
	//	断言sliceresult是否为[1,2,3,4,5]
	if len(result) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(result))
	}
	for i, value := range result {
		if value != i+1 {
			t.Errorf("Expected %d, got %d", i+1, value)
		}
	}
}
