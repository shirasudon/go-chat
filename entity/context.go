package entity

import (
	"context"
)

// TxContext is the extension for context.Context,
// which holds current transaction state.
type TxContext interface {
	context.Context

	// Tx returns the transaction object which executes consistent operations
	// to handle some entity object.
	//
	// Because the transaction object depends on the external
	// infrastructures such as Sql-like DB,
	// it returns interface{} value which is used with type assertion.
	//
	// If the transaction is not started, Tx() should return nil.
	Tx() interface{}
}
