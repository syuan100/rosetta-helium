package helium

import (
	"fmt"
	"log"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/syuan100/rosetta-helium/utils"
	"github.com/ybbus/jsonrpc"
)

var (
	NodeClient = jsonrpc.NewClient("http://localhost:4467")
)

func CurrentBlockHeight() *int64 {

	var result int64

	if err := NodeClient.CallFor(&result, "block_height", nil); err != nil {
		log.Fatal(err)
	}

	return &result
}

func GetBlock(blockIdentifier *types.PartialBlockIdentifier) (*types.Block, *types.Error) {

	type request struct {
		Height int64  `json:"height,omitempty"`
		Hash   string `json:"hash,omitempty"`
	}

	var result Block
	var req request

	if blockIdentifier.Index != nil && blockIdentifier.Hash != nil {
		req = request{
			Height: *blockIdentifier.Index,
		}
	} else if blockIdentifier.Index == nil && blockIdentifier.Hash != nil {
		req = request{
			Hash: *blockIdentifier.Hash,
		}
	} else if blockIdentifier.Index != nil && blockIdentifier.Hash == nil {
		req = request{
			Height: *blockIdentifier.Index,
		}
	}

	if err := NodeClient.CallFor(&result, "block_get", req); err != nil {
		return nil, WrapErr(ErrNotFound, err)
	}

	var processedTxs []*types.Transaction
	for _, tx := range result.Transactions {
		ptx, txErr := GetTransaction(tx)
		if txErr != nil {
			return nil, txErr
		}

		processedTxs = append(processedTxs, ptx)
	}

	currentBlock := &types.Block{
		BlockIdentifier: &types.BlockIdentifier{
			Index: result.Height,
			Hash:  result.Hash,
		},
		ParentBlockIdentifier: &types.BlockIdentifier{
			Index: result.Height,
			Hash:  result.Hash,
		},
		Timestamp:    result.Time,
		Transactions: processedTxs,
		Metadata:     nil,
	}

	return currentBlock, nil
}

func GetTransaction(txHash string) (*types.Transaction, *types.Error) {

	type request struct {
		Hash string `json:"hash"`
	}

	var result map[string]interface{}

	req := request{Hash: txHash}
	if err := NodeClient.CallFor(&result, "transaction_get", req); err != nil {
		return nil, WrapErr(
			ErrNotFound,
			err,
		)
	}

	operations, _ := OperationsFromTx(result)
	// if oErr != nil {
	// 	return nil, oErr
	// }

	transaction := &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: fmt.Sprint(result["hash"]),
		},
		Operations:          operations,
		RelatedTransactions: nil,
		Metadata:            nil,
	}

	return transaction, nil

}

func GetBalance(address string) (*types.Amount, *types.Error) {

	type request struct {
		Address string `json:"address"`
	}

	var result map[string]interface{}

	req := request{Address: address}
	if err := NodeClient.CallFor(&result, "account_get", req); err != nil {
		return nil, WrapErr(
			ErrNotFound,
			err,
		)
	}

	amount := &types.Amount{
		Value:    fmt.Sprint(utils.MapToInt64(result["balance"])),
		Currency: HNT,
	}

	return amount, nil

}

func GetOraclePrice(height int64) (*int64, *types.Error) {
	type request struct {
		Height int64 `json:"height"`
	}

	var result map[string]interface{}

	req := request{Height: height}
	if err := NodeClient.CallFor(&result, "oracle_price_get", req); err != nil {
		return nil, WrapErr(
			ErrNotFound,
			err,
		)
	}

	price := utils.MapToInt64(result["price"])

	return &price, nil
}

func GetImplicitBurn(hash string) (*Fee, *types.Error) {
	type request struct {
		Hash string `json:"hash"`
	}

	var result map[string]interface{}

	req := request{Hash: hash}
	if err := NodeClient.CallFor(&result, "implicit_burn_get", req); err != nil {
		return nil, WrapErr(
			ErrNotFound,
			err,
		)
	}

	feeResult := &Fee{
		Amount:   utils.MapToInt64(result["fee"]),
		Payer:    fmt.Sprint(result["payer"]),
		Currency: HNT,
	}

	return feeResult, nil

}
