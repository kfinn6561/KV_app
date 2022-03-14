package KVStore

import "time"

func doneWithTimeout(done chan struct{}, timeout time.Duration) chan struct{} {
	outChan := make(chan struct{})
	go func() {
		defer close(outChan)
		select {
		case <-done:
		case <-time.After(timeout):
		}
	}()
	return outChan
}

// IntPow calculates x to the nth power modulo n. Since the result is an int, it is assumed that n is a positive power
func IntPowMod(x, n, m int) int {
	if x == 0 {
		return 1
	}
	result := x % m
	for i := 2; i <= n; i++ {
		result = (result * x) % m //take a mod at each step to avoid storing a big number
	}
	return result
}

//hash uses a polynomial rolling hashing algorithm to convert a string into a number between 0 and 1
func hash(input string) float64 {
	p := 122          //this is roughly the number of characters
	m := int(1e9 + 9) //a big prime number
	var output int

	for i := 0; i < len(input); i++ {
		output += (int(input[i]) * IntPowMod(p, i, m)) % m
	}
	output = output % m

	return float64(output) / float64(m)
}
