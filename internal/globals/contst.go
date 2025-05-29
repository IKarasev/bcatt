package globals

const (
	LOG_LVL_INFO  int = iota
	LOG_LVL_ERROR     = iota
	LOG_LVL_OK        = iota
	LOG_LVL_EVIL      = iota
)

const (
	LOG_DATE_FORMAT           = "[2006-01-02 15:04:05] "
	RSS_LOG_EVENT             = "log"
	RSS_EVENT_MINER_SET       = "rssMiner"
	RSS_EVENT_WALLET_COINS    = "rssCoins"
	RSS_EVENT_LASTBLOCK_H     = "rssBHeight"
	RSS_EVENT_LASTBLOCK_CB    = "rssBCB"
	RSS_EVENT_LASTBLOCK_NONCE = "rssBN"
	RSS_EVENT_LASTBLOCK_HASH  = "rssBH"
	RSS_EVENT_LASTBLOCK_ROOT  = "rssBR"
	RSS_EVENT_LASTBLOCK       = "rssLB"
	RSS_EVENT_NODE_COINBASE   = "rssNCB"
	RSS_EVENT_TICK            = "rssTick"
	RSS_NET_UPDATE            = "rssnetupdate"
	RSS_FORK_UPDATE           = "rssforkupdate"
)
