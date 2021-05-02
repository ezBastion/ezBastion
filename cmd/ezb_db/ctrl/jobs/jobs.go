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

package jobs

import (
	"net/http"

	"ezBastion/cmd/ezb_db/models"

	"ezBastion/cmd/ezb_db/tools"

	"github.com/gin-gonic/gin"
)

func Find(c *gin.Context) {
	var Raw []models.EzbJobs
	tools.Findraw(c, &Raw, "name asc")
}

func Findone(c *gin.Context) {
	t := "name"
	if tools.StrIsInt(c.Param("name")) {
		t = "id"
	}
	var Raw models.EzbJobs
	tools.Findoneraw(c, &Raw, t)
}
func Add(c *gin.Context) {
	var Raw models.EzbJobs
	tools.Addraw(c, &Raw)
}

func Update(c *gin.Context) {
	var Raw models.EzbJobs
	tools.Updateraw(c, &Raw)
}

func Remove(c *gin.Context) {
	var Raw models.EzbJobs
	var actions []models.EzbActions
	db, err := tools.Getdbconn(c)
	if err != "" {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	id := c.Param("id")
	if tools.StrIsInt(id) == false {
		c.JSON(http.StatusConflict, "WRONG PARAMETER")
		return
	}
	if err := db.First(&Raw, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Where("ezb_jobs_id = ?", Raw.ID).Find(&actions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	for _, action := range actions {
		action.EzbJobsID = 0
		if err := db.Save(&action).Error; err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := db.Delete(Raw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusNoContent, "")
}

func Enable(c *gin.Context) {
	var Raw models.EzbJobs
	if err := c.BindJSON(&Raw); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	tools.Enableraw(c, &Raw, Raw.Enable)
}
