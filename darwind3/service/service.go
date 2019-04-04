package service

import (
	"github.com/peter-mount/golib/kernel"
	"github.com/peter-mount/golib/kernel/cron"
	"github.com/peter-mount/golib/rest"
	"github.com/peter-mount/nre-feeds/bin"
	"github.com/peter-mount/nre-feeds/darwind3"
)

type DarwinD3Service struct {
	darwind3    darwind3.DarwinD3
	config      *bin.Config
	cron        *cron.CronService
	restService *rest.Server
}

func (a *DarwinD3Service) Name() string {
	return "DarwinD3Service"
}

func (a *DarwinD3Service) Init(k *kernel.Kernel) error {
	service, err := k.AddService(&bin.Config{})
	if err != nil {
		return err
	}
	a.config = (service).(*bin.Config)

	service, err = k.AddService(&cron.CronService{})
	if err != nil {
		return err
	}
	a.cron = (service).(*cron.CronService)

	service, err = k.AddService(&rest.Server{})
	if err != nil {
		return err
	}
	a.restService = (service).(*rest.Server)

	// ReferenceUpdate
	return nil
}

func (a *DarwinD3Service) PostInit() error {
	a.darwind3.Config = a.config

	if a.config.D3.ResolveSched {
		// Allow D3 to resolve schedules from the timetable
		a.darwind3.Timetable = a.config.Services.Timetable
	}

	// Rest services
	a.restService.Handle("/message/broadcast", a.BroadcastStationMessagesHandler).Methods("POST")
	a.restService.Handle("/message/{id}", a.StationMessageHandler).Methods("GET")
	a.restService.Handle("/messages", a.AllMessageHandler).Methods("GET")
	a.restService.Handle("/messages/{crs}", a.CrsMessageHandler).Methods("GET")

	a.restService.Handle("/schedule/{rid}", a.ScheduleHandler).Methods("GET")

	a.restService.Handle("/status", a.StatusHandler).Methods("GET")
	return nil
}

func (a *DarwinD3Service) Start() error {

	// Connect to Rabbit & name the connection so its easier to debug
	a.config.RabbitMQ.ConnectionName = "darwin d3"
	err := a.config.RabbitMQ.Connect()
	if err != nil {
		return err
	}

	em := darwind3.NewDarwinEventManager(&a.config.RabbitMQ, a.config.D3.EventKeyPrefix)

	a.config.DbPath(&a.config.Database.PushPort, "dwd3.db")
	if err := a.darwind3.OpenDB(a.config.Database.PushPort, em); err != nil {
		return err
	}

	// Expire old station messages every 15 minutes
	_, err = a.cron.AddFunc("0 0/15 * * * *", a.darwind3.ExpireStationMessages)
	if err != nil {
		return err
	}

	// Purge old schedules every hour
	_, err = a.cron.AddFunc("0 0 * * * *", a.darwind3.PurgeSchedules)
	if err != nil {
		return err
	}

	// Check for any orphans once every 6 hours
	_, err = a.cron.AddFunc("0 0 0/6 * * *", a.darwind3.PurgeOrphans)
	if err != nil {
		return err
	}

	// The V16 PushPort queue
	err = a.darwind3.BindConsumer(&a.config.RabbitMQ, a.config.D3.PushPort.QueueName, a.config.D3.PushPort.RoutingKey)
	if err != nil {
		return err
	}

	return nil
}
