/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-08-29 20:20:28
 * LastEditors: liguoqiang
 * LastEditTime: 2025-03-08 23:55:59
 * Description:
********************************************************************************/
/*
 * @Author: liguoqiang
 * @Date: 2021-07-23 16:25:27
 * @LastEditors: liguoqiang
 * @LastEditTime: 2023-09-27 02:35:16
 * @Description:
 */
package cfg

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	TmFmtStr        = "2006-01-02 15:04:05"
	DateFmtStr      = "2006-01-02"
	StaticPicPath   = "public/picture/"
	StaticVideoPath = "public/video/"
	StaticVoicePath = "public/voice/"
	StaticFilePath  = "public/file/"
)

type Cfg struct {
	Svr        SvrCfg      `yaml:"server"`
	DB         DbCfg       `yaml:"database"`
	Mq         MqCfg       `yaml:"mq"`
	Wx         WxCfg       `yaml:"wx"`
	Redis      RedisCfg    `yaml:"redis"`
	StaticPath string      `yaml:"staticPath"`
	Log        LogCfg      `yaml:"log"`
	AlarmMsg   AlarmMsgCfg `yaml:"alarm_msg"`
}

type SvrCfg struct {
	Host        string `yaml:"host"`
	OutUrl      string `yaml:"out_url"`
	ApiVersion  string `yaml:"api_version"`
	Location    string `yaml:"location"`
	EnableTls   bool   `yaml:"enable_tls"`
	CertFile    string `yaml:"cert_file"`
	KeyFile     string `yaml:"key_file"`
	CaFile      string `yaml:"ca_file"`
	EnableX1    bool   `yaml:"enable_x1"`
	EnableX1s   bool   `yaml:"enable_x1s"`
	EnableEd713 bool   `yaml:"enable_ed713"`
	EnableHl77  bool   `yaml:"enable_hl77"`
	EnableH03   bool   `yaml:"enable_h03"`
	EnableT1    bool   `yaml:"enable_t1"`
	EnableWx    bool   `yaml:"enable_wx"`
}
type DbCfg struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
}

