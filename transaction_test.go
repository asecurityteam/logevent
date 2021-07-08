package logevent

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndGetTransactionID(t *testing.T) {

	var cases = []struct {
		name         string
		mocks        func(ctx context.Context)
		ctx          context.Context
		logger       Logger
		txid         string
		expectedTxid string
	}{
		{
			name:         "Empty txid",
			ctx:          context.Background(),
			txid:         "",
			expectedTxid: "",
		},
		{
			name:         "Txid in context",
			ctx:          context.WithValue(context.Background(), transactionIDContextKey, "1234"),
			txid:         "",
			expectedTxid: "1234",
		},
		{
			name:         "Txid passed in",
			ctx:          context.Background(),
			txid:         "abcd",
			expectedTxid: "abcd",
		},
	}

	for _, currentCase := range cases {
		t.Run(currentCase.name, func(tb *testing.T) {

			buf := &bytes.Buffer{}
			currentCase.logger = New(Config{Output: buf})

			ctx := SetTransactionID(currentCase.ctx, &currentCase.logger, currentCase.txid)
			txid := GetTransactionID(ctx)

			if currentCase.expectedTxid == "" {
				assert.NotEmpty(t, txid)
			} else {
				assert.Equal(t, currentCase.expectedTxid, txid)
			}

			currentCase.logger.Info("Some log statement")
			assert.Contains(t, buf.String(), txid)
		})
	}
}
