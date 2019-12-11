package worker

import (
	"encoding/json"
	"github.com/crontab/common"
	"github.com/crontab/constants"
	"github.com/crontab/model"
	"github.com/streadway/amqp"
)

/**
  @author 胡烨
  @version 创建时间：2019/11/26 15:24
  @todo
*/

type MqMgrCustomer struct {
	Conn   *amqp.Connection
	Ch     *amqp.Channel
	QDelay amqp.Queue
}

var GMqMgrCustomer *MqMgrCustomer

func InitMq() (err error) {
	var (
		conn   *amqp.Connection
		ch     *amqp.Channel
		qNow   amqp.Queue
		qDelay amqp.Queue
	)
	if conn, err = amqp.Dial("amqp://huye:huye@192.168.233.128:5672/test"); err != nil {
		return
	}
	defer conn.Close()

	if ch, err = conn.Channel(); err != nil {
		return
	}
	defer ch.Close()

	//声明交换器
	if err = ch.ExchangeDeclare(constants.DELAY_EXCHANGE, "topic", true, false, false, false, nil); err != nil {
		return
	}

	//声明一个普通队列，该队列接收到消息就马上处理
	if qNow, err = ch.QueueDeclare(
		"qNow",
		true,
		false,
		true,
		false,
		nil,
	); err != nil {
		return
	}

	//声明延时队列，该队列中消息如果过期，就将消息发送到交换器上，交换器就分发消息到普通队列
	if qDelay, err = ch.QueueDeclare(
		"qDelay",
		true,
		false,
		true,
		false,
		amqp.Table{
			//当消息过期时把消息发送到logs这个交换器
			"x-dead-letter-exchange":    constants.DELAY_EXCHANGE,
			"x-dead-letter-routing-key": "now.t",
		},
	); err != nil {
		return
	}
	//绑定队列
	if err = ch.QueueBind(
		qNow.Name,
		"now.*",
		constants.DELAY_EXCHANGE,
		false,
		nil,
	); err != nil {
		return
	}

	//再绑定一个队列
	if err = ch.QueueBind(
		qDelay.Name,
		constants.DELAY_KEY,
		constants.DELAY_EXCHANGE,
		false,
		nil,
	); err != nil {
		return
	}
	GMqMgrCustomer = &MqMgrCustomer{
		Conn:   conn,
		Ch:     ch,
		QDelay: qNow,
	}
	go GMqMgrCustomer.listenMg()
	return
}

//监听队列
func (m *MqMgrCustomer) listenMg() {
	var (
		msg  amqp.Delivery
		msgs <-chan amqp.Delivery
	)

	msgs, _ = m.Ch.Consume(
		m.QDelay.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		var (
			pipeline *model.Pipeline
			err      error
			jobEvent *common.JobEvent
		)
		for {
			select {
			case msg = <-msgs: //从消息队列中读取到任务
				//反序列化
				if err = json.Unmarshal(msg.Body, pipeline); err != nil {
					GLogMgr.WriteLog("延时任务读取失败：", err.Error())
				}
			}

			//创建事件
			jobEvent = common.BuildJobEvent(pipeline, constants.DELAY_JOB_TYPE)
			//通知调度器
			GScheduler.PushScheduler(jobEvent)
		}
	}()
}
