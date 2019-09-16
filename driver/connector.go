package driver

import (
	"context"
	"database/sql/driver"
)

type connector struct {
}

// Connect implements driver.Connector interface.
// Connect returns a connection to the database.
func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	mc := &driverConn{
		maxAllowedPacket: 100,
		maxWriteSize:     100,
		closech:          make(chan struct{}),
	}
	return mc, nil
}

// Driver implements driver.Connector interface.
// Driver returns &MySQLDriver{}.
func (c *connector) Driver() driver.Driver {
	return &CacheDriver{}
}
