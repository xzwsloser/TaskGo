package request

type (
	ReqNodeSearch struct {
		PageInfo
		IP     string `json:"ip" form:"ip"` // node ip
		UUID   string `json:"uuid" form:"uuid"`
		UpTime int64  `json:"up" form:"up"`        // start time
		Status int    `son:"status" form:"status"` // status
	}

	ByUUID struct {
		UUID string `json:"uuid" form:"uuid"`
	}
)

func (r *ReqNodeSearch) Check() {
	if r.PageSize <= 0 {
		r.PageSize = 10
	}

	if r.Page <= 0 {
		r.Page = 1
	}
}
