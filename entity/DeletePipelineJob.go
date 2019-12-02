package entity

import (
	"crontab/master/manager"
	"crontab/model"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/2 14:25
  @todo
*/
type DeleteIds struct {
	PipelineId string
	JobIds     []string
}

//redis删除job
func (d *DeleteIds) DelJobsRedis() (err error) {
	var (
		id string
	)
	for _, id = range d.JobIds {
		if _, err = manager.GRedis.Conn.Do("ZREM", d.PipelineId, id); err != nil {
			manager.GLogMgr.WriteLog("任务" + id + "删除失败")
			continue
		}
	}
	return nil
}

//数据库删除job
func (d *DeleteIds) DelJobsDB() (err error) {
	var (
		jobId string
	)
	for _, jobId = range d.JobIds {
		if err = manager.GDB.DB.Where("job_id = ? and pipeline_id", jobId, d.PipelineId).Delete(model.PipelineJob{}).Error; err != nil {
			return
		}
	}

	return
}
