package testentity

import (
	"bytes"
	"encoding/hex"
	"github.com/Ankr-network/ankrscan-proto-contract/go/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	proto2 "github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strings"
	"testing"
)

func ProtoEqual(t *testing.T, message1 proto2.Message, message2 proto2.Message) {
	buf1 := proto2.NewBuffer([]byte{})
	require.NoError(t, buf1.EncodeMessage(message1))
	buf2 := proto2.NewBuffer([]byte{})
	require.NoError(t, buf2.EncodeMessage(message2))
	require.Equal(t, len(buf1.Bytes()), len(buf2.Bytes()))
	require.Equal(t, 0, bytes.Compare(buf1.Bytes(), buf2.Bytes()))
}

func GenAddress() []byte {
	token := make([]byte, 20)
	rand.Read(token)
	return token
}

func GenHash() []byte {
	token := make([]byte, 32)
	rand.Read(token)
	return token
}

func GenBlockHeight() uint64 {
	return uint64(rand.Uint32())
}

func GenBytes(size int) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func GetSequenceParent(blocks []*proto.Block) []byte {
	return blocks[len(blocks)-1].Header.BlockHash
}

func GenSequence(blocksCount int, blockchainName string, blockHeight uint64, blockHash []byte, parentHash []byte, txsCount int) []*proto.Block {
	blocks := make([]*proto.Block, 0)
	for i := 0; i < blocksCount; i++ {
		blocks = append(blocks, GenBlockWithTxs(blockchainName, blockHeight+uint64(i), blockHash, parentHash, txsCount))
		parentHash = blockHash
		blockHash = GenHash()
	}
	return blocks
}

func GenBlockWithTxs(blockchainName string, blockHeight uint64, blockHash []byte, parentHash []byte, txsCount int) *proto.Block {
	transactions := make([]*proto.Transaction, 0)
	for i := 0; i < txsCount; i++ {
		transactions = append(transactions, GenTx(blockchainName, blockHeight, blockHash, GenHash(), uint64(i), 500, 10, 40))
	}
	return GenBlock(blockchainName, blockHeight, blockHash, parentHash, transactions)
}

func GenBlockWithLargeTxs(blockchainName string, blockHeight uint64, blockHash []byte, txsCount int, inputSize int) *proto.Block {
	transactions := make([]*proto.Transaction, 0)
	for i := 0; i < txsCount; i++ {
		transactions = append(transactions, GenTx(blockchainName, blockHeight, blockHash, GenHash(), uint64(i), inputSize, 10, 40))
	}
	return GenBlock(blockchainName, blockHeight, blockHash, GenHash(), transactions)
}

func GenBlock(blockchainName string, blockHeight uint64, blockHash []byte, parentHash []byte, transactions []*proto.Transaction) *proto.Block {
	block := &proto.Block{
		Header: &proto.BlockHeader{
			BlockchainName:    blockchainName,
			BlockHeight:       blockHeight,
			BlockHash:         blockHash,
			ParentHash:        parentHash,
			Timestamp:         rand.Uint64(),
			TransactionsCount: uint64(len(transactions)),
			Specific: &proto.BlockHeader_EthBlock{
				EthBlock: &proto.EthBlock{
					Nonce:            GenAddress(),
					Sha3Uncles:       GenHash(),
					TransactionsRoot: GenHash(),
					StateRoot:        GenHash(),
					Miner:            GenAddress(),
					Difficulty:       GenHash(),
					TotalDifficulty:  GenHash(),
					ExtraData:        GenHash(),
					Size:             rand.Uint64(),
					GasLimit:         rand.Uint64(),
					GasUsed:          rand.Uint64(),
					Uncles:           [][]byte{GenHash(), GenHash()},
					LogsBloom:        GenLogsBloom(transactions),
				},
			},
		},
		Transactions: transactions,
	}
	return block
}

func GenTxWithLogs(blockchainName string, blockHeight uint64, blockHash []byte, transactionHash []byte, transactionIndex uint64, inputSize int, logs []*proto.EthLog) *proto.Transaction {
	return &proto.Transaction{
		BlockchainName:   blockchainName,
		TransactionHash:  transactionHash,
		BlockHash:        blockHash,
		BlockHeight:      blockHeight,
		TransactionIndex: transactionIndex,
		Timestamp:        rand.Uint64(),
		Specific: &proto.Transaction_EthTx{
			EthTx: &proto.EthTransaction{
				Nonce:             rand.Uint64(),
				From:              GenAddress(),
				To:                GenAddress(),
				Value:             GenHash(),
				Gas:               rand.Uint64(),
				GasPrice:          GenHash(),
				Input:             GenBytes(inputSize),
				ContractAddress:   GenAddress(),
				CumulativeGasUsed: rand.Uint64(),
				GasUsed:           rand.Uint64(),
				Status:            1,
				Logs:              logs,
			},
		},
	}
}

func GenTx(blockchainName string, blockHeight uint64, blockHash []byte, transactionHash []byte, transactionIndex uint64, inputSize, logsCount, logsDataSize int) *proto.Transaction {
	logs := make([]*proto.EthLog, 0)
	for i := 0; i < logsCount; i++ {
		logs = append(logs, GenLog(uint32(i), logsDataSize))
	}
	return GenTxWithLogs(blockchainName, blockHeight, blockHash, transactionHash, transactionIndex, inputSize, logs)
}

func GenLog(logIndex uint32, dataSize int) *proto.EthLog {
	return &proto.EthLog{
		Address: GenAddress(),
		Topics:  [][]byte{GenHash(), GenHash()},
		Data:    GenBytes(dataSize),
		Removed: false,
	}
}

func GenLogsBloom(transactions []*proto.Transaction) []byte {
	var bin types.Bloom
	for _, tx := range transactions {
		for _, log := range tx.Specific.(*proto.Transaction_EthTx).EthTx.Logs {
			bin.Add(log.Address)
			for _, topic := range log.Topics {
				bin.Add(topic)
			}
		}
	}
	return bin[:]
}

func GenBlockchainName() string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return strings.Join([]string{string(b), "chain"}, "")
}

func GenCurrencyName() string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 3)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return strings.Join([]string{string(b), "currency"}, "")
}

func CheckBlocks(t *testing.T, initial []*proto.Block, result []*proto.Block) {
	require.Equal(t, len(initial), len(result))
	for i := 0; i < len(initial); i++ {
		require.Equal(t, initial[i].Header.BlockchainName, result[i].Header.BlockchainName, "index %d", i)
		require.Equal(t, int(initial[i].Header.BlockHeight), int(result[i].Header.BlockHeight), "index %d", i)
		require.Equal(t, common.BytesToHash(initial[i].Header.BlockHash).String(), common.BytesToHash(result[i].Header.BlockHash).String(), "index %d", i)
		ProtoEqual(t, initial[i].Header, result[i].Header)
		ProtoEqual(t, initial[i], result[i])
	}
}

func GenCurrency(blockchainName string, address []byte, decimals uint64) *proto.CurrencyDetails {
	return &proto.CurrencyDetails{
		BlockchainName: blockchainName,
		Address:        address,
		Name:           GenCurrencyName(),
		Decimals:       decimals,
		Symbol:         GenCurrencyName(),
	}
}

func GenConsumer(blockchainName string) *proto.BlockConsumer {
	return &proto.BlockConsumer{
		BlockchainName: blockchainName,
		ConsumerName:   "test-consumer",
		UserId:         hex.EncodeToString(GenHash()),
	}
}
