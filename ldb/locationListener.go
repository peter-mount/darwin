package ldb

import (
  "darwind3"
)

// locationListener listens to location updates and updates the relevant
// Station with the new/updated entry
func (d *LDB) locationListener( e *darwind3.DarwinEvent ) {
  e.Schedule.View( func() error {
    // Ignore anything without a location & no public times
    for _, l := range e.Schedule.Locations {
      if l.Times.IsPublic() {

        // Retrieve the station, it should be a valid one if we have public times
        station := d.GetStationTiploc( l.Tiploc )
        if station != nil {
          station.addService( e, l.Clone() )
        }
      }
    }

    return nil
  })
}
