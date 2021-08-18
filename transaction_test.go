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
			name:         "Txid overwritten with new value",
			ctx:          context.WithValue(context.Background(), transactionIDContextKey, "1234"),
			txid:         "abcd",
			expectedTxid: "abcd",
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

func TestOverwriteTransactionID(t *testing.T) {

	ctx := context.Background()
	txid1 := "abcd"
	txid2 := "1234"

	buf := &bytes.Buffer{}
	logger := New(Config{Output: buf})

	ctx = SetTransactionID(ctx, &logger, txid1)
	assert.Equal(t, txid1, GetTransactionID(ctx))
	logger.Info("Some log statement")
	assert.Contains(t, buf.String(), txid1, "Log statement: "+buf.String())
	assert.NotContains(t, buf.String(), txid2, "Log statement: "+buf.String())

	buf.Reset()

	ctx = SetTransactionID(ctx, &logger, txid2)
	assert.Equal(t, txid2, GetTransactionID(ctx))
	logger.Info("Some log statement")
	assert.Contains(t, buf.String(), txid2, "Log statement: "+buf.String())
	assert.NotContains(t, buf.String(), txid1, "Log statement: "+buf.String())
}
