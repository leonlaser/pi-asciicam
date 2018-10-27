package main

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// Linear BrightnessMethod to brighten an image
	Linear int = 1
	// Exponential BrightnessMethod to brighten an imge
	Exponential int = 2
)

// BrightnessMethod describes the way an image should be brightened, "linear" or "exponential"
type BrightnessMethod struct {
	value int
}

// Set is uses to set the wanten method
func (method *BrightnessMethod) Set(s string) error {
	switch strings.ToLower(s) {
	case "linear":
		method.value = Linear
	case "exponential":
		method.value = Exponential
	default:
		return errors.New("The brightness method is not available")
	}
	return nil
}

func (method *BrightnessMethod) String() string {
	switch method.value {
	case Linear:
		return "Linear"
	case Exponential:
		return "Exponential"
	default:
		return fmt.Sprintf("Method '%d' does not exist", method.value)
	}
}

// Is tests if the BrightnessMethod equals one of the value constants
func (method *BrightnessMethod) Is(value int) bool {
	return method.value == value
}
