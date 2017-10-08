package domain

import (
	"context"
	"database/sql"
)

const txObjectKey = "_TRANSACTION_"

// Get Tx object from context.
func GetTx(ctx context.Context) (tx Tx, exists bool) {
	tx, exists = ctx.Value(txObjectKey).(Tx)
	if exists && tx != nil {
		return tx, exists
	}
	return nil, false
}

// set transaction to the new clild context and return it.
func SetTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txObjectKey, tx)
}

// Tx is a interface for the transaction context.
// the transaction must end by calling Commit() or Rollback().
//
// Because the transaction object depends on the external
// infrastructures such as Sql-like DB,
// it can be used with type assertion to use full features
// for the implementation-specific transaction object.
type Tx interface {
	Commit() error
	Rollback() error
}

// TxBeginner can start transaction with context object.
// TxOptions are typically used to specify
// the transaction level.
// A nil TxOptions means to use default transaction level.
type TxBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (Tx, error)
}

// EmptyTxBeginner implements Tx and TxBeginner interfaces.
// It is used to implement the repository with the no-operating
// transaction, typically used as embedded struct.
type EmptyTxBeginner struct{}

func (EmptyTxBeginner) Commit() error                                          { return nil }
func (EmptyTxBeginner) Rollback() error                                        { return nil }
func (tx EmptyTxBeginner) BeginTx(context.Context, *sql.TxOptions) (Tx, error) { return tx, nil }
