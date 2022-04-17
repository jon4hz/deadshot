package uniswap

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestRoute(t *testing.T) {
	token0, err := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000001"), "t0", "t0", 18)
	if err != nil {
		t.Fatal(err)
	}
	token1, err := NewToken(common.HexToAddress("0x0000000000000000000000000000000000000002"), "t1", "t1", 18)
	if err != nil {
		t.Fatal(err)
	}
	weth, err := NewToken(common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), "WETH", "WETH", 18)
	if err != nil {
		t.Errorf("expect no error, but got %v", err)
	}

	tokenAmount00, err := NewTokenAmount(token0, big.NewInt(100))
	if err != nil {
		t.Fatal(err)
	}
	tokenAmount01, err := NewTokenAmount(token1, big.NewInt(200))
	if err != nil {
		t.Fatal(err)
	}
	tokenAmount0Weth0, err := NewTokenAmount(token0, big.NewInt(100))
	if err != nil {
		t.Fatal(err)
	}
	tokenAmount0Weth1, err := NewTokenAmount(weth, big.NewInt(100))
	if err != nil {
		t.Fatal(err)
	}
	tokenAmount1Weth0, err := NewTokenAmount(token1, big.NewInt(175))
	if err != nil {
		t.Fatal(err)
	}
	tokenAmount1Weth1, err := NewTokenAmount(weth, big.NewInt(100))
	if err != nil {
		t.Fatal(err)
	}

	pair01, err := NewPair(common.HexToAddress("0x0000000000000000000000000000000000000002"), tokenAmount00, tokenAmount01)
	if err != nil {
		t.Fatal(err)
	}
	pair0Weth, err := NewPair(common.HexToAddress("0x0000000000000000000000000000000000000002"), tokenAmount0Weth0, tokenAmount0Weth1)
	if err != nil {
		t.Fatal(err)
	}
	pair1Weth, err := NewPair(common.HexToAddress("0x0000000000000000000000000000000000000002"), tokenAmount1Weth0, tokenAmount1Weth1)
	if err != nil {
		t.Fatal(err)
	}

	// constructs a path from the tokens
	{
		route, err := NewRoute([]*Pair{pair01}, token0, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(route.Pairs) != 1 || route.Pairs[0] != pair01 {
			t.Error("wrong pairs for route")
		}
		if len(route.Path) != 2 || route.Path[0] != token0 || route.Path[1] != token1 {
			t.Error("wrong path for route")
		}
		if route.Input != token0 {
			t.Error("wrong input for route")
		}
		if route.Output != token1 {
			t.Error("wrong output for route")
		}
	}

	// can have a token as both input and output
	{
		pairs := []*Pair{pair0Weth, pair01, pair1Weth}
		route, err := NewRoute(pairs, weth, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(route.Pairs) != len(pairs) {
			t.Fatal("wrong pairs for route")
		}
		for i, pair := range route.Pairs {
			if pair != pairs[i] {
				t.Error("wrong pairs for route")
			}
		}
		if route.Input != weth {
			t.Error("wrong input for route")
		}
		if route.Output != weth {
			t.Error("wrong output for route")
		}
	}

	{
		// supports ether output
		pairs := []*Pair{pair0Weth}
		route, err := NewRoute(pairs, token0, weth)
		if err != nil {
			t.Fatal(err)
		}
		if len(route.Pairs) != len(pairs) {
			t.Fatal("wrong pairs for route")
		}
		for i, pair := range route.Pairs {
			if pair != pairs[i] {
				t.Error("wrong pairs for route")
			}
		}
		if route.Input != token0 {
			t.Error("wrong input for route")
		}
		if route.Output != weth {
			t.Error("wrong output for route")
		}
	}
}
