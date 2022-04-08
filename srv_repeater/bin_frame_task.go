package srv_repeater

type BinFrameTask struct {
	IsConnectedSignal    bool
	IsDisconnectedSignal bool

	Client *RepeaterBinClient
	Frame  *BinFrame
}
