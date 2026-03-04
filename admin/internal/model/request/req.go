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




