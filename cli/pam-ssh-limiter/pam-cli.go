package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"pam-limit-user-login/internal/logger"
	"strings"

	"pam-limit-user-login/internal/configurations"
	"pam-limit-user-login/internal/utilities"
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
	// Replace with your configuration filename extensions: json, jsonc, yaml, yml
	configFilename := "/etc/pam-ssh-limiter/config.yaml"

	cfgFile := os.Getenv("PAM_SSH_LIMITER_CONFIG")
	if cfgFile != "" {
		configFilename = cfgFile
	}

	var muteLogs bool = false
	muteEnvVar := os.Getenv("PAM_SSH_LIMITER_MUTE")
	muteLogs = muteEnvVar != "" && strings.ToLower(muteEnvVar[:1]) == "y"
	if muteLogs {
		logger.GetLogger().SetLevel(logrus.FatalLevel)
	}

	config := &configurations.Config{}
	err = configurations.LoadConfigFile(configFilename, config)
	if err != nil {
		logger.GetLogger().Warnf("Oops! Unknown error happened to get config. I will let you continue with default setting! [%s][%v]\n", configFilename, err)
	} else {
		// Assign values to the global variables
		configurations.AdminUsers = config.AdminUsers
		configurations.UsersLimit = config.UsersLimit
		configurations.PamConfig = config.PamConfig
	}
	if !muteLogs && configurations.PamConfig["debug"].(bool) {
		logger.GetLogger().SetLevel(logrus.DebugLevel)
	}
	if configurations.PamConfig["pseudo"].(bool) {
		logger.GetLogger().Warnln("pseudo mode is active, every one is authorized")
	}

	// Get all parameters
	parameters := make(map[int]string)

	for i, arg := range os.Args {
		parameters[i] = arg
	}

	// Ensure at least six parameters
	for i := 0; i < 6; i++ {
		if _, exists := parameters[i]; !exists {
			parameters[i] = ""
		}
	}

	logger.GetLogger().Debugln(parameters)

	username := os.Getenv("PAM_USER")
	systemNameIsSSH := os.Getenv("PAM_SERVICE") == "sshd"
	if !systemNameIsSSH {
		logger.GetLogger().Infoln("Not sshd service", os.Getenv("PAM_SERVICE"))
		os.Exit(0)
	}
	var userId int
	userId, err = utilities.GetUserID(username)
	logger.GetLogger().Debugln("GetUserID:", username, userId, err)

	if userId < 512 || configurations.AdminUsers.IsInList(username) {
		if userId < 512 {
			logger.GetLogger().Debugln("User is below 512")
		}
		logger.GetLogger().Infoln("Access granted to special users. You can log in.")
		utilities.ReturnExitCode(config.PamConfig, 0)
	}
	serviceName := os.Getenv("PAM_SERVICE")
	if !strings.Contains(strings.ToLower(serviceName), "sshd") {
		logger.GetLogger().Infoln("This is not ssh login so access granted")
		utilities.ReturnExitCode(config.PamConfig, 0)
	}

	currentSessions, err := utilities.CountSshProcesses(username, userId, true)
	if err != nil {
		// ignore counting user processes
		logger.GetLogger().Errorf("Oops! an error happened. I will let you log in for now! [%v]", err)
		utilities.ReturnExitCode(config.PamConfig, 0)
	}

	if currentSessions >= configurations.UsersLimit.GetUserLimits(username) {
		logger.GetLogger().Infoln("User reached the maximum ")
		utilities.ReturnExitCode(config.PamConfig, 1)
	}

}
