package constants

const (
	FILE_PATH = "../../file/"

	SAVE_JOB_DIR = "/cron/job/save/"

	KILL_JOB_DIR = "/cron/job/kill/"

	JOB_LOCK_DIR = "/cron/job/lock/"

	JOB_DEl_DIR = "/cron/job/delete/"

	//服务注册目录
	JOB_WORKER_DIR = "/cron/workers/"

	//主leader
	JOB_JOB_MASTER = "/cron/job/master/"

	//保存/更新任务事件
	SAVE_JOB_EVENT = 1

	//删除任务事件
	DELETE_JOB_EVENT = 2

	//强杀任务事件
	KILL_JOB_EVENT = 3

	//延时消息交换器
	DELAY_EXCHANGE = "delay_topic"

	//延时消息队列
	DELAY_KEY = "delay.*"

	//定时任务
	CRON_JOB_TYPE = 0

	//延时任务
	DELAY_JOB_TYPE = 1

	//节点在线
	NODE_ONLINE = 1
	//节点不在线
	NODE_OFFLINE = 0
)
