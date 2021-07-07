package logevent

import (
	"context"

	"github.com/google/uuid"
)

const (
	// TransactionIDKey is the key name being set in the logger
	TransactionIDKey = "transaction_id"

	transactionIDContextKey = ctxKey("__logevent_transaction_id")
)

// SetTransactionID sets a transaction id string in the logger and context.
// If an empty string is passed in, then a randomly generated uuid will be used as the transaction id.
func SetTransactionID(ctx context.Context, logger *Logger, transactionID string) context.Context {

	if transactionID == "" {

		var ok bool
		transactionID, ok = ctx.Value(transactionIDContextKey).(string)
		if !ok {
			transactionID = uuid.New().String()
		}

	}

	(*logger).SetField(TransactionIDKey, transactionID)

	return context.WithValue(ctx, transactionIDContextKey, transactionID)
}

// GetTransactionID retrieves the transaction id after `SetTransactionID` has been called.
// It will return an empty string if no transaction id has been set.
func GetTransactionID(ctx context.Context) string {
	transactionID, _ := ctx.Value(transactionIDContextKey).(string)
	return transactionID
}
