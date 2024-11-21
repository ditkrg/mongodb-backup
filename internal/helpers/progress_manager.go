package helpers

import (
	"fmt"
	"time"

	"github.com/mongodb/mongo-tools/common/progress"
	"github.com/rs/zerolog/log"
)

type ProgressManager struct{}

var stopProgress = make(chan bool)

func (p *ProgressManager) Attach(name string, progressor progress.Progressor) {
	_, max := progressor.Progress()
	log.Info().Msgf("dumping %s, total records %d", name, max)
	go process(name, progressor)
}

func (p *ProgressManager) Detach(name string) {
	stopProgress <- true
	log.Info().Msgf("finished dumping %s", name)

}

func process(name string, progressor progress.Progressor) {
	for {
		select {
		case <-stopProgress:
			return
		default:
			current, max := progressor.Progress()
			if current != 0 {
				log.Info().Msg(fmt.Sprintf("Progress: %s: %d/%d", name, current, max))
			}
			time.Sleep(1 * time.Second)
		}
	}
}
