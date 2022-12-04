package main

import (
	"context"
	"fmt"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-user-manager/controller"
	"github.com/zicops/zicops-user-manager/global"
	cry "github.com/zicops/zicops-user-manager/lib/crypto"
	"github.com/zicops/zicops-user-manager/lib/identity"
	"github.com/zicops/zicops-user-manager/lib/sendgrid"
)

const defaultPort = "8080"

func main() {
	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "zicops-cc.json")
	log.Infof("Starting zicops user manager service")
	ctx, cancel := context.WithCancel(context.Background())
	crySession := cry.New("09afa9f9544a7ff1ae9988f73ba42134")
	idp, err := identity.NewIDPEP(ctx, "zicops-one")
	if err != nil {
		log.Errorf("Error connecting to identity: %s", err)
		log.Infof("zicops user manager initialization failed")
	}
	global.IDP = idp
	sgClient := sendgrid.NewSendGridClient()
	err = sgClient.InitializeSendGridClient()
	if err != nil {
		log.Errorf("Error connecting to sendgrid: %s", err)
		log.Infof("zicops user manager initialization failed")
	}
	global.SGClient = sgClient
	global.CTX = ctx
	global.Cancel = cancel
	global.CryptSession = &crySession
	log.Infof("zicops course query initialization complete")
	portFromEnv := os.Getenv("PORT")
	port, err := strconv.Atoi(portFromEnv)

	if err != nil {
		port = 8094
	}
	gin.SetMode(gin.ReleaseMode)
	bootUPErrors := make(chan error, 1)
	go monitorSystem(cancel, bootUPErrors)
	go checkAndInitCassandraSession()
	controller.CCBackendController(ctx, port, bootUPErrors)
	err = <-bootUPErrors
	if err != nil {
		log.Errorf("There is an issue starting backend server for course query: %v", err.Error())
		global.WaitGroupServer.Wait()
		os.Exit(1)
	}
	log.Infof("course query server started successfully.")
}

func monitorSystem(cancel context.CancelFunc, errorChannel chan error) {
	holdSignal := make(chan os.Signal, 1)
	signal.Notify(holdSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	// if system throw any termination stuff let channel handle it and cancel
	<-holdSignal
	cancel()
	// send error to channel
	errorChannel <- fmt.Errorf("System termination signal received")
}

func checkAndInitCassandraSession() error {
	// get user session every 1 minute
	// if session is nil then create new session
	//test cassandra connection
	_, err1 := cassandra.GetCassSession("coursez")
	_, err2 := cassandra.GetCassSession("qbankz")
	_, err3 := cassandra.GetCassSession("userz")
	if err1 != nil || err2 != nil || err3 != nil {
		log.Errorf("Error connecting to cassandra: %v and %v ", err1, err2, err3)
	} else {
		log.Infof("Cassandra connection successful")
	}
	_, err := redis.Initialize()
	if err != nil {
		log.Errorf("Error connecting to redis: %v", err)
	} else {
		log.Infof("Redis connection successful")
	}
	for {
		for key := range cassandra.GlobalSession {
			session, err := cassandra.GetCassSession(key)
			restart := false
			if session != nil {
				qX := session.Query("select now() from system.local", nil)
				if qX != nil {
					err = qX.Exec()
					if err != nil {
						restart = true
					}
				}
			} else {
				restart = true
			}

			if err != nil || restart {
				cassandra.GlobalSession[key] = nil
				session, err := cassandra.GetCassSession(key)
				if err == nil && session != nil {
					cassandra.GlobalSession[key] = nil
				}
			}
		}
		time.Sleep(10 * time.Minute)
	}
}
