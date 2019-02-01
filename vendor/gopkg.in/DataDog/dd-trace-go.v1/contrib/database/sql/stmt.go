package sql

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"
)

var _ driver.Stmt = (*tracedStmt)(nil)

// tracedStmt is traced version of sql.Stmt
type tracedStmt struct {
	driver.Stmt
	*traceParams
	ctx   context.Context
	query string
}

// Close sends a span before closing a statement
func (s *tracedStmt) Close() (err error) {
	start := time.Now()
	err = s.Stmt.Close()
	s.tryTrace(s.ctx, "Close", "", start, err)
	return err
}

// ExecContext is needed to implement the driver.StmtExecContext interface
func (s *tracedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	start := time.Now()
	if stmtExecContext, ok := s.Stmt.(driver.StmtExecContext); ok {
		res, err := stmtExecContext.ExecContext(ctx, args)
		s.tryTrace(ctx, "Exec", s.query, start, err)
		return res, err
	}
	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	res, err = s.Exec(dargs)
	s.tryTrace(ctx, "Exec", s.query, start, err)
	return res, err
}

// QueryContext is needed to implement the driver.StmtQueryContext interface
func (s *tracedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	start := time.Now()
	if stmtQueryContext, ok := s.Stmt.(driver.StmtQueryContext); ok {
		rows, err := stmtQueryContext.QueryContext(ctx, args)
		s.tryTrace(ctx, "Query", s.query, start, err)
		return rows, err
	}
	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	rows, err = s.Query(dargs)
	s.tryTrace(ctx, "Query", s.query, start, err)
	return rows, err
}

// copied from stdlib database/sql package: src/database/sql/ctxutil.go
func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}
