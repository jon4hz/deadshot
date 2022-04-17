package uniswap

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// nolint funlen
func TestToken(t *testing.T) {
	addressOne := common.HexToAddress("0x0000000000000000000000000000000000000001")
	addressTwo := common.HexToAddress("0x0000000000000000000000000000000000000002")

	// fails if address differs
	{
		token1, err := NewToken(addressOne, "", "", 18)
		if err != nil {
			t.Fatal(err)
		}
		token2, err := NewToken(addressTwo, "", "", 18)
		if err != nil {
			t.Fatal(err)
		}
		if token1.equals(token2.Address()) {
			t.Error("should false if address differs")
		}
	}

	// true if only decimals differs
	{
		token1, err := NewToken(addressOne, "", "", 9)
		if err != nil {
			t.Fatal(err)
		}
		token2, err := NewToken(addressOne, "", "", 18)
		if err != nil {
			t.Fatal(err)
		}
		if !token1.equals(token2.Address()) {
			t.Error("should true if only decimals differs")
		}
	}

	// true if address is the same
	{
		token1, err := NewToken(addressOne, "", "", 18)
		if err != nil {
			t.Fatal(err)
		}
		token2, err := NewToken(addressOne, "", "", 18)
		if err != nil {
			t.Fatal(err)
		}
		if !token1.equals(token2.Address()) {
			t.Error("should true if address is the same")
		}
	}

	// true on reference equality
	{
		token, err := NewToken(addressOne, "", "", 18)
		if err != nil {
			t.Fatal(err)
		}
		if !token.equals(token.Address()) {
			t.Error("should true on reference equality")
		}
	}

	// true even if name/symbol/decimals differ
	{
		token1, err := NewToken(addressOne, "asdf", "ghjk", 18)
		if err != nil {
			t.Fatal(err)
		}
		token2, err := NewToken(addressOne, "qwer", "tzui", 18)
		if err != nil {
			t.Fatal(err)
		}
		if !token1.equals(token2.Address()) {
			t.Error("true even if name/symbol/decimals differ")
		}
	}
}
