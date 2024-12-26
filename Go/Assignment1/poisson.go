package main

import (
    "log"
    "math"
    "math/rand"
)

// PoissonProcess simulates a Poisson process
type PoissonProcess struct {
    lambda float64 // rate parameter
    rng    *rand.Rand // random number generator
}

// NewPoissonProcess creates a new PoissonProcess with a given rate and random seed
func NewPoissonProcess(lambda float64, seed int64) *PoissonProcess {
    if lambda <= 0 {
        log.Fatalf("Supplied rate parameter must be positive: %f", lambda)
    }
    rng := rand.New(rand.NewSource(seed))
    return &PoissonProcess{lambda: lambda, rng: rng}
}

// TimeForNextEvent generates the time until the next event based on the exponential distribution
func (pp *PoissonProcess) TimeForNextEvent() float64 {
    return -math.Log(1.0-pp.rng.Float64()) / pp.lambda
}