// This file is part of ezBastion.

//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

package middleware

import (
	"errors"
	"ezBastion/pkg/confmanager"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"ezBastion/cmd/ezb_srv/models"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
func SelectWorker(conf *confmanager.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		tr, _ := c.Get("trace")
		trace := tr.(models.EzbLogs)
		logg := log.WithFields(log.Fields{
			"middleware": "SelectWorker",
			"xtrack":     trace.Xtrack,
		})
		logg.Debug("start")

		routeType, _ := c.MustGet("routeType").(string)

		if routeType == "worker" {
			ac, _ := c.Get("action")
			action := ac.(models.EzbActions)
			workers := action.Workers
			nbW := len(workers)
			if nbW == 0 {
				logg.Error("NO WORKER FOUND FOR ", action.Name)
				c.AbortWithError(http.StatusServiceUnavailable, errors.New("#W0001"))
				return
			}
			if nbW == 1 {
				worker := workers[0]
				if !worker.Enable {
					logg.Error("ONLY ONE DISABLE WORKER ", worker.Name, " FOUND FOR ", action.Name)
					c.AbortWithError(http.StatusServiceUnavailable, errors.New("#W0002"))
					return
				}
				logg.Debug("one worker found: ", worker.Name, " (", worker.Comment, ")")
				c.Set("worker", worker)

				trace.Worker = worker.Name
				c.Set("trace", trace)
			}
			if nbW > 1 {
				var enableWorkers []models.EzbWorkers
				//var keys []int
				var wg sync.WaitGroup
				for _, w := range workers {
					if w.Enable {
						wg.Add(1)
						go checksumISok(&w, &action, &wg, &enableWorkers)
					}
				}
				wg.Wait()
				if len(enableWorkers) == 0 {
					logg.Error("NO WORKERS AVAILABLE FOR ", action.Name)
					c.AbortWithError(http.StatusServiceUnavailable, errors.New("#W0003"))
					return
				}
				var worker models.EzbWorkers

				// switch on config worker algo
				switch conf.EZBSRV.LB {
				case "rrb":
					sort.Slice(workers, func(i, j int) bool {
						return workers[i].LastRequest.Unix() < workers[j].LastRequest.Unix()
					})
					worker = workers[0]
					break
				default:
					//random
					j := rand.Intn(len(workers))
					if j < len(workers) {
						worker = workers[j]
					} else {
						worker = workers[j-1]
					}
					break
				}

				logg.Debug("found ", len(enableWorkers), " worker, select ", worker.Name, "by", conf.EZBSRV.LB)
				c.Set("worker", worker)

				trace.Worker = worker.Name
				c.Set("trace", trace)
				c.Next()
			}
		}
		c.Next()
	}
}

func checksumISok(w *models.EzbWorkers, action *models.EzbActions, wg *sync.WaitGroup, enableWorkers *[]models.EzbWorkers) bool {
	defer wg.Done()
	if action.Jobs.Checksum == "" {
		*enableWorkers = append(*enableWorkers, *w)
		return true
	}
	/*
		connect to w.fqdn  /healthcheck/scripts  (cache)
		type wksScript struct {
			Name     string `json:name`
			Path     string `json:path`
			Checksum string `json:checksum`
		}
	*/
	*enableWorkers = append(*enableWorkers, *w)
	return true
}
