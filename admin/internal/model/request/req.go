package request

type (
	PageInfo struct {
		Page		int		`json:"page" form:"page"`
		PageSize	int		`json:"page_size" form:"page_size"`
	}

	ByID struct {
		ID	int	`json:"id" form:"id"`
	}

	ByIDS struct {
		IDS	[]int	`json:"ids" form:"ids"`
	}
)


func (p *PageInfo) Check() {
	if p.PageSize <= 0 {
		p.PageSize = 20
	}

	if p.Page <= 0 {
		p.Page = 1
	}
}

