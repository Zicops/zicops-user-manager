package main

import (
	"context"
	"fmt"
	"os/signal"
	"strconv"
	"syscall"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-user-manager/controller"
	"github.com/zicops/zicops-user-manager/global"
	cry "github.com/zicops/zicops-user-manager/lib/crypto"
	"github.com/zicops/zicops-user-manager/lib/identity"
	"github.com/zicops/zicops-user-manager/lib/sendgrid"
)

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
	_, err = redis.Initialize()
	if err != nil {
		log.Errorf("Error connecting to redis: %v", err)
	} else {
		log.Infof("Redis connection successful")
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
	// get global Monitor object
	m := ginmetrics.GetMonitor()

	// +optional set metric path, default /debug/metrics
	m.SetMetricPath("/metrics")
	// +optional set slow time, default 5s
	m.SetSlowTime(10)
	// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
	// used to p95, p99
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})

	// set middleware for gin
	r := gin.Default()
	m.Use(r)
	gin.SetMode(gin.ReleaseMode)
	bootUPErrors := make(chan error, 1)
	go monitorSystem(cancel, bootUPErrors)
	go checkAndInitCassandraSession()
	controller.CCBackendController(ctx, port, bootUPErrors, r)
	err = <-bootUPErrors
	if err != nil {
		log.Errorf("there is an issue starting backend server for course query: %v", err.Error())
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
	errorChannel <- fmt.Errorf("system termination signal received")
}

func checkAndInitCassandraSession() {
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
}
