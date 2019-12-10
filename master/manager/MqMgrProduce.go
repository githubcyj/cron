package manager

import (
	"crontab/constants"
	"crontab/model"
	"encoding/json"
	"github.com/streadway/amqp"
	"strconv"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/11/26 16:05
  @todo
*/
type MqMgrProduce struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

var GMqMgrProduce *MqMgrProduce

func InitMq() (err error) {
	var (
		conn *amqp.Connection
		ch   *amqp.Channel
	)

	if conn, err = amqp.Dial("amqp://huye:huye@192.168.233.128:5672/test"); err != nil {
		return
	}
	//defer conn.Close()

	if ch, err = conn.Channel(); err != nil {
		return
	}
	//defer ch.Close()

	//声明交换器
	if err = ch.ExchangeDeclare(constants.DELAY_EXCHANGE, "topic", true, false, false, false, nil); err != nil {
		return
	}

	GMqMgrProduce = &MqMgrProduce{
		Conn: conn,
		Ch:   ch,
	}

	return
}

//将任务推送到消息队列
func (m *MqMgrProduce) PushMq(pipeline *model.Pipeline) (err error) {
	var (
		pipelineStr []byte
		now         time.Time
		diff        time.Duration
	)
	//序列化
	if pipelineStr, err = json.Marshal(pipeline); err != nil {
		return
	}
	//设置过期时间
	loc, _ := time.LoadLocation("Local")
	the_time, _ := time.ParseInLocation("2006-01-02 15:04:05", pipeline.TimerExecuter, loc)

	//获取当前时间
	now = time.Now()
	//与当前时间的差
	diff = now.Sub(the_time)

	if err = m.Ch.Publish(
		constants.DELAY_EXCHANGE,
		constants.DELAY_KEY,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        pipelineStr,
			Expiration:  strconv.FormatFloat(diff.Seconds(), 'E', -1, 64), //设置过期时间
		},
	); err != nil {
		return
	}
	return
}
