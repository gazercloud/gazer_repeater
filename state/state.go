package state

import "time"

type Channel struct {
	ChannelId  string   `json:"channel_id"`
	Routes     []string `json:"routes"`
	DataSize   int      `json:"data_size"`
	DataSource string   `json:"data_source"`
}

type Worker struct {
	Channels []Channel `json:"channels"`
	Auth     *Users    `json:"auth"`
}

type Node struct {
	NodeId string `json:"node_id"`
}

type Repeater struct {
	Nodes []Node `json:"nodes"`
	Auth  *Users `json:"auth"`
}

type Reg struct {
}

type Rdb struct {
}

type Edge struct {
	Channels []Channel `json:"channels"`
}

type ServerStatistics struct {
	Name         string  `json:"name"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemUsage     float32 `json:"mem_usage"`
	Mem          float64 `json:"mem"`
	FDs          int     `json:"f_ds"`
	Connections  int     `json:"connections"`
	ThreadsCount int     `json:"threads_count"`
	Status       string  `json:"status"`
	InnerState   string  `json:"inner_state"`
	TrafficIn    float64 `json:"traffic_in"`
	TrafficOut   float64 `json:"traffic_out"`
	Scores       int     `json:"scores"`
}

type OnlineChannel struct {
	ChannelId string `json:"channel_id"`
	Host      string `json:"host"`
}

type ChannelsIDs struct {
	NextId      uint64 `json:"next_id"`
	NextIdSaved uint64 `json:"next_id_saved"`

	// statistics
	StatRegCount                  int `json:"stat_reg_count"`
	StatSaveNextIdCount           int `json:"stat_save_next_id_count"`
	StatSaveNextIdCountTry        int `json:"stat_save_next_id_count_try"`
	StatSaveNextIdCountSuccess    int `json:"stat_save_next_id_count_success"`
	StatSaveNextIdCountErrorOpen  int `json:"stat_save_next_id_count_error_open"`
	StatSaveNextIdCountErrorWrite int `json:"stat_save_next_id_count_error_write"`
	StatSaveNextIdCountErrorClose int `json:"stat_save_next_id_count_error_close"`
}

type Router struct {
	Auth                  *Users             `json:"auth"`
	ServerStatisticsItems []ServerStatistics `json:"server_statistics_items"`
	OnlineChannels        []OnlineChannel    `json:"online_channels"`
	ChannelsIDs           *ChannelsIDs       `json:"channels_ids"`
	Info                  string             `json:"info"`
}

type Public struct {
}

type Session struct {
	Disposed  bool      `json:"disposed"`
	Id        string    `json:"id"`
	BeginTime time.Time `json:"begin_time"`
	UserName  string    `json:"user_name"`
}

type User struct {
	Name     string    `json:"name"`
	Sessions []Session `json:"sessions"`
}

type Users struct {
	Users []User `json:"users"`
}

type System struct {
	Worker   *Worker   `json:"worker"`
	Router   *Router   `json:"router"`
	Public   *Public   `json:"public"`
	Edge     *Edge     `json:"edge"`
	Repeater *Repeater `json:"repeater"`
	Reg      *Reg      `json:"reg"`
	Rdb      *Rdb      `json:"rdb"`
}
