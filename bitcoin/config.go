package bitcoin

import (
	"os"
	"path"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/viper"
)

const (
	// MAINNET...
	MAINNET = "mainnet"
	// TESTNET...
	TESTNET = "testnet"
	// SIGNET...
	SIGNET = "signet"
	// SIMNET...
	SIMNET = "simnet"
	// REGTEST...
	REGTEST = "regtest"
)

// BitconConfig defines the bitcoin config
// TODO: defined different config group eg: bitcoin, bridge, indexer, commiter
type BitconConfig struct {
	// NetworkName defines the bitcoin network name
	NetworkName string `mapstructure:"network-name"`
	// RPCHost defines the bitcoin rpc host
	RPCHost string `mapstructure:"rpc-host"`
	// RPCPort defines the bitcoin rpc port
	RPCPort string `mapstructure:"rpc-port"`
	// RPCUser defines the bitcoin rpc user
	RPCUser string `mapstructure:"rpc-user"`
	// RPCPass defines the bitcoin rpc password
	RPCPass string `mapstructure:"rpc-pass"`
	// WalletName defines the bitcoin wallet name
	WalletName string `mapstructure:"wallet-name"`
	// Destination defines the taproot transaction destination address
	Destination string `mapstructure:"destination"`
	// EnableIndexer defines whether to enable the indexer
	EnableIndexer bool `mapstructure:"enable-indexer"`
	// EnableCommitter defines whether to enable the committer
	EnableCommitter bool `mapstructure:"enable-committer"`
	// IndexerListenAddress defines the address to listen on
	IndexerListenAddress string `mapstructure:"indexer-listen-address"`
	// Bridge defines the bridge config
	Bridge BridgeConfig `mapstructure:"bridge"`
	// SourceAddress defines the bitcoin send source address
	SourceAddress string `mapstructure:"source-address"`
	// Fee defines the bitcoin tx fee
	Fee int64 `mapstructure:"fee"`
	Evm struct {
		// EnableListener defines whether to enable the listener
		EnableListener bool `mapstructure:"enable-listener"`
		// RPCHost defines the evm rpc host
		RPCHost string `mapstructure:"rpc-host"`
		// RPCPort defines the evm rpc port
		RPCPort string `mapstructure:"rpc-port"`
		// ContractAddress defines the  contract address
		ContractAddress string `mapstructure:"contract-address"`
		// StartHeight defines the start height
		StartHeight int64 `mapstructure:"start-height"`
		// Deposit defines the deposit event hash
		Deposit string `mapstructure:"deposit"`
		// Withdraw defines the withdraw event hash
		Withdraw string `mapstructure:"withdraw"`
	}
}

type BridgeConfig struct {
	// EthRPCURL defines the ethereum rpc url
	EthRPCURL string `mapstructure:"eth-rpc-url"`
	// EthPrivKey defines the invoke ethereum private key
	EthPrivKey string `mapstructure:"eth-priv-key"`
	// ContractAddress defines the l1 -> l2 bridge contract address
	ContractAddress string `mapstructure:"contract-address"`
	// ABI defines the l1 -> l2 bridge contract abi
	ABI string `mapstructure:"abi"`
	// GasLimit defines the  contract gas limit
	GasLimit uint64 `mapstructure:"gas-limit"`
	// AASCARegistry defines the  contract AASCARegistry address
	AASCARegistry string `mapstructure:"aa-sca-registry"`
	// AAKernelFactory defines the  contract AAKernelFactory address
	AAKernelFactory string `mapstructure:"aa-kernel-factory"`
}

const (
	BitcoinRPCConfigFileName = "bitcoin.toml"
)

func LoadBitcoinConfig(homePath string) (*BitconConfig, error) {
	config := BitconConfig{}
	configFile := path.Join(homePath, BitcoinRPCConfigFileName)
	_, err := os.Stat(configFile)
	if err != nil {
		// if file not exist use default config
		// TODO: add gen config command after, The default configuration may not be required
		if os.IsNotExist(err) {
			config.NetworkName = "signet"
			config.RPCHost = "localhost"
			config.RPCPort = "8332"
			config.RPCUser = "user"
			config.RPCPass = "password"
			config.WalletName = "walletname"
			config.Destination = "tb1qgm39cu009lyvq93afx47pp4h9wxq5x92lxxgnz"
			config.IndexerListenAddress = "tb1qsja4hvx66jr9grgmt8452letmz37gmludcrnup"
			config.EnableIndexer = false
			config.EnableCommitter = false
			return &config, nil
		}
		return nil, err
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.AutomaticEnv()
	v.SetEnvPrefix("BITCOIN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ChainParams get chain params by network name
func ChainParams(network string) *chaincfg.Params {
	switch network {
	case MAINNET:
		return &chaincfg.MainNetParams
	case TESTNET:
		return &chaincfg.TestNet3Params
	case SIGNET:
		return &chaincfg.SigNetParams
	case SIMNET:
		return &chaincfg.SimNetParams
	case REGTEST:
		return &chaincfg.RegressionNetParams
	default:
		return &chaincfg.TestNet3Params
	}
}
