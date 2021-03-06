package darwind3

// Process processes an inbound Train Status update, merging it with an existing
// schedule in the database
func (p *TS) Process(tx *Transaction) error {

	// Retrieve the schedule to be updated
	sched := tx.d3.GetSchedule(p.RID)

	// Still no schedule then We've got a TS for a train with no known schedule so create one
	if sched == nil {
		sched = &Schedule{
			RID: p.RID,
			UID: p.UID,
			SSD: p.SSD,
		}
		sched.Defaults()
	}

	// If the TS is older than what's in the schedule then then do nothing as it's
	// presumably old data that's sent out of sync or it's from a snapshot
	if tx.pport.TS.Before(sched.Date) {
		return nil
	}

	// SnapshotUpdate the LateReason
	sched.LateReason = p.LateReason

	sortRequired := false

	// Merge or append the inbound locations
	for _, a := range p.Locations {
		// set forecast date of the new entries
		a.Forecast.Date = tx.pport.TS

		b := LocationSliceFind(sched.Locations, a)
		if b == nil {
			// New inbound location
			sched.Locations = append(sched.Locations, a)
			sortRequired = true
		} else {
			// Merge the new data with the old
			b.MergeFrom(a)
		}
	}

	// Sort if required else just update the times
	if sortRequired {
		sched.Sort()
	} else {
		sched.UpdateTime()
	}

	// Update schedule time
	sched.Date = tx.pport.TS

	tx.d3.UpdateAssociations(sched)

	tx.d3.PutSchedule(sched)

	tx.d3.EventManager.PostEvent(&DarwinEvent{
		Type:     Event_ScheduleUpdated,
		RID:      sched.RID,
		Schedule: sched,
	})

	return nil
}
