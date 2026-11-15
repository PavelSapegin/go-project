package main

import "fmt"

func twoSum(nums []int, target int) []int {
	history := make(map[int]int)

	for i := range len(nums) {
		val, ok := history[target-nums[i]]
		if ok {
			return []int{val, i}
		} else {
			history[nums[i]] = i
		}
	}

	return []int{0, 0}
}

func main() {

	arr1 := []int{1, 2, 3, 4, 5}
	target1 := 7
	expected1 := []int{2, 3}

	arr2 := []int{5, 11, 12, 43}
	target2 := 23
	expected2 := []int{1, 2}

	arr3 := []int{1, 45, 65, 32, 65, 78, 90}
	target3 := 122
	expected3 := []int{3, 6}

	fmt.Println("Test1:")
	fmt.Printf("Nums = %v; target = %d; output = %v; expected = %v\n",
		arr1, target1, twoSum(arr1, target1), expected1)

	fmt.Println("Test2:")
	fmt.Printf("Nums = %v; target = %d; output = %v; expected = %v\n",
		arr2, target2, twoSum(arr2, target2), expected2)

	fmt.Println("Test3:")
	fmt.Printf("Nums = %v; target = %d; output = %v; expected = %v\n",
		arr3, target3, twoSum(arr3, target3), expected3)

}
