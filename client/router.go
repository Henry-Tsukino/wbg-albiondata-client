package client

import (
	"encoding/gob"
	"os"

	"github.com/ao-data/albiondata-client/client/photon"
	"github.com/ao-data/albiondata-client/log"
)

// Router struct definitions
type Router struct {
	albionstate     *albionState
	newOperation    chan operation
	recordRawPacket chan photon.RawPacket
	quit            chan bool
	buildingAgg     *buildingAggregator
}

func newRouter() *Router {
	r := &Router{
		albionstate: &albionState{LocationId: "",
			LocationHistory: NewLocationBuffer()},
		newOperation:    make(chan operation, 1000),
		recordRawPacket: make(chan photon.RawPacket, 1000),
		quit:            make(chan bool, 1),
	}

	// Собираем evNewBuilding/evCraftBuildingInfo/evAccessStatus в единую
	// сущность и, как только она готова, отправляем её боту через общий
	// ingest-путь (тот же uploader interface, что и остальной клиент).
	var agg *buildingAggregator
	agg = newBuildingAggregator(func(b *BuildingEntity) {
		agg.sendBuildingUpdate(r.albionstate, b)
	}, newHTTPUploader(IPinok))
	r.buildingAgg = agg

	return r
}

func (r *Router) run() {
	var encoder *gob.Encoder
	var file *os.File
	if ConfigGlobal.RecordPath != "" {
		file, err := os.Create(ConfigGlobal.RecordPath)
		if err != nil {
			log.Error("Could not open commands output file ", err)
		} else {
			encoder = gob.NewEncoder(file)
		}
	}

	for {
		select {
		case <-r.quit:
			log.Debug("Closing router...")
			r.buildingAgg.stop()
			if file != nil {
				err := file.Close()
				if err != nil {
					log.Error("Could not close commands output file ", err)
				}
			}
			return
		case op := <-r.newOperation:
			go op.Process(r.albionstate)
		case raw := <-r.recordRawPacket:
			if encoder != nil {
				err := encoder.Encode(raw)
				if err != nil {
					log.Error("Could not encode raw packet ", err)
				}
			}
		}
	}
}
