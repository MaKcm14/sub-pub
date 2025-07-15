package app

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MaKcm14/vk-test/internal/config"
	"github.com/MaKcm14/vk-test/internal/controller/spserv"
	"github.com/MaKcm14/vk-test/pkg/subpub"
)

// Service defines the main sub-pub service's builder.
type Service struct {
	log     *slog.Logger
	logFile *os.File

	serv spserv.SubPubService
}

func NewService() Service {
	const op = "app.NewService"

	logFile, err := os.Create("../../logs/main_log_file.txt")

	if err != nil {
		panic(fmt.Sprintf("error of the %s: %s", op, err))
	}
	log := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo}))

	log.Info("configuring the service begun")

	conf, err := config.New(
		config.ConfigSocket,
	)
	if err != nil {
		critErr := fmt.Errorf("error of the %s: %s", op, err)
		log.Error(critErr.Error())
		panic(critErr)
	}

	log.Info("configuring the sub-pub service started")

	service, err := spserv.NewSubPubService(log, conf.Socket,
		spserv.NewSubPubServer(log, subpub.NewSubPub()),
	)

	if err != nil {
		critErr := fmt.Errorf("error of the %s: %s", op, err)
		log.Error(critErr.Error())
		panic(critErr)
	}

	return Service{
		log:     log,
		logFile: logFile,
		serv:    service,
	}
}

// Start starts the sub-pub service.
func (s *Service) Start() {
	defer s.close()

	s.log.Info("starting the service")
	go s.serv.Run()

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig
}

// сlose calls the close funcs for releasing the resources.
func (s *Service) close() {
	s.serv.Close()
	s.log.Info("the service was FULLY STOPPED")
	s.logFile.Close()
}
