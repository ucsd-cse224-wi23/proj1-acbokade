package main

import "fmt"

func main() {
	x := make([] int, 0)
	// x = append(x, 1)
	// x = append(x, 2)
    double(x)
    fmt.Println(x) // ----> 3 will print [2, 20, 200, 2000] (original slice changed)
}

func double(y []int) {
    fmt.Println(y) // ----> 1 will print [1, 10, 100, 1000]
    // for i := 0; i < len(y); i++ {
    //     y[i] *= 2
    // }
    y = append(y, 1)
    fmt.Println(y) // ----> 2 will print [2, 20, 200, 2000] (copy slice + under array changed)
}