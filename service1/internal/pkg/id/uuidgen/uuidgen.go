package uuidgen

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrGenerating = errors.New("gen: failed to generate")
)

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Gen() (int, error) {
	return int(uuid.New().ID()), nil
}
