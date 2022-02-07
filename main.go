package main

import (
	"github.com/djumanoff/amqp"
	setdata_common "github.com/kirigaikabuto/setdata-common"
	tik_lib "github.com/kirigaikabuto/tik-lib"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"log"
	"os"
)

var (
	configName           = "main"
	configPath           = "/config/"
	version              = "0.0.1"
	amqpHost             = ""
	amqpPort             = ""
	amqpUrl              = ""
	postgresUser         = ""
	postgresPassword     = ""
	postgresDatabaseName = ""
	postgresHost         = ""
	postgresPort         = 5432
	postgresParams       = ""
	flags                = []cli.Flag{
		&cli.StringFlag{
			Name:        "config, c",
			Usage:       "path to .env config file",
			Destination: &configPath,
		},
	}
)

func parseEnvFile() {
	filepath, err := os.Getwd()
	if err != nil {
		panic("main, get rootDir error" + err.Error())
		return
	}
	viper.AddConfigPath(filepath + configPath)
	viper.SetConfigName(configName)
	err = viper.ReadInConfig()
	if err != nil {
		panic("main, fatal error while reading config file: " + err.Error())
		return
	}
	amqpHost = viper.GetString("rabbit.primary.host")
	amqpPort = viper.GetString("rabbit.primary.port")
	amqpUrl = viper.GetString("rabbit.primary.url")
	if amqpUrl == "" {
		amqpUrl = "amqps://" + amqpHost + ":" + amqpPort
	}
	postgresUser = viper.GetString("db.primary.user")
	postgresPassword = viper.GetString("db.primary.pass")
	postgresDatabaseName = viper.GetString("db.primary.name")
	postgresParams = viper.GetString("db.primary.param")
	postgresPort = viper.GetInt("db.primary.port")
	postgresHost = viper.GetString("db.primary.host")
}

func run(c *cli.Context) error {
	parseEnvFile()
	rabbitConfig := amqp.Config{
		AMQPUrl:  amqpUrl,
		LogLevel: 5,
	}
	serverConfig := amqp.ServerConfig{
		ResponseX: "response",
		RequestX:  "request",
	}
	sess := amqp.NewSession(rabbitConfig)
	err := sess.Connect()
	if err != nil {
		return err
	}
	srv, err := sess.Server(serverConfig)
	if err != nil {
		return err
	}
	cfg := tik_lib.PostgresConfig{
		Host:     postgresHost,
		Port:     postgresPort,
		User:     postgresUser,
		Password: postgresPassword,
		Database: postgresDatabaseName,
		Params:   postgresParams,
	}
	userStore, err := tik_lib.NewPostgreUserStore(cfg)
	if err != nil {
		return err
	}
	fileStore, err := tik_lib.NewPostgresFileStore(cfg)
	if err != nil {
		return err
	}
	service := tik_lib.NewService(userStore, fileStore)
	amqpEndpoints := tik_lib.NewAmqpEndpoints(setdata_common.NewCommandHandler(service))
	srv.Endpoint("user.create", amqpEndpoints.CreateUser())
	srv.Endpoint("user.getById", amqpEndpoints.GetUserById())
	srv.Endpoint("user.getByPhoneNumber", amqpEndpoints.GetUserByPhoneNumber())
	srv.Endpoint("user.update", amqpEndpoints.UpdateUser())
	srv.Endpoint("user.delete", amqpEndpoints.DeleteUser())
	srv.Endpoint("user.list", amqpEndpoints.ListUser())

	srv.Endpoint("file.create", amqpEndpoints.CreateFile())
	srv.Endpoint("file.getById", amqpEndpoints.GetFileById())
	srv.Endpoint("file.update", amqpEndpoints.UpdateFile())
	srv.Endpoint("file.delete", amqpEndpoints.DeleteFile())
	srv.Endpoint("file.list", amqpEndpoints.ListFiles())

	err = srv.Start()

	if err != nil {
		return err
	}
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app := cli.NewApp()
	app.Name = "tik internal api"
	app.Description = ""
	app.Usage = "tik api run"
	app.UsageText = "tik api  run"
	app.Version = version
	app.Flags = flags
	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
