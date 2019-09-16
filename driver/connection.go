package driver

import (
	"context"
	"database/sql/driver"
	"net"
	"time"
)

type driverConn struct {
	netConn          net.Conn
	rawConn          net.Conn // underlying connection when netConn is TLS connection.
	affectedRows     uint64
	insertId         uint64
	maxAllowedPacket int
	maxWriteSize     int
	writeTimeout     time.Duration
	sequence         uint8
	parseTime        bool
	reset            bool // set when the Go SQL package calls ResetSession

	// for context support (Go 1.8+)
	watching bool
	watcher  chan<- context.Context
	closech  chan struct{}
	finished chan<- struct{}
}

func (mc *driverConn) Begin() (driver.Tx, error) {
	return mc.begin(false)
}

func (mc *driverConn) begin(readOnly bool) (driver.Tx, error) {
	return nil, nil
}

func (mc *driverConn) Close() (err error) {
	return
}

func (mc *driverConn) Prepare(query string) (driver.Stmt, error) {

	return nil, nil
}

func (mc *driverConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (mc *driverConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return nil, nil
}

func (mc *driverConn) query(query string, args []driver.Value) (*textRows, error) {
	return nil, nil
}

// Ping implements driver.Pinger interface
func (mc *driverConn) Ping(ctx context.Context) (err error) {
	return nil
}

// BeginTx implements driver.ConnBeginTx interface
func (mc *driverConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return nil, nil
}

func (mc *driverConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	rows := new(textRows)
	rows.args = args
	return rows, nil
}

func (mc *driverConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (mc *driverConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return nil, nil
}
