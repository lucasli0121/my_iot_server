/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-07 23:43:02
 * LastEditors: liguoqiang
 * LastEditTime: 2024-11-21 16:22:23
 * Description:
********************************************************************************/
package common

import (
	"testing"
)

func TestEncriptData(t *testing.T) {
	data := "888888"
	result, err := EncryptDataWithDefaultkey(data)
	if err != nil {
		t.Errorf("TestEncriptData error:%v", err)
		return
	}
	t.Errorf("TestEncriptData success, %v", result)
}

func TestDecriptData(t *testing.T) {
	data := "j/gtI5oC4uOljjXit8qrK5COf3DSDiGVYIcVmjYnKEwZrUSAW+Uevhc8KnyRyw8T"
	result, err := DecryptDataNoCBCWithDefaultkey(data)
	if err != nil {
		t.Errorf("TestDecriptData error:%v", err)
		return
	}
	t.Errorf("TestDecriptData success, %v", result)
}

func TestMQPassword(t *testing.T) {
	clientId := "GID_hjyserver_mqtt_client_1"
	secretKey := "5wRKdXzIeWHBIBeYeUqLPDFlo2vbjx"
	result, err := GenerateMQPassword(clientId, secretKey)
	if err != nil {
		t.Errorf("TestMQPassword error:%v", err)
		return
	}
	t.Errorf("TestMQPassword success, %v", result)
}

func TestUUID(t *testing.T) {
	result := GenerateUUID()
	t.Errorf("TestUUID success, %v", result)
}

func TestTimeRandom(t *testing.T) {
	result := GenerateTimeRandom()
	t.Errorf("TestTimeRandom success, %v", result)
}
