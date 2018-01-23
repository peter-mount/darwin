// CIF Rest server
package main

import (
  "darwinref"
  "darwintimetable"
  "flag"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/golib/statistics"
  "log"
  "os"
  "os/signal"
  "syscall"
)

func main() {
  log.Println( "darwin v0.1" )

  refFile := flag.String( "ref", "", "The reference database file" )
  ttFile := flag.String( "timetable", "", "The timetable database file" )

  // Port for the webserver
  port := flag.Int( "p", 8080, "Port to use" )

  flag.Parse()

  stats := statistics.Statistics{ Log: true }
  stats.Configure()

  server := rest.NewServer( *port )

  var ref *darwinref.DarwinReference
  if *refFile != "" {
    ref = &darwinref.DarwinReference{}

    if err := ref.OpenDB( *refFile ); err != nil {
      log.Fatal( err )
    }

    ref.RegisterRest( server.Context( "/ref" ) )
  }

  var tt *darwintimetable.DarwinTimetable
  if *ttFile != "" {
    tt = &darwintimetable.DarwinTimetable{}

    if err := tt.OpenDB( *ttFile ); err != nil {
      log.Fatal( err )
    }

    //tt.RegisterRest( server.Context( "/timetable" ) )
  }

  // Listen to signals & close the db before exiting
  // SIGINT for ^C, SIGTERM for docker stopping the container
  sigs := make( chan os.Signal, 1 )
  signal.Notify( sigs, syscall.SIGINT, syscall.SIGTERM )
  go func() {
    sig := <-sigs
    log.Println( "Signal", sig )

    if( ref != nil ) {
      ref.Close()
    }

    log.Println( "Database closed" )

    os.Exit( 0 )
  }()

  server.Start()
}
