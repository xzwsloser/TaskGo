package service

import (
	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/model"
)

type ScriptService struct {
}

func (*ScriptService) Search(r *request.ReqScriptSearch) ([]model.Script, int64, error) {
	script := &model.Script{}
	script.ID = r.ID
	script.Name = r.Name
	pageSize, page := r.PageSize, r.Page

	scripts, total, err := script.FindAndPage(page, pageSize)
	return scripts, total, err
}

