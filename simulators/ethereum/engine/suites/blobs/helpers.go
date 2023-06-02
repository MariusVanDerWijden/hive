package suite_blobs

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/hive/simulators/ethereum/engine/client"
	e_typ "github.com/ethereum/hive/simulators/ethereum/engine/types"
)

type TestBlobTxPool struct {
	Transactions map[common.Hash]e_typ.Transaction
}

func (pool *TestBlobTxPool) AddBlobTransaction(tx e_typ.Transaction) {
	if pool.Transactions == nil {
		pool.Transactions = make(map[common.Hash]e_typ.Transaction)
	}
	pool.Transactions[tx.Hash()] = tx
}

// Test two different transactions with the same blob, and check the blob bundle.

func VerifyTransactionFromNode(ctx context.Context, eth client.Eth, tx e_typ.Transaction) error {
	returnedTx, _, err := eth.TransactionByHash(ctx, tx.Hash())
	if err != nil {
		return err
	}

	// Verify that the tx fields are all the same
	if returnedTx.Nonce() != tx.Nonce() {
		return fmt.Errorf("nonce mismatch: %d != %d", returnedTx.Nonce(), tx.Nonce())
	}
	if returnedTx.Gas() != tx.Gas() {
		return fmt.Errorf("gas mismatch: %d != %d", returnedTx.Gas(), tx.Gas())
	}
	if returnedTx.GasPrice().Cmp(tx.GasPrice()) != 0 {
		return fmt.Errorf("gas price mismatch: %d != %d", returnedTx.GasPrice(), tx.GasPrice())
	}
	if returnedTx.Value().Cmp(tx.Value()) != 0 {
		return fmt.Errorf("value mismatch: %d != %d", returnedTx.Value(), tx.Value())
	}
	if returnedTx.To() != nil && tx.To() != nil && returnedTx.To().Hex() != tx.To().Hex() {
		return fmt.Errorf("to mismatch: %s != %s", returnedTx.To().Hex(), tx.To().Hex())
	}
	if returnedTx.Data() != nil && tx.Data() != nil && !bytes.Equal(returnedTx.Data(), tx.Data()) {
		return fmt.Errorf("data mismatch: %s != %s", hex.EncodeToString(returnedTx.Data()), hex.EncodeToString(tx.Data()))
	}
	if returnedTx.AccessList() != nil && tx.AccessList() != nil && !reflect.DeepEqual(returnedTx.AccessList(), tx.AccessList()) {
		return fmt.Errorf("access list mismatch: %v != %v", returnedTx.AccessList(), tx.AccessList())
	}
	if returnedTx.ChainId().Cmp(tx.ChainId()) != 0 {
		return fmt.Errorf("chain id mismatch: %d != %d", returnedTx.ChainId(), tx.ChainId())
	}
	if returnedTx.BlobGas() != tx.BlobGas() {
		return fmt.Errorf("data gas mismatch: %d != %d", returnedTx.BlobGas(), tx.BlobGas())
	}
	if returnedTx.GasFeeCap().Cmp(tx.GasFeeCap()) != 0 {
		return fmt.Errorf("max fee per gas mismatch: %d != %d", returnedTx.GasFeeCap(), tx.GasFeeCap())
	}
	if returnedTx.GasTipCap().Cmp(tx.GasTipCap()) != 0 {
		return fmt.Errorf("max priority fee per gas mismatch: %d != %d", returnedTx.GasTipCap(), tx.GasTipCap())
	}
	if returnedTx.BlobGasFeeCap().Cmp(tx.BlobGasFeeCap()) != 0 {
		return fmt.Errorf("max fee per data gas mismatch: %d != %d", returnedTx.BlobGasFeeCap(), tx.BlobGasFeeCap())
	}
	if returnedTx.BlobHashes() != nil && tx.BlobHashes() != nil && !reflect.DeepEqual(returnedTx.BlobHashes(), tx.BlobHashes()) {
		return fmt.Errorf("blob versioned hashes mismatch: %v != %v", returnedTx.BlobHashes(), tx.BlobHashes())
	}
	if returnedTx.Type() != tx.Type() {
		return fmt.Errorf("type mismatch: %d != %d", returnedTx.Type(), tx.Type())
	}

	return nil
}
