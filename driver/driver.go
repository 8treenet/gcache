package driver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net"
	"sync"
)

// CacheDriver is exported to make the driver directly accessible.
// In general the driver is used via the database/sql package.
type CacheDriver struct{}

// DialFunc is a function which can be used to establish the network connection.
// Custom dial functions must be registered with RegisterDial
//
// Deprecated: users should register a DialContextFunc instead
type DialFunc func(addr string) (net.Conn, error)

// DialContextFunc is a function which can be used to establish the network connection.
// Custom dial functions must be registered with RegisterDialContext
type DialContextFunc func(ctx context.Context, addr string) (net.Conn, error)

var (
	dialsLock sync.RWMutex
	dials     map[string]DialContextFunc
)

// RegisterDialContext registers a custom dial function. It can then be used by the
// network address mynet(addr), where mynet is the registered new network.
// The current context for the connection and its address is passed to the dial function.
func RegisterDialContext(net string, dial DialContextFunc) {
	dialsLock.Lock()
	defer dialsLock.Unlock()
	if dials == nil {
		dials = make(map[string]DialContextFunc)
	}
	dials[net] = dial
}

// RegisterDial registers a custom dial function. It can then be used by the
// network address mynet(addr), where mynet is the registered new network.
// addr is passed as a parameter to the dial function.
//
// Deprecated: users should call RegisterDialContext instead
func RegisterDial(network string, dial DialFunc) {
	RegisterDialContext(network, func(_ context.Context, addr string) (net.Conn, error) {
		return dial(addr)
	})
}

func (d CacheDriver) Open(dsn string) (driver.Conn, error) {
	c := &connector{}
	return c.Connect(context.Background())
}

func init() {
	sql.Register("cache_plugin", &CacheDriver{})
}
