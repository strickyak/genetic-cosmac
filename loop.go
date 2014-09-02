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

func Run() {
  w := new(ProductWorld)
  for gen := 0 ; ; gen++ {
    code := make([]byte, 64)
    rand.Read(code)
    sim, ok := RunSimulation(code, w)
    fit := w.Fitness(sim)
    if fit > 0.0 {
      fmt.Printf("%d\t%v\t%d\t%30.0f\n", gen, ok, sim.Time, fit)
    }
  }
}
