package handlers

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/helpers"
)

func GetUserFromCass(ctx context.Context) (*userz.User, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	users := []userz.User{}

	getQueryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' `, emailCreatorID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return &users[0], nil
}
