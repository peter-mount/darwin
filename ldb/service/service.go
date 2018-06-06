package service

import (
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/golib/kernel/cron"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/nre-feeds/bin"
  "github.com/peter-mount/nre-feeds/darwind3"
  "github.com/peter-mount/nre-feeds/ldb"
)

type LDBService struct {
  ldb           ldb.LDB

  config       *bin.Config
  cron         *cron.CronService
  restService  *rest.Server
}

func (a *LDBService) Name() string {
  return "LDBService"
}

func (a *LDBService) Init( k *kernel.Kernel ) error {
  service, err := k.AddService( &bin.Config{} )
  if err != nil {
    return err
  }
  a.config = (service).(*bin.Config)

  service, err = k.AddService( &cron.CronService{} )
  if err != nil {
    return err
  }
  a.cron = (service).(*cron.CronService)

  service, err = k.AddService( &rest.Server{} )
  if err != nil {
    return err
  }
  a.restService = (service).(*rest.Server)

  // ReferenceUpdate
  return nil
}

func (a *LDBService) PostInit() error {
  a.ldb.Darwin = a.config.Services.DarwinD3
  a.ldb.Reference = a.config.Services.Reference

  // Rest services
  a.restService.Handle( "/boards/{crs}", a.stationHandler ).Methods( "GET" )
  a.restService.Handle( "/service/{rid}", a.serviceHandler ).Methods( "GET" )

  return nil
}

func (a *LDBService) Start() error {

  // Connect to Rabbit & name the connection so its easier to debug
  a.config.RabbitMQ.ConnectionName = "darwin ldb"
  a.config.RabbitMQ.Connect()

  a.ldb.EventManager = darwind3.NewDarwinEventManager( &a.config.RabbitMQ, a.config.D3.EventKeyPrefix )

  if err := a.ldb.Init(); err != nil {
    return err
  }

  // Expire old schedules every 15 minutes
  a.cron.AddFunc( "0 * * * * *", a.ldb.Stations.Cleanup )
  return nil
}