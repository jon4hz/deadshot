package uniswap

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrInvalidPairs         = errors.New("invalid pairs")
	ErrInvalidPairsChainIDs = errors.New("invalid pairs chainIDs")
	ErrInvalidInput         = errors.New("invalid token input")
	ErrInvalidOutput        = errors.New("invalid token output")
	ErrInvalidPath          = errors.New("invalid pairs for path")
)

type Route struct {
	Pairs    []*Pair
	Path     []*Token
	Input    *Token
	Output   *Token
	MidPrice *Price
}

func NewRoute(pairs []*Pair, input, output *Token) (*Route, error) {
	if len(pairs) == 0 {
		return nil, ErrInvalidPairs
	}

	if !pairs[0].InvolvesToken(input) {
		return nil, ErrInvalidInput
	}
	if !(output == nil || pairs[len(pairs)-1].InvolvesToken(output)) {
		return nil, ErrInvalidOutput
	}

	path := make([]*Token, len(pairs)+1)
	path[0] = input
	for i := range pairs {
		currentInput := path[i]
		if !(currentInput.equals(pairs[i].Token0().Address()) || currentInput.equals(pairs[i].Token1().Address())) {
			return nil, ErrInvalidPath
		}
		currentOutput := pairs[i].Token0()
		if currentInput.equals(pairs[i].Token0().Address()) {
			currentOutput = pairs[i].Token1()
		}
		path[i+1] = currentOutput
	}

	if output == nil {
		output = path[len(pairs)]
	}

	route := &Route{
		Pairs:  pairs,
		Path:   path,
		Input:  input,
		Output: output,
	}
	var err error
	route.MidPrice, err = NewPriceFromRoute(route)
	return route, err
}

// GetAddresses returns the addresses of the tokens involved in the route.
func (r *Route) GetAddresses() []common.Address {
	addresses := make([]common.Address, len(r.Path))
	for i := range r.Path {
		addresses[i] = r.Path[i].CommonAddress()
	}
	return addresses
}
