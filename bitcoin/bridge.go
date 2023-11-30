package bitcoin

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path"

	b2aa "github.com/b2network/b2-go-aa-utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Bridge bridge
// TODO: only L1 -> L2, More calls may be supported later
type Bridge struct {
	EthRPCURL       string
	EthPrivKey      *ecdsa.PrivateKey
	ContractAddress common.Address
	ABI             string
	GasLimit        uint64
	// AA contract address
	AASCARegistry   common.Address
	AAKernelFactory common.Address
}

const (
	B2DevRpcUrl   = "http://43.135.203.73:8123"
	SCARegistry   = "0x231aec684Ad0e63c2F4d176EddCE97A1B666247c"
	KernelFactory = "0x7516283Ff7090B8286E23a16f8b5b35B3ba541A2"
)

// NewBridge new bridge
func NewBridge(bridgeCfg BridgeConfig, abiFileDir string) (*Bridge, error) {
	rpcURL, err := url.ParseRequestURI(bridgeCfg.EthRPCURL)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(bridgeCfg.EthPrivKey)
	if err != nil {
		return nil, err
	}

	abi, err := os.ReadFile(path.Join(abiFileDir, bridgeCfg.ABI))
	if err != nil {
		return nil, err
	}

	return &Bridge{
		EthRPCURL:       rpcURL.String(),
		ContractAddress: common.HexToAddress(bridgeCfg.ContractAddress),
		EthPrivKey:      privateKey,
		ABI:             string(abi),
		GasLimit:        bridgeCfg.GasLimit,
		AASCARegistry:   common.HexToAddress(bridgeCfg.AASCARegistry),
		AAKernelFactory: common.HexToAddress(bridgeCfg.AAKernelFactory),
	}, nil
}

// Deposit to ethereum
func (b *Bridge) Deposit(bitcoinAddress string, amount int64) error {
	if bitcoinAddress == "" {
		return fmt.Errorf("bitcoin address is empty")
	}

	ctx := context.Background()

	toAddress, err := b.BitcoinAddressToEthAddress(bitcoinAddress)
	if err != nil {
		return err
	}

	data, err := b.ABIPack(b.ABI, "deposit", common.HexToAddress(toAddress), new(big.Int).SetInt64(amount))
	if err != nil {
		return err
	}

	receipt, err := b.ethContractCall(ctx, b.EthPrivKey, data)
	if err != nil {
		return err
	}

	if receipt.Status != 1 {
		return fmt.Errorf("tx failed, receipt:%v", receipt)
	}
	return nil
}

func (b *Bridge) ethContractCall(ctx context.Context, priv *ecdsa.PrivateKey, data []byte) (*types.Receipt, error) {
	client, err := ethclient.Dial(b.EthRPCURL)
	if err != nil {
		return nil, err
	}

	publicKey := priv.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	nonce, err := client.PendingNonceAt(ctx, crypto.PubkeyToAddress(*publicKeyECDSA))
	if err != nil {
		return nil, err
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &b.ContractAddress,
		Value:    big.NewInt(0),
		Gas:      b.GasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	// sign tx
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), priv)
	if err != nil {
		return nil, err
	}

	// send tx
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	// wait tx confirm
	return bind.WaitMined(ctx, client, signedTx)
}

// ABIPack the given method name to conform the ABI. Method call's data
func (b *Bridge) ABIPack(abiData string, method string, args ...interface{}) ([]byte, error) {
	contractAbi, err := abi.JSON(bytes.NewReader([]byte(abiData)))
	if err != nil {
		return nil, err
	}
	return contractAbi.Pack(method, args...)
}

// BitcoinAddressToEthAddress bitcoin address to eth address
func (b *Bridge) BitcoinAddressToEthAddress(bitcoinAddress string) (string, error) {
	client, err := ethclient.Dial(b.EthRPCURL)
	if err != nil {
		return "", err
	}

	targetEthAddress, err := b2aa.GetSCAAddress(client, b.AASCARegistry, b.AAKernelFactory, bitcoinAddress)
	if err != nil {
		return "", err
	}
	return targetEthAddress.String(), nil
}
