package manager

import (
	"github.com/crontab/constants"
	"github.com/streadway/amqp"
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

	//ch.Confirm()

	GMqMgrProduce = &MqMgrProduce{
		Conn: conn,
		Ch:   ch,
	}

	return
}
