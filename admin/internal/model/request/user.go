package request

type (
	ReqUserLogin struct {
		UserName string `json:"username" form:"username" binding:"required,min=2,max=20"`
		Password string `json:"password" form:"password" binding:"required,min=4,max=20,alphanum"`
	}
	ReqUserRegister struct {
		UserName string `json:"username" form:"username" binding:"required,min=2,max=20"`
		Password string `json:"password" form:"password" binding:"required,min=4,max=20,alphanum"`
		Email    string `json:"email" form:"email"`
		Role     int    `json:"role" form:"email"`
	}
)
