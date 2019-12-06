package manager

/**
  @author 胡烨
  @version 创建时间：2019/11/18 15:12
  @todo 初始化一些内容，如将任务全部放入reids中
*/
//
////查询数据库获得所有任务，然后将任务放进redis中
//func InitJobsToRedis() (err error) {
//	var (
//		jobs []*common.Job
//	)
//	if jobs, err = GDB.ListJob(); err != nil {
//		return
//	}
//
//	GRedis.AddAllJob(jobs)
//
//	return
//}
