package global

import (
	"context"
	"sync"

	"github.com/scylladb/gocqlx/v2"
	cry "github.com/zicops/zicops-user-manager/lib/crypto"
	"github.com/zicops/zicops-user-manager/lib/identity"
	"github.com/zicops/zicops-user-manager/lib/sendgrid"
)

// some global variables commonly used
var (
	CTX             context.Context
	CassUserSession *gocqlx.Session
	CryptSession    *cry.Cryptography
	Cancel          context.CancelFunc
	WaitGroupServer sync.WaitGroup
	IDP             *identity.IDP
	SGClient        *sendgrid.ClientSendGrid
)

// initializes global package to read environment variables as needed
func init() {
}
