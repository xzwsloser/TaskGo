package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/admin/internal/model/resp"
	"github.com/xzwsloser/TaskGo/admin/internal/service"
	"github.com/xzwsloser/TaskGo/pkg/logger"
)

type NodeHandler struct {
}

var (
	nodeHandler		= new(NodeHandler)
	nodeService		= new(service.NodeService)
)

// @Router: /node/delete
// @Method: POST
// @Description: Delete Node

func (n *NodeHandler) Delete(c *gin.Context) {
	var req request.ByUUID
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_node] request parameter error", c)
		return
	}

	err := nodeService.DeleteNode(req.UUID)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to delete node: %v",err))
		resp.FailWithMessage(resp.ERROR, "[delete node] failed", c)
		return 
	}

	resp.OkWithMessage("delete success", c)
}


// @Router: /node/search
// @Method: POST
// @Description: Search Node By Condition
func (n *NodeHandler) Search(c *gin.Context) {
	var req request.ReqNodeSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_node] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_node] request parameter error", c)
		return
	}
	req.Check()
	nodes, total, err := nodeService.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_node] search node error:%s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[search_node] search node  error", c)
		return
	}
	var resultNodes []resp.RspNodeSearch
	for _, node := range nodes {
		resultNode := resp.RspNodeSearch{
			Node: node,
		}
		resultNode.TaskCount, _ = service.GetNodeWatcherService().GetTasksCount(node.UUID)
		resultNodes = append(resultNodes, resultNode)
	}

	resp.OkWithDetailed(resp.PageResult{
		List:     resultNodes,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}




