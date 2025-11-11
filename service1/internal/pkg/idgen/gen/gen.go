package gen

import (
	"errors"
	"fmt"
	"strconv"
	"time"

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
	i, err := strconv.Atoi(fmt.Sprintf("%v%v", uuid.New().ID(), time.Now().UnixNano()))
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrGenerating, err)
	}
	return i, nil
}
