package global

import (
	"context"
	"sync"

	"github.com/zicops/zicops-cass-pool/cassandra"
	cry "github.com/zicops/zicops-user-manager/lib/crypto"
	"github.com/zicops/zicops-user-manager/lib/identity"
	"github.com/zicops/zicops-user-manager/lib/sendgrid"
)

// some global variables commonly used
var (
	CTX             context.Context
	CryptSession    *cry.Cryptography
	Cancel          context.CancelFunc
	WaitGroupServer sync.WaitGroup
	IDP             *identity.IDP
	SGClient        *sendgrid.ClientSendGrid
	CassPool        *cassandra.CassandraPool
)

// initializes global package to read environment variables as needed
func init() {
}
