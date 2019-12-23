package server

import (
	"github.com/crontab/common"
	"github.com/crontab/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/16 11:19
  @todo
*/

//获得所有的node节点，包括在线与不在线
func HandlerNodeList(c *gin.Context) {
	var (
		node  *model.Node
		err   error
		list  []*model.Node
		bytes common.HttpReply
	)
	node = &model.Node{}
	if list, err = node.GetAllNode(); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", list)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
	return

ERR:
	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
}

//流水线绑定node节点，如果流水线绑定了node节点，则该流水线只能在该节点执行，否则则在所有节点执行
func HandlerNodePipeline(c *gin.Context) {
	var (
		bytes        common.HttpReply
		pipelineNode *model.PipelineNode
		err          error
		pipeline     *model.Pipeline
		pipelineR    *model.Pipeline
	)
	pipelineNode = &model.PipelineNode{}

	if err = c.ShouldBindBodyWith(pipelineNode, binding.JSON); err != nil {
		goto ERR
	}

	//保存进入数据库
	if err = pipelineNode.SaveDB(); err != nil {
		goto ERR
	}

	if err = pipelineNode.SaveRedis(); err != nil {
		goto ERR
	}

	//修改流水线中的nodes内容
	pipeline = &model.Pipeline{PipelineId: pipelineNode.PipelineId}
	if pipelineR, err = pipeline.GetRedis(); err != nil {
		goto ERR
	}
	pipelineR.Nodes = append(pipeline.Nodes, pipelineNode.NodeId)

	//3. 删除原本redis中的数据
	if err = pipelineR.DelRedis(); err != nil {
		goto ERR
	}

	//重新存入redis
	if err = pipelineR.SaveRedis(); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", pipelineR)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
	return

ERR:
	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
}
