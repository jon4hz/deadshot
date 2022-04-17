package multicall

import (
	"testing"

	"github.com/jon4hz/deadshot/internal/database"

	"github.com/jon4hz/web3-go/ethrpc"
	"github.com/jon4hz/web3-multicall-go/multicall"
)

var client *Client

func teardown() {
	client.Close()

	if err := database.Close(); err != nil {
		panic(err)
	}
}

// unittest are executed on the polygon network
func init() {
	rpc, err := ethrpc.NewWithDefaults("https://polygon-rpc.com/")
	if err != nil {
		panic(err)
	}
	client, err = Init(rpc, "0x8a233a018a2e123c0D96435CF99c8e65648b429F")
	if err != nil {
		panic(err)
	}
}

/* func TestMain(m *testing.M) {
	teardown()
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	) /* goleak.IgnoreTopFunction("github.com/rjeczalik/notify.(*nonrecursiveTree).dispatch"),
	goleak.IgnoreTopFunction("github.com/rjeczalik/notify.(*nonrecursiveTree).internal"),
	goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"), * /
} */

func TestMulticallClient(t *testing.T) {
	vcs := multicall.ViewCalls{
		multicall.NewViewCall(
			"key-1",
			"0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270",
			"symbol()(string)",
			[]any{},
		),
	}
	res, err := client.call(vcs, nil)
	if err != nil {
		panic(err)
	}
	if !res.Calls["key-1"].Success {
		t.Error("Expected success, got false")
	}
	if res.Calls["key-1"].Decoded[0].(string) != "WMATIC" {
		t.Error("Expected WMATIC, got", res.Calls["key-1"].Decoded[0])
	}
}

func TestGetTokenInfo(t *testing.T) {
	tokens := []string{
		"0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270",
		"0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
		"0x8f3cf7ad23cd3cadbd9735aff958023239c6a063",
	}
	infos, err := client.GetTokenInfo(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != len(tokens) {
		t.Error("Expected", len(tokens), "tokens, got", len(infos))
	}
	for _, info := range infos {
		if info.Symbol == "" {
			t.Error("Expected token name, got empty string")
		}
		if info.Decimals == 0 {
			t.Error("Expected token decimals, got 0")
		}
		if info.Contract == "" {
			t.Error("Expected token contract, got empty string")
		}
	}
}

func TestGetPairInfo(t *testing.T) {
	pair := []string{"0xc45092e7e73951c6668f6C46AcFCa9F2B1c69aEf"} // WMATIC/UNI

	info, err := client.GetPairInfo(pair)
	if err != nil {
		t.Fatal(err)
	}
	if info["0xc45092e7e73951c6668f6C46AcFCa9F2B1c69aEf"].Token0.Contract != "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270" {
		t.Error("Expected WMATIC, got", info["0xc45092e7e73951c6668f6C46AcFCa9F2B1c69aEf"].Token0.Contract)
	}
	if info["0xc45092e7e73951c6668f6C46AcFCa9F2B1c69aEf"].Token1.Contract != "0xb33EaAd8d922B1083446DC23f610c2567fB5180f" {
		t.Error("Expected UNI, got", info["0xc45092e7e73951c6668f6C46AcFCa9F2B1c69aEf"].Token1.Contract)
	}
	t.Log(info["0xc45092e7e73951c6668f6C46AcFCa9F2B1c69aEf"].Token0.Symbol)
}

func TestGetPairToken(t *testing.T) {
	quickswapFactory := "0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32"
	tokens := []string{
		"0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270",
		"0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
	}
	tokenPairs := []TokenPair{
		{Token0: tokens[0], Token1: tokens[1], Factory: quickswapFactory},
	}

	pairs, err := client.GetPairToken(tokenPairs)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pairs)
}
