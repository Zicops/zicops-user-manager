package global

import (
	"context"
	"sync"

	cry "github.com/zicops/zicops-user-manager/lib/crypto"
	"github.com/zicops/zicops-user-manager/lib/db/cassandra"
)

// some global variables commonly used
var (
	CTX             context.Context
	CassSession     *cassandra.Cassandra
	CryptSession    *cry.Cryptography
	Cancel          context.CancelFunc
	WaitGroupServer sync.WaitGroup
)

// initializes global package to read environment variables as needed
func init() {
}
