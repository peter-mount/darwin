// ldb Microservice
package main

import (
  "bin"
  "darwind3"
  "ldb"
  "log"
)

func main() {
  log.Println( "ldb v0.1" )

  bin.RunApplication( app )
}

func app( config *bin.Config ) ( func(), error ) {

  // Connect to Rabbit & name the connection so its easier to debug
  config.RabbitMQ.ConnectionName = "darwin ldb"
  config.RabbitMQ.Connect()

  db := &ldb.LDB{
    Darwin: config.Services.DarwinD3,
    Reference: config.Services.Reference,
  }

  db.EventManager = darwind3.NewDarwinEventManager( &config.RabbitMQ )

  if err := db.Init(); err != nil {
    return nil, err
  }

  db.RegisterRest( config.Server.Ctx )

  // Expire old schedules every 15 minutes
  config.Cron.AddFunc( "0 * * * * *", db.Stations.Cleanup )

  return nil, nil
}
