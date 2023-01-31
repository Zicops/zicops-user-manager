package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/zicops/contracts/vendorz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

func CreateVendor(ctx context.Context, input *model.VendorInput) (string, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return "", nil
	}
	lspID := claims["lsp_id"].(string)
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return "", err
	}
	CassUserSession := session
	email := claims["email"].(string)
	userId := base64.URLEncoding.EncodeToString([]byte(email))

	var photoUrl string
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return "", err
	}

	var userEmail []string
	for _, vv := range input.Users {
		v := vv
		userEmail = append(userEmail, *v)
	}
	_, err = InviteUsers(ctx, userEmail, lspID)
	if err != nil {
		return "", err
	}

	createdAt := time.Now().Unix()
	id, _ := uuid.NewUUID()
	vendorId := id.String()

	var photoBucket string
	if input.Photo != nil {
		//upload the photo to Google Bucket
		bucketPath := fmt.Sprintf("%s/%s/%s", "vendor", userId, input.Photo.Filename)
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return "", err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Photo.File); err != nil {
			return "", err
		}
		currentBytes := fileBuffer.Bytes()
		_, err = io.Copy(writer, bytes.NewReader(currentBytes))
		if err != nil {
			return "", err
		}
		photoBucket = bucketPath
		photoUrl = storageC.GetSignedURLForObject(bucketPath)
	} else {
		photoBucket = ""
		if input.PhotoURL != nil {
			photoUrl = *input.PhotoURL
		}
	}
	vendor := vendorz.Vendor{
		VendorId:    vendorId,
		Type:        input.Type,
		Level:       input.Level,
		Name:        input.Name,
		Address:     input.Address,
		PhotoUrl:    photoUrl,
		PhotoBucket: photoBucket,
		Status:      input.Status,
		CreatedAt:   createdAt,
		CreatedBy:   email,
		Users:       userEmail,
	}

	if input.Website != nil {
		vendor.Website = *input.Website
	}
	if input.Facebook != nil {
		vendor.Facebook = *input.Facebook
	}
	if input.Instagram != nil {
		vendor.Instagram = *input.Instagram
	}
	if input.LinkedIn != nil {
		vendor.LinkedIn = *input.LinkedIn
	}
	if input.Twitter != nil {
		vendor.Twitter = *input.Twitter
	}

	insertQuery := CassUserSession.Query(vendorz.VendorTable.Insert()).BindStruct(vendor)
	if err := insertQuery.ExecRelease(); err != nil {
		return "", err
	}

	return vendorId, nil
}

func CreateProfileVendor(ctx context.Context, input *model.VendorProfile) (string, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Println("Got error while getting claims: %v", err)
		return "", err
	}
	lspID := claims["lsp_id"].(string)

	return "", nil
}

func CreateExperienceVendor(ctx context.Context, input model.ExperienceInput) (string, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Println("Got error while getting claims: %v", err)
		return "", err
	}
	lspID := claims["lsp_id"].(string)

	return "", nil
}