type MqCfg struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	InstanceId string `yaml:"instance_id"`
	GroupId    string `yaml:"group_id"`
	ClientId   string `yaml:"client_id"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

type WxCfg struct {
	MinAppId                      string `yaml:"min_appId"`
	MinAppSecret                  string `yaml:"min_app_secret"`
	PublicAppId                   string `yaml:"public_appId"`
	PublicAppSecret               string `yaml:"public_app_secret"`
	PublicToken                   string `yaml:"public_token"`
	WelcomePath                   string `yaml:"welcome_path"`
	WelcomeTitle                  string `yaml:"welcome_title"`
	WelcomeCardMediaId            string `yaml:"welcome_media_id"`
	ReportPath                    string `yaml:"report_path"`
	ReportOfficalTemplateId       string `yaml:"report_offical_template_id"`
	ReportOfficalDayTemplateId    string `yaml:"report_offical_day_template_id"`
	ReportOfficalEveryTemplateId  string `yaml:"report_offical_every_template_id"`
	DeviceOnlineOfficalTemplateId string `yaml:"device_online_offical_template_id"`
	DeviceStatusOfficalTemplateId string `yaml:"device_status_offical_template_id"`
	ReportMiniTemplateId          string `yaml:"report_mini_template_id"`
	AccessTokenUri                string `yaml:"accessTokenUri"`
}

type RedisCfg struct {
	Host      string `yaml:"host"`
	Password  string `yaml:"password"`
	Db        int    `yaml:"db"`
	EnableTls bool   `yaml:"enable_tls"`
	CertFile  string `yaml:"cert_file"`
	KeyFile   string `yaml:"key_file"`
	CaFile    string `yaml:"ca_file"`
}

type LogCfg struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	Maxsize    int    `yaml:"maxsize"`
	Maxage     int    `yaml:"maxage"`
	Maxbackups int    `yaml:"maxbackups"`
	Console    bool   `yaml:"console"`
	Format     string `yaml:"format"`
}

type AlarmMsgCfg struct {
	PullRopeAlarmMsgCN          string `yaml:"pull_rope_alarm_msg_cn"`
	PullRopeAlarmMsgHK          string `yaml:"pull_rope_alarm_msg_hk"`
	PullRopeAlarmMsgEN          string `yaml:"pull_rope_alarm_msg_en"`
	DisarmedPullRopeAlarmMsgCN  string `yaml:"disarmed_pull_rope_alarm_msg_cn"`
	DisarmedPullRopeAlarmMsgHK  string `yaml:"disarmed_pull_rope_alarm_msg_hk"`
	DisarmedPullRopeAlarmMsgEN  string `yaml:"disarmed_pull_rope_alarm_msg_en"`
	ApneaAlarmMsgCN             string `yaml:"apnea_alarm_msg_cn"`
	ApneaAlarmMsgHK             string `yaml:"apnea_alarm_msg_hk"`
	ApneaAlarmMsgEN             string `yaml:"apnea_alarm_msg_en"`
	LeaveBedAlarmMsgCN          string `yaml:"leave_bed_alarm_msg_cn"`
	LeaveBedAlarmMsgHK          string `yaml:"leave_bed_alarm_msg_hk"`
	LeaveBedAlarmMsgEN          string `yaml:"leave_bed_alarm_msg_en"`
	InBedAlarmMsgCN             string `yaml:"in_bed_alarm_msg_cn"`
	InBedAlarmMsgHK             string `yaml:"in_bed_alarm_msg_hk"`
	InBedAlarmMsgEN             string `yaml:"in_bed_alarm_msg_en"`
	BreathingAbnormalAlarmMsgCN string `yaml:"breathing_abnormal_alarm_msg_cn"`
	BreathingAbnormalAlarmMsgHK string `yaml:"breathing_abnormal_alarm_msg_hk"`
	BreathingAbnormalAlarmMsgEN string `yaml:"breathing_abnormal_alarm_msg_en"`
	HeartRateAbnormalAlarmMsgCN string `yaml:"heart_rate_abnormal_alarm_msg_cn"`
	HeartRateAbnormalAlarmMsgHK string `yaml:"heart_rate_abnormal_alarm_msg_hk"`
	HeartRateAbnormalAlarmMsgEN string `yaml:"heart_rate_abnormal_alarm_msg_en"`
	CheckPersonActivityMsgCN    string `yaml:"check_person_activity_msg_cn"`
	CheckPersonActivityMsgHK    string `yaml:"check_person_activity_msg_hk"`
	CheckPersonActivityMsgEN    string `yaml:"check_person_activity_msg_en"`
	CheckPersonNoActivityMsgCN  string `yaml:"check_person_no_activity_msg_cn"`
	CheckPersonNoActivityMsgHK  string `yaml:"check_person_no_activity_msg_hk"`
	CheckPersonNoActivityMsgEN  string `yaml:"check_person_no_activity_msg_en"`
	CheckPersonBreathHighMsgCN  string `yaml:"check_person_breath_high_msg_cn"`
	CheckPersonBreathHighMsgHK  string `yaml:"check_person_breath_high_msg_hk"`
	CheckPersonBreathHighMsgEN  string `yaml:"check_person_breath_high_msg_en"`
	CheckPersonBreathLowMsgCN   string `yaml:"check_person_breath_low_msg_cn"`
	CheckPersonBreathLowMsgHK   string `yaml:"check_person_breath_low_msg_hk"`
	CheckPersonBreathLowMsgEN   string `yaml:"check_person_breath_low_msg_en"`
	CheckPersonHeartHighMsgCN   string `yaml:"check_person_heart_high_msg_cn"`
	CheckPersonHeartHighMsgHK   string `yaml:"check_person_heart_high_msg_hk"`
	CheckPersonHeartHighMsgEN   string `yaml:"check_person_heart_high_msg_en"`
	CheckPersonHeartLowMsgCN    string `yaml:"check_person_heart_low_msg_cn"`
	CheckPersonHeartLowMsgHK    string `yaml:"check_person_heart_low_msg_hk"`
	CheckPersonHeartLowMsgEN    string `yaml:"check_person_heart_low_msg_en"`
}

var This *Cfg = nil

func InitConfig(iniFile string) error {
	// _, fileName, _, _ := runtime.Caller(0)
	// filePath := path.Join(path.Dir(fileName), "cfg.yml")
	filePath := iniFile
	_, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("config file is not exist,%s", filePath)
		return err
	}
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("ReadFile config error,%v", err)
		return err
	}
	This = new(Cfg)
	err = yaml.Unmarshal(yamlFile, This)
	if err != nil {
		fmt.Printf("yaml unmarshal error, %v", err)
		return err
	}
	fmt.Printf("initialize config successful")
	return nil
}

func IsCN() bool {
	return This.Svr.Location == "zh-cn"
}
func IsHK() bool {
	return This.Svr.Location == "zh-hk"
}
