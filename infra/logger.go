package infra

import (
	"fmt"
	"os"

	"github.com/badico-cloud-hub/log-driver/logger"
	"github.com/badico-cloud-hub/log-driver/producer"
)

func NewLogManager() *producer.LoggerManager {

	logManager := producer.NewLoggerManager(logger.LogContext{
		AppName:    "Pubsubb API",
		AppType:    "API",
		AppVersion: "BEST BE FROM GIT",
		Machine:    os.Getenv("MACHINE_IP"),
	})

	err := logManager.Setup(
		os.Getenv("LOG_ACCESS_KEY"),
		os.Getenv("LOG_ACCESS_KEY_ID"),
		os.Getenv("LOG_QUEUE_URL"),
	)

	if err != nil {
		fmt.Println("============================================")
		fmt.Println("WARN: CLOUDLOG ENGINE NOT WORKING")
		fmt.Println("============================================")
		fmt.Println(err)
		return nil
	}

	return logManager

}
