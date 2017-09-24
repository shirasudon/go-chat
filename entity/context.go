package entity

import (
	"context"
)

const txObjectKey = "_TRANSACTION_"

// Get Tx object from context.
func GetTx(ctx context.Context) (tx Tx, exists bool) {
	tx, exists = ctx.Value(txObjectKey).(Tx)
	return
}

// set transaction to the new clild context and return it.
func WithTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txObjectKey, tx)
}

// Tx is a interface for the transaction context.
// the transaction must end by calling Commit() or Rollback().
//
// Because the transaction object depends on the external
// infrastructures such as Sql-like DB,
// it can be used with type assertion to use full features
// for the implementation-specfic transaction object.
type Tx interface {
	Commit() error
	Rollback() error
}
