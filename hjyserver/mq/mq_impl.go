/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-09-06 17:50:12
 * LastEditors: liguoqiang
 * LastEditTime: 2023-12-23 10:51:41
 * Description:
********************************************************************************/
package mq

import (
	"encoding/json"
	"fmt"
	"hjyserver/cfg"
	"hjyserver/gopool"
	mylog "hjyserver/log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var mqttClient mqtt.Client
var taskPool *gopool.Pool = nil
var mqConnected bool = false

/******************************************************************************
 * description: define a interface to process mqtt message
 * others can implement this interface to process mqtt message
 * return {*}
********************************************************************************/
type MessageProc interface {
	HandleMqttMsg(topic string, payload []byte)
}

var topicMap = make(map[string]MessageProc)

var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var t = gopool.Task{
		Params: []interface{}{msg.Topic(), msg.Payload()},
		Do: func(params ...interface{}) {
			var topic = params[0].(string)
			var payload = params[1].([]byte)
			proc, exist := topicMap[topic]
			if exist {
				proc.HandleMqttMsg(topic, payload)
			}
		},
	}
	if taskPool != nil {
		taskPool.Put(&t)
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	mylog.Log.Infoln("mqtt connected")
	mqConnected = true
	for k, _ := range topicMap {
		topic := k
		token := client.Subscribe(topic, 0, nil)
		if token.Wait() && token.Error() != nil {
			mylog.Log.Errorln(token.Error())
		}
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	mylog.Log.Infoln("mqtt disconnected")
	mqConnected = false
}

func InitMqtt() bool {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.This.Mq.Host, cfg.This.Mq.Port))
	opts.SetClientID(cfg.This.Mq.ClientId)
	opts.SetUsername(cfg.This.Mq.Username)
	opts.SetPassword(cfg.This.Mq.Password)
	opts.SetDefaultPublishHandler(msgHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetCleanSession(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetConnectTimeout(60 * time.Second)
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		mylog.Log.Errorln(token.Error())
		return false
	}
	taskPool, _ = gopool.InitPool(64)
	return true
}

/******************************************************************************
 * function: Close
 * description: close mqtt client
 * return {*}
********************************************************************************/
func CloseMqtt() {
	taskPool.Close()
	if mqttClient != nil {
		UnsubscribeAllTopic()
		mqttClient.Disconnect(250)
	}
}

/******************************************************************************
 * function: SubscribeTopic
 * description: subscribe custom topic, that is not default topic
 * return {*}
********************************************************************************/
func SubscribeTopic(topic string, msgProc MessageProc) bool {
	topicMap[topic] = msgProc
	if mqttClient != nil && mqConnected {
		token := mqttClient.Subscribe(topic, 0, nil)
		return token.WaitTimeout(2 * time.Second)
	}
	return false
}

/******************************************************************************
 * function: UnsubscribeTopic
 * description: unsubscribe a topic
 * return {*}
********************************************************************************/
func UnsubscribeTopic(topic string) bool {
	delete(topicMap, topic)
	if mqttClient != nil && mqConnected {
		token := mqttClient.Unsubscribe(topic)
		return token.WaitTimeout(2 * time.Second)
	}
	return true
}
func UnsubscribeAllTopic() bool {
	for k, _ := range topicMap {
		topic := k
		if mqttClient != nil && mqConnected {
			token := mqttClient.Unsubscribe(topic)
			token.WaitTimeout(1 * time.Second)
		}
		delete(topicMap, k)
	}
	return true
}

/******************************************************************************
 * function: PublishData
 * description:
 * return {*}
********************************************************************************/
func PublishData(topic string, payload interface{}) bool {
	jsBytes, err := json.Marshal(payload)
	if err != nil {
		mylog.Log.Errorln("json marshal failed, err:", err)
		return false
	}
	mylog.Log.Infoln("topic:", topic, ", payload:", string(jsBytes))
	if mqttClient != nil {
		token := mqttClient.Publish(topic, 0, false, jsBytes)
		return token.WaitTimeout(2 * time.Second)
	}
	return false
}
