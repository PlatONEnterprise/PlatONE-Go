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

func newPageInfo(pageIndex, pageSize, total int64, items interface{}) *pageInfo {
	return &pageInfo{PageIndex: pageIndex, PageSize: pageSize, Total: total, Items: items}
}
