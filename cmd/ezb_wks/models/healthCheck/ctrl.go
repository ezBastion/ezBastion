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

package healthCheck

import (
	"crypto/sha256"
	"ezBastion/pkg/confmanager"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func dealwithErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
func getJobs(c *gin.Context) {
	var thread wksThread
	thread.Thread = runtime.NumGoroutine()
	// conf, _ := c.MustGet("conf").(models.Configuration)
	// Limit := conf.Limit
	c.JSON(http.StatusOK, thread)
}

func getConf(c *gin.Context) {
	conf, _ := c.MustGet("conf").(confmanager.Configuration)
	c.JSON(http.StatusOK, conf)
}
func getVersion(c *gin.Context) {
	version, _ := c.MustGet("version").(string)
	c.JSON(http.StatusOK, gin.H{"version": version})
}

// func setConf(c *gin.Context)  {
// 	conf, _ := c.MustGet("conf").(models.Configuration)
// }

func getLoad(c *gin.Context) {
	fmt.Println("get load")
	var load wksLoad
	percentage, err := cpu.Percent(0, false)
	dealwithErr(err)
	vmStat, err := mem.VirtualMemory()
	dealwithErr(err)
	load.CPU = strconv.FormatFloat(percentage[0], 'f', 2, 64)
	load.MEM = strconv.FormatFloat(vmStat.UsedPercent, 'f', 2, 64)
	c.JSON(http.StatusOK, load)
}

func getScripts(c *gin.Context) {
	var scripts []wksScript
	conf, _ := c.MustGet("conf").(confmanager.Configuration)
	err := filepath.Walk(conf.EZBWKS.ScriptPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			var f wksScript
			f.Name = info.Name()
			f.Path = path[len(conf.EZBWKS.ScriptPath):]
			data, _ := ioutil.ReadFile(path)
			f.Checksum = fmt.Sprintf("%x", sha256.Sum256(data))
			scripts = append(scripts, f)
		}
		return nil
	})
	dealwithErr(err)

	c.JSON(http.StatusOK, scripts)
}
