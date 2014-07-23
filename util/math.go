package util

import "math"

func ErfInverse(z float64, n int) float64 {
	if n == 0 {
		return 0
	}
	ck := make([]float64, n)
	ck[0] = 1
	e := 0.0
	for k := 0; k < n; k++ {
		for i := 0; i < k; i++ {
			ck[k] += (ck[i] * ck[k-1-i]) / ((float64(i) + 1) * (2*float64(i) + 1))
		}
		e += (ck[k] / (2*float64(k) + 1)) * math.Pow(z*math.Sqrt(math.Pi)/2, 2*float64(k)+1)
	}
	return e
}
