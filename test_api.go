package main

import (
	"encoding/json"
	"fmt"
	"github.com/djumanoff/amqp"
	"github.com/google/uuid"
	setdata_common "github.com/kirigaikabuto/setdata-common"
	tik_lib "github.com/kirigaikabuto/tik-lib"
	"github.com/spf13/viper"
	"log"
	"os"
)

var (
	currentUser    *tik_lib.User
	configNameTest = "main"
	configPathTest = "./config/"
	amqpHostTest   = ""
	amqpPortTest   = ""
	amqpUrlTest    = ""
)

func parse() {
	filepath, err := os.Getwd()
	if err != nil {
		panic("main, get rootDir error" + err.Error())
		return
	}
	viper.AddConfigPath(filepath + configPathTest)
	viper.SetConfigName(configNameTest)
	err = viper.ReadInConfig()
	if err != nil {
		panic("main, fatal error while reading config file: " + err.Error())
		return
	}
	amqpHostTest = viper.GetString("rabbit.primary.host")
	amqpPortTest = viper.GetString("rabbit.primary.port")
	amqpUrlTest = viper.GetString("rabbit.primary.url")
	if amqpUrlTest == "" {
		amqpUrlTest = "amqps://" + amqpHostTest + ":" + amqpPortTest
	}
}

func CreateUserTest(clt amqp.Client) error {
	hashPass, err := setdata_common.HashPassword(viper.GetString("user.password"))
	if err != nil {
		return err
	}
	userTest := &tik_lib.User{
		Id:                  uuid.New().String(),
		FirstName:           viper.GetString("user.first_name"),
		LastName:            viper.GetString("user.last_name"),
		Username:            viper.GetString("user.username"),
		PhoneNumber:         viper.GetString("user.phone_number"),
		Email:               viper.GetString("user.email"),
		Password:            hashPass,
		AvatarUrl:           viper.GetString("user.avatar_url"),
		EmailVerified:       viper.GetBool("user.email_verified"),
		PhoneNumberVerified: viper.GetBool("user.phone_number_verified"),
		TypeOfUser:          tik_lib.ToUserType(viper.GetString("user.type_of_user")),
	}
	jsonBody, err := json.Marshal(userTest)
	if err != nil {
		return err
	}
	resp, err := clt.Call("user.create", amqp.Message{Body: jsonBody})
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp.Body, &currentUser)
	if err != nil {
		return err
	}
	fmt.Println(string(resp.Body))
	return nil
}

func ListUsers(clt amqp.Client) error {
	cmd := &tik_lib.ListUserCommand{TypeOfUser: "buyer"}
	jsonBody, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	resp, err := clt.Call("user.list", amqp.Message{Body: jsonBody})
	if err != nil {
		return err
	}
	fmt.Println(string(resp.Body))
	return nil
}

func GetUserById(clt amqp.Client) error {
	cmd := &tik_lib.GetUserByIdCommand{Id: currentUser.Id}
	jsonBody, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	resp, err := clt.Call("user.getById", amqp.Message{Body: jsonBody})
	if err != nil {
		return err
	}
	fmt.Println(string(resp.Body))
	return nil
}

func GetUserByPhoneNumber(clt amqp.Client) error {
	cmd := &tik_lib.GetUserByPhoneNumberCommand{PhoneNumber: currentUser.PhoneNumber}
	jsonBody, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	resp, err := clt.Call("user.getByPhoneNumber", amqp.Message{Body: jsonBody})
	if err != nil {
		return err
	}
	fmt.Println(string(resp.Body))
	return nil
}

func main() {
	parse()
	amqpConfig := amqp.Config{
		AMQPUrl: amqpUrlTest,
	}
	sess := amqp.NewSession(amqpConfig)
	err := sess.Connect()
	if err != nil {
		log.Fatal(err)
		return
	}
	clt, err := sess.Client(amqp.ClientConfig{})
	if err != nil {
		log.Fatal(err)
		return
	}
	err = CreateUserTest(clt)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = ListUsers(clt)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = GetUserById(clt)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = GetUserByPhoneNumber(clt)
	if err != nil {
		log.Fatal(err)
		return
	}
}
