package main

import (
	"context"
	"fmt"
	"os/signal"
	"strconv"
	"syscall"

	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-user-manager/config"
	"github.com/zicops/zicops-user-manager/controller"
	"github.com/zicops/zicops-user-manager/global"
	cry "github.com/zicops/zicops-user-manager/lib/crypto"
	"github.com/zicops/zicops-user-manager/lib/db/cassandra"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

const defaultPort = "8080"

func main() {
	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "zicops-cc.json")
	log.Infof("Starting zicops user manager service")
	ctx, cancel := context.WithCancel(context.Background())
	crySession := cry.New("09afa9f9544a7ff1ae9988f73ba42134")
	cassConfig := config.NewCassandraConfig()
	cassSession, err := cassandra.New(cassConfig)
	if err != nil {
		log.Errorf("Error connecting to cassandra: %s", err)
		log.Infof("zicops user manager intialization failed")
	}
	idp, err := identity.NewIDPEP(ctx, "zicops-one")
	if err != nil {
		log.Errorf("Error connecting to identity: %s", err)
		log.Infof("zicops user manager initialization failed")
	}
	global.IDP = idp
	global.CTX = ctx
	global.CassSession = cassSession
	global.Cancel = cancel
	global.CryptSession = &crySession
	log.Infof("zicops course query initialization complete")
	portFromEnv := os.Getenv("PORT")
	port, err := strconv.Atoi(portFromEnv)

	if err != nil {
		port = 8094
	}
	bootUPErrors := make(chan error, 1)
	go monitorSystem(cancel, bootUPErrors)
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
