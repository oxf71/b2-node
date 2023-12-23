package bitcoin

import (
	"strconv"
	"strings"
	"time"

	"github.com/evmos/ethermint/types"
	"github.com/tendermint/tendermint/libs/service"
	dbm "github.com/tendermint/tm-db"
)

const (
	ServiceName = "BitcoinIndexerService"

	BitcoinIndexBlockKey = "bitcoinIndexBlock" // key: currentBlock + "."+ currentTxIndex

	NewBlockWaitTimeout = 60 * time.Second
)

// IndexerService indexes transactions for json-rpc service.
type IndexerService struct {
	service.BaseService

	txIdxr types.BITCOINTxIndexer
	bridge types.BITCOINBridge

	db dbm.DB
}

// NewIndexerService returns a new service instance.
func NewIndexerService(
	txIdxr types.BITCOINTxIndexer,
	bridge types.BITCOINBridge,
	db dbm.DB,
) *IndexerService {
	is := &IndexerService{txIdxr: txIdxr, bridge: bridge, db: db}
	is.BaseService = *service.NewBaseService(nil, ServiceName, is)
	return is
}

// OnStart
func (bis *IndexerService) OnStart() error {
	latestBlock, err := bis.txIdxr.LatestBlock()
	if err != nil {
		bis.Logger.Error("bitcoin indexer latestBlock", "error", err.Error())
		return err
	}

	var (
		currentBlock   int64 // index current block number
		currentTxIndex int64 // index current block tx index
	)
	// btcIndexBlock
	btcIndexBlockMax, err := bis.db.Get([]byte(BitcoinIndexBlockKey))
	if err != nil {
		bis.Logger.Error("failed to get bitcoin index block from db", "error", err)
		return err
	}

	bis.Logger.Info("bitcoin indexer load db", "data", string(btcIndexBlockMax))

	// set default value
	currentBlock = latestBlock
	currentTxIndex = 0

	if btcIndexBlockMax != nil {
		indexBlock := strings.Split(string(btcIndexBlockMax), ".")
		bis.Logger.Info("bitcoin indexer db data split", "indexBlock", indexBlock)
		if len(indexBlock) > 1 {
			currentBlock, err = strconv.ParseInt(indexBlock[0], 10, 64)
			if err != nil {
				bis.Logger.Error("failed to parse block", "error", err)
				return err
			}
			currentTxIndex, err = strconv.ParseInt(indexBlock[1], 10, 64)
			if err != nil {
				bis.Logger.Error("failed to parse tx index", "error", err)
				return err
			}
		}
	}
	bis.Logger.Info("bitcoin indexer init data", "latestBlock", latestBlock,
		"currentBlock", currentBlock, "db data", string(btcIndexBlockMax), "currentTxIndex", currentTxIndex)

	ticker := time.NewTicker(NewBlockWaitTimeout)
	for {
		bis.Logger.Info("bitcoin indexer", "latestBlock",
			latestBlock, "currentBlock", currentBlock, "currentTxIndex", currentTxIndex)

		if latestBlock <= currentBlock {
			<-ticker.C
			ticker.Reset(NewBlockWaitTimeout)

			// update latest block
			latestBlock, err = bis.txIdxr.LatestBlock()
			if err != nil {
				bis.Logger.Error("bitcoin indexer latestBlock", "error", err.Error())
			}
			continue
		}

		// index > 0, start index from currentBlock currentTxIndex + 1
		// index == 0, start index from currentBlock + 1
		if currentTxIndex == 0 {
			currentBlock++
		} else {
			currentTxIndex++
		}

		for i := currentBlock; i <= latestBlock; i++ {
			txResults, err := bis.txIdxr.ParseBlock(i, currentTxIndex)
			if err != nil {
				bis.Logger.Error("bitcoin indexer parseblock", "error", err.Error(), "currentBlock", i, "currentTxIndex", currentTxIndex)
				continue
			}

			if len(txResults) > 0 {
				for _, v := range txResults {
					// if from is listen address, skip
					if v.From[0] == v.To {
						bis.Logger.Info("bitcoin indexer current transaction from is listen address", "currentBlock", i, "currentTxIndex", v.Index, "data", v)
						continue
					}
					hex, err := bis.bridge.Deposit(v.TxId, v.From[0], v.Value)
					if err != nil {
						if strings.Contains(err.Error(), ErrBrdigeDepositTxIDExist.Error()) {
							bis.Logger.Error("bitcoin indexer current transaction deposited", "error", err.Error(),
								"currentBlock", i, "currentTxIndex", v.Index, "data", v)
						} else {
							// TODO: only wirte log, not return
							bis.Logger.Error("bitcoin indexer invoke deposit bridge unknown err", "error", err.Error(),
								"currentBlock", i, "currentTxIndex", v.Index, "data", v)
						}
					}
					currentBlockStr := strconv.FormatInt(i, 10)
					currentTxIndexStr := strconv.FormatInt(v.Index, 10)
					err = bis.db.Set([]byte(BitcoinIndexBlockKey), []byte(currentBlockStr+"."+currentTxIndexStr))
					if err != nil {
						bis.Logger.Error("failed to set bitcoin index block", "error", err)
					}
					bis.Logger.Info("bitcoin indexer invoke deposit bridge", "deposit data", v, "hex", hex)
				}
			}

			currentBlock = i
			currentTxIndex = 0

			currentBlockStr := strconv.FormatInt(currentBlock, 10)
			currentTxIndexStr := strconv.FormatInt(currentTxIndex, 10)
			err = bis.db.Set([]byte(BitcoinIndexBlockKey), []byte(currentBlockStr+"."+currentTxIndexStr))
			if err != nil {
				bis.Logger.Error("failed to set bitcoin index block", "error", err)
			}
			bis.Logger.Info("bitcoin indexer parsed", "txResult", txResults, "currentBlock", i,
				"currentTxIndex", currentTxIndex, "latestBlock", latestBlock)
		}
	}
}
