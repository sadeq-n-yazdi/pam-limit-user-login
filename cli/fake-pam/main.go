package main

import (
	"io"
	"os"
	"pam-limit-user-login/internal/logger"
)

func main() {
	file, err := os.OpenFile("/var/log/pam-ssh-limiter.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		multiWriter := io.MultiWriter(os.Stdout, file)
		logger.GetLogger().Out = multiWriter
	} else {
		logger.GetLogger().Warnln("Failed to log to file, using default stderr")
	}
	// Close the file when your app exits
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	logger.GetLogger().Infoln("Do not let to login")
	logger.GetLogger().Infoln("Args:", os.Args)
	logger.GetLogger().Infoln("Envs:", os.Environ())
	os.Exit(1)
}
