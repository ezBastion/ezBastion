package wkslog

import "time"

type Logs struct {
	FieldsTime time.Time `json:"fields.time"`
	File       string    `json:"file"`
	Func       string    `json:"func"`
	Ip         string    `json:"ip"`
	Latency    int64     `json:"latency"`
	Level      string    `json:"level"`
	Method     string    `json:"method"`
	Msg        string    `json:"msg"`
	Path       string    `json:"path"`
	Status     int       `json:"status"`
	Time       time.Time `json:"time"`
	UserAgent  string    `json:"user-agent"`
	Controller string    `json:"controller"`
	Xtrack     string    `json:"xtrack"`
}
