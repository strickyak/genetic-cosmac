package genetic

import "fmt"
import "crypto/rand"

type ProductWorld struct {
}

func (pw *ProductWorld) Tick(sim *Sim) bool {
	return true
}

func (pw *ProductWorld) Tock(sim *Sim) bool {
	return true
}

func (pw *ProductWorld) Fitness(sim *Sim) float64 {
	z := 1.0
	for _, e := range sim.M.Out {
		z *= (1.0 + float64(e))
	}
	return z
}

func Run1(ch chan float64) {
	w := new(ProductWorld)
	code := make([]byte, 64)
	rand.Read(code)
	sim, ok := RunSimulation(code, w)
	fit := w.Fitness(sim)
	if fit > 0.0 {
		fmt.Printf("%v\t%d\t%30.0f\n", ok, sim.Time, fit)
	}

	ch <- fit
}
func RunN(n int) float64 {
	ch := make(chan float64)
	var z float64
	for i := 0; i < n; i++ {
		go Run1(ch)
	}
	for i := 0; i < n; i++ {
		fit := <-ch
		fmt.Printf("FIT %30.0f\n", fit)
		if z < fit {
			z = fit // Maximum.
		}
	}
	return z
}
