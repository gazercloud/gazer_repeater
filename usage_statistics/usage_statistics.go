package usage_statistics

import "http-server.org/gazer/hostid"

type UsageStatistics struct {
	HostId       hostid.HostId `json:"host_id"`
	Comment      string
	SensorsCount int
}
