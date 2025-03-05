/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-09-06 17:50:12
 * LastEditors: liguoqiang
 * LastEditTime: 2024-12-06 17:35:43
 * Description:
********************************************************************************/
package mq

import (
	"encoding/json"
	"fmt"
	"hjyserver/cfg"
	"hjyserver/gopool"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var mqttClient mqtt.Client
var taskPool *gopool.Pool = nil
var mqConnected bool = false
var uuid string = common.GenerateUUID()

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
			} else {
				for k, v := range topicMap {
					if matchTopic(k, topic) {
						v.HandleMqttMsg(topic, payload)
						return
					}
				}
			}
		},
	}
	if taskPool != nil {
		taskPool.Put(&t)
	}
}

func matchTopic(subscribed, topic string) bool {
	subLevels := strings.Split(subscribed, "/")
	topicLevels := strings.Split(topic, "/")

	match := true
	for i, subLevel := range subLevels {
		if subLevel == "#" || subLevel == "+" {
			continue
		}
		if subLevel != topicLevels[i] {
			match = false
			break
		}
	}
	return match
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	mylog.Log.Infoln("mqtt connected")
	mqConnected = true
	for k, _ := range topicMap {
		topic := k
		token := client.Subscribe(topic, 0, msgHandler)
		token.Wait()
		if token.Error() != nil {
			mylog.Log.Errorln("Subscribe error:", token.Error())
		} else {
			mylog.Log.Infoln("Subscribe success to", topic)
		}
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	mylog.Log.Infoln("mqtt disconnected")
	mqConnected = false
}

func getMqClientId() string {
	clientId := fmt.Sprintf("%s@@@%s", cfg.This.Mq.GroupId, uuid)
	return clientId
}
func getMqUserName() string {
	userName := fmt.Sprintf("Signature|%s|%s", cfg.This.Mq.AccessKey, cfg.This.Mq.InstanceId)
	return userName
}
func getMqPassword() string {
	clientId := getMqClientId()
	password, err := common.GenerateMQPassword(clientId, cfg.This.Mq.SecretKey)
	if err != nil {
		mylog.Log.Errorln("generate mq password failed, err:", err)
		return cfg.This.Mq.Password
	}
	return password
}

/******************************************************************************
 * function: InitMqtt
 * description: MQ 初始化函数，在项目启动时调用
 * return {*}
********************************************************************************/
func InitMqtt() bool {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.This.Mq.Host, cfg.This.Mq.Port))
	opts.SetClientID(getMqClientId())
	if cfg.This.Mq.Username != "" {
		opts.SetUsername(getMqUserName())
		opts.SetPassword(getMqPassword())
	}
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
		token := mqttClient.Subscribe(topic, 0, msgHandler)
		token.Wait()
		if token.Error() != nil {
			mylog.Log.Errorln("Subscribe error:", token.Error())
			return false
		} else {
			mylog.Log.Infoln("Subscribe success to", topic)
			return true
		}
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
