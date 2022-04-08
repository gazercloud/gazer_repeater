package srv_repeater

import (
	"sync"
	"time"
)

type Node struct {
	NodeId             string
	NodeName           string
	mtx                sync.Mutex
	LastWriteDT        time.Time
	sourceClient       *RepeaterBinClient
	clientsToTranslate map[*RepeaterBinClient]bool
	UserId             int64
}
