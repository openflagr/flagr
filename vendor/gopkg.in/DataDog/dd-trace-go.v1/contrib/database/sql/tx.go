package sql

import (
	"context"
	"database/sql/driver"
	"time"
)

var _ driver.Tx = (*tracedTx)(nil)

// tracedTx is a traced version of sql.Tx
type tracedTx struct {
	driver.Tx
	*traceParams
	ctx context.Context
}

// Commit sends a span at the end of the transaction
func (t *tracedTx) Commit() (err error) {
	start := time.Now()
	err = t.Tx.Commit()
	t.tryTrace(t.ctx, "Commit", "", start, err)
	return err
}

// Rollback sends a span if the connection is aborted
func (t *tracedTx) Rollback() (err error) {
	start := time.Now()
	err = t.Tx.Rollback()
	t.tryTrace(t.ctx, "Rollback", "", start, err)
	return err
}
