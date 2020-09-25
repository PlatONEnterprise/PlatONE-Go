package controller

type page struct {
	PageIndex int64 `form:"page_index"`
	PageSize  int64 `form:"page_size"`
}

type pageInfo struct {
	PageIndex int64       `json:"page_index"`
	PageSize  int64       `json:"page_size"`
	Total     int64       `json:"total"`
	Items     interface{} `json:"items"`
}

func setPageDefaultIfEmpty(p *page) {
	if p.PageSize <= 0 {
		p.PageSize = 50
	}

	if p.PageIndex <= 0 {
		p.PageIndex = 1
	}
}

func newPageInfo(pageIndex, pageSize, total int64, items interface{}) *pageInfo {
	return &pageInfo{PageIndex: pageIndex, PageSize: pageSize, Total: total, Items: items}
}

type contract struct {
	Address   string `json:"address"`
	CNSName   string `json:"name"`
	Creator   string `json:"creator"`
	TxHash    string `json:"tx_hash"`
	Timestamp int64  `json:"timestamp"`
}
