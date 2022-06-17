package global

import (
	"context"
	"sync"

	cry "github.com/zicops/zicops-user-manager/lib/crypto"
	"github.com/zicops/zicops-user-manager/lib/db/cassandra"
	"github.com/zicops/zicops-user-manager/lib/identity"
	"github.com/zicops/zicops-user-manager/lib/sendgrid"
)

// some global variables commonly used
var (
	CTX             context.Context
	CassUserSession *cassandra.Cassandra
	CryptSession    *cry.Cryptography
	Cancel          context.CancelFunc
	WaitGroupServer sync.WaitGroup
	IDP             *identity.IDP
	SGClient        *sendgrid.ClientSendGrid
)

// initializes global package to read environment variables as needed
func init() {
}
