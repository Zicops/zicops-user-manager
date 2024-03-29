package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph"
	"github.com/zicops/zicops-user-manager/graph/generated"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
	"github.com/zicops/zicops-user-manager/lib/jwt"
)

// CCRouter ... the router for the controller
func CCRouter(restRouter *gin.Engine) (*gin.Engine, error) {
	// set up cors
	configCors := cors.DefaultConfig()
	configCors.AllowAllOrigins = true
	configCors.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	restRouter.Use(cors.New(configCors))
	// user a middleware to get context values
	restRouter.Use(func(c *gin.Context) {
		currentRequest := c.Request
		incomingToken := jwt.GetToken(currentRequest)
		claimsFromToken, _ := jwt.GetClaims(incomingToken)
		c.Set("zclaims", claimsFromToken)

	})
	restRouter.GET("/healthz", HealthCheckHandler)
	restRouter.POST("/reset-password", ResetPasswordHandler)
	restRouter.GET("/org", org)
	// create group for restRouter
	version1 := restRouter.Group("/api/v1")
	version1.POST("/query", graphqlHandler())
	return restRouter, nil
}

func org(c *gin.Context) {
	ctx := c.Request.Context()
	d := c.Request.Host
	redisKey := fmt.Sprintf("org:%s", d)
	res, err := redis.GetRedisValue(ctx, redisKey)
	var outputInt userz.Organization
	if err == nil && res != "" {
		json.Unmarshal([]byte(res), &outputInt)
	}
	modelUser, tmp := sendOriginInfo(c.Request.Context(), d, outputInt)
	if modelUser == nil || modelUser.OrgID == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	redisBytes, err := json.Marshal(tmp)
	if err == nil {
		redis.SetRedisValue(ctx, redisKey, string(redisBytes))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": modelUser,
	})
}

func sendOriginInfo(ctx context.Context, domain string, orgDetails userz.Organization) (*model.Organization, *userz.Organization) {
	if domain == "demo.zicops.com" || domain == "https://demo.zicops.com" {
		return nil, nil
	}
	if orgDetails.ID == "" {
		session, err := global.CassPool.GetSession(ctx, "userz")
		if err != nil {
			log.Println("Got error while creating session ", err)
		}
		CassUserSession := session
		queryStr := fmt.Sprintf(`SELECT * from userz.organization where zicops_subdomain='%s' ALLOW FILTERING`, domain)
		getOrgs := func() (orgDomain []userz.Organization, err error) {
			q := CassUserSession.Query(queryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return orgDomain, iter.Select(&orgDomain)
		}
		details, err := getOrgs()
		orgDetails = details[0]
		if err != nil {
			return nil, nil
		}
	}
	eCount, _ := strconv.Atoi(orgDetails.EmpCount)
	storageC := bucket.NewStorageHandler()
	projectID := googleprojectlib.GetGoogleProjectID()
	err := storageC.InitializeStorageClient(ctx, projectID)
	if err != nil {
		log.Error(err)
	}
	// get the logo url
	logoUrl := orgDetails.LogoURL
	if orgDetails.LogoBucket != "" {
		logoUrl = storageC.GetSignedURLForObject(ctx, orgDetails.LogoBucket)
	}
	res := &model.Organization{
		OrgID:         &orgDetails.ID,
		Name:          orgDetails.Name,
		LogoURL:       &logoUrl,
		Industry:      orgDetails.Industry,
		Type:          orgDetails.Type,
		Subdomain:     orgDetails.ZicopsSubdomain,
		EmployeeCount: eCount,
		Website:       orgDetails.Website,
		LinkedinURL:   &orgDetails.Linkedin,
		FacebookURL:   &orgDetails.Facebook,
		TwitterURL:    &orgDetails.Twitter,
		Status:        orgDetails.Status,
		CreatedAt:     orgDetails.CreatedBy,
		UpdatedAt:     orgDetails.UpdatedBy,
		CreatedBy:     &orgDetails.CreatedBy,
		UpdatedBy:     &orgDetails.UpdatedBy,
	}
	return res, &orgDetails
}

type ResetPasswordRequest struct {
	Email string `json:"email"`
}

func ResetPasswordHandler(c *gin.Context) {
	// get the request body
	var resetPasswordRequest ResetPasswordRequest
	if err := c.ShouldBindJSON(&resetPasswordRequest); err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := resetPasswordRequest.Email
	origin := c.Request.Header.Get("Origin")
	ctx := c.Request.Context()
	passwordReset, err := global.IDP.GetResetPasswordURL(ctx, email, origin, origin)
	if err != nil {
		log.Error(err)
		return
	}
	err = global.SGClient.SendPasswordResetEmail(email, passwordReset, "")
	if err != nil {
		log.Error(err)
		return
	}
}

func HealthCheckHandler(c *gin.Context) {
	log.Debugf("HealthCheckHandler Method --> %s", c.Request.Method)

	switch c.Request.Method {
	case http.MethodGet:
		GetHealthStatus(c.Writer)
	default:
		err := errors.New("Method not supported")
		ResponseError(c.Writer, http.StatusBadRequest, err)
	}
}

// GetHealthStatus ...
func GetHealthStatus(w http.ResponseWriter) {
	healthStatus := "Zicops user manager service is healthy"
	response, _ := json.Marshal(healthStatus)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(response); err != nil {
		log.Errorf("GetHealthStatus ... unable to write JSON response: %v", err)
	}
}

// ResponseError ... essentially a single point of sending some error to route back
func ResponseError(w http.ResponseWriter, httpStatusCode int, err error) {
	log.Errorf("Response error %s", err.Error())
	response, _ := json.Marshal(err)
	w.Header().Add("Status", strconv.Itoa(httpStatusCode)+" "+err.Error())
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(httpStatusCode)

	if _, err := w.Write(response); err != nil {
		log.Errorf("ResponseError ... unable to write JSON response: %v", err)
	}
}

func graphqlHandler() gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	return func(c *gin.Context) {
		ctxValue := c.Value("zclaims").(map[string]interface{})
		// set ctxValue to request context
		lspIdInt := c.Request.Header.Get("tenant")
		lspID := ""
		if lspIdInt != "" {
			lspID = lspIdInt
		}
		ctxValue["lsp_id"] = lspID
		// get current origin in https format
		origin := c.Request.Header.Get("Origin")
		mobileHeader := c.Request.Header.Get("zmobile")
		ctxValue["origin"] = origin
		ctxValue["mobile"] = mobileHeader
		ctxValue["role"] = c.Request.Header.Get("role")
		request := c.Request
		requestWithValue := request.WithContext(context.WithValue(request.Context(), "zclaims", ctxValue))
		h.ServeHTTP(c.Writer, requestWithValue)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
