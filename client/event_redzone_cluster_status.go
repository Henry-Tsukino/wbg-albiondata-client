package client

import (
	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

/*
The event is received when cluster red zone status changes.

Example:
> EventDataType: [472]evRedZoneEventClusterStatus - map[0:0 1:0 3:0 252:472]
*/

type eventRedZoneClusterStatus struct {
	Status0 int64 `mapstructure:"0"`
	Status1 int64 `mapstructure:"1"`
	Status3 int64 `mapstructure:"3"`
}

func (event eventRedZoneClusterStatus) Process(state *albionState) {
	log.Debug("Got red zone cluster status event...")

	log.Infof("Red Zone Cluster Status - Status0: %d, Status1: %d, Status3: %d", event.Status0, event.Status1, event.Status3)

	identifier, _ := uuid.NewV4()
	upload := lib.RedZoneEventClusterStatus{
		Status0: event.Status0,
		Status1: event.Status1,
		Status3: event.Status3,
	}
	log.Infof("Sending red zone cluster status event to Discord bot (Identifier: %s)", identifier)
	sendMsgToRedZoneUploader(upload, "redzoneventclusterstatus.ingest", state, identifier.String())
}
