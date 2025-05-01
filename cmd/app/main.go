package main

import "fmt"

func main() {
	slice := make([]string, 0, 10)

	func() {
		slice = append(slice, "hello")
	}()

	fmt.Println(slice)
}
