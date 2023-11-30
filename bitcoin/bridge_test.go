package bitcoin_test

import (
	"errors"
	"math/big"
	"os"
	"path"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/bitcoin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBridge(t *testing.T) {
	abiPath := path.Join("./testdata")

	abi, err := os.ReadFile(path.Join("./testdata", "abi.json"))
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}

	bridgeCfg := bitcoin.BridgeConfig{
		EthRPCURL:       "http://localhost:8545",
		ContractAddress: "0x123456789abcdef",
		EthPrivKey:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ABI:             "abi.json",
		GasLimit:        1000000,
		AASCARegistry:   "0x123456789abcdefgh",
		AAKernelFactory: "0x123456789abcdefg",
	}

	bridge, err := bitcoin.NewBridge(bridgeCfg, abiPath)
	assert.NoError(t, err)
	assert.NotNil(t, bridge)
	assert.Equal(t, bridgeCfg.EthRPCURL, bridge.EthRPCURL)
	assert.Equal(t, common.HexToAddress("0x123456789abcdef"), bridge.ContractAddress)
	assert.Equal(t, privateKey, bridge.EthPrivKey)
	assert.Equal(t, string(abi), bridge.ABI)
	assert.Equal(t, common.HexToAddress("0x123456789abcdefgh"), bridge.AASCARegistry)
	assert.Equal(t, common.HexToAddress("0x123456789abcdefg"), bridge.AAKernelFactory)
}

// TestLocalDeposit only test in local
func TestLocalDeposit(t *testing.T) {
	bridge := bridgeWithConfig(t)
	testCase := []struct {
		name string
		args []interface{}
		err  error
	}{
		{
			name: "success",
			args: []interface{}{"tb1qjda2l5spwyv4ekwe9keddymzuxynea2m2kj0qy", int64(1234)},
			err:  nil,
		},
		{
			name: "fail: address empty",
			args: []interface{}{"", int64(1234)},
			err:  errors.New("bitcoin address is empty"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			err := bridge.Deposit(tc.args[0].(string), tc.args[1].(int64))
			if err != nil {
				assert.Equal(t, tc.err, err)
			}
		})
	}
}

// TestLocalBitcoinAddressToEthAddress only test in local
func TestLocalBitcoinAddressToEthAddress(t *testing.T) {

}

func TestABIPack(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		abiData, err := os.ReadFile(path.Join("./testdata", "abi.json"))
		if err != nil {
			t.Fatal(err)
		}
		expectedMethod := "deposit"
		expectedArgs := []interface{}{common.HexToAddress("0x12345678"), new(big.Int).SetInt64(1111)}
		expectedResult := []byte{71, 231, 239, 36, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 18, 52, 86, 120, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 4, 87}

		// Create a mock bridge object
		b := &bitcoin.Bridge{}

		// Call the ABIPack method
		result, err := b.ABIPack(string(abiData), expectedMethod, expectedArgs...)

		// Check for errors
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Compare the result with the expected result
		require.Equal(t, result, expectedResult)
	})

	t.Run("Invalid ABI data", func(t *testing.T) {
		abiData := `{"inputs": [{"type": "address", "name": "to"}, {"type": "uint256", "name": "value"}`
		expectedError := errors.New("unexpected EOF")

		// Create a mock bridge object
		b := &bitcoin.Bridge{}

		// Call the ABIPack method
		_, err := b.ABIPack(abiData, "method", "arg1", "arg2")

		require.EqualError(t, err, expectedError.Error())
	})
}

func bridgeWithConfig(t *testing.T) *bitcoin.Bridge {
	config, err := bitcoin.LoadBitcoinConfig("./testdata")
	require.NoError(t, err)

	bridge, err := bitcoin.NewBridge(config.Bridge, "./testdata")
	require.NoError(t, err)
	return bridge
}
