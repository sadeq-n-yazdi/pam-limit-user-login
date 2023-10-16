package utilities

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"pam-limit-user-login/internal/configurations"
	"pam-limit-user-login/internal/logger"
	"regexp"
	"strconv"
	"strings"
)

type (
	UserIDToUserNameMap map[int]string
)

var userIdMapToUserName UserIDToUserNameMap

func CountSshProcesses(userName string, userId int, userUserName bool) (int, error) {
	logger.GetLogger().Debugln("CountSshProcesses(", userName, " string, ", userId, " int)")
	// Read the contents of /proc directory to get process information
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return 0, err
	}

	count := 0

	if userId <= 0 && userUserName {
		logger.GetLogger().Debugln("Going to extract userId from user name")
		userId, err = GetUserID(userName)
		if err != nil {
			return 0, fmt.Errorf("can not get user id for %s", userName)
		}
	}
	logger.GetLogger().Debugln("userId to check is ", userId)

	userIdString := strconv.FormatInt(int64(userId), 10)
	userMatcher := regexp.MustCompile(`(?im)^Uid:\s+` + userIdString + `\s`)

	// Iterate through the /proc directory entries
	for _, entry := range entries {
		// Check if the entry is a directory (a process)
		if entry.IsDir() {
			// Check if the directory name is a number (process ID)
			if pid, err := strconv.Atoi(entry.Name()); err == nil {
				//logger.GetLogger().Debugln("Checking pid ", pid)
				// Read the process status file to get process information
				statusFilePath := fmt.Sprintf("%s/%s/status", procDir, entry.Name())
				statusFileContents, err := os.ReadFile(statusFilePath)
				if err != nil {
					logger.GetLogger().Warnf("[%v] happend. go for next entry\n", err)
					continue
				}

				// Check if the status file contains "sshd", and the user matches.
				statusString := string(statusFileContents)
				//fmt.Println("Status:", statusString)

				if strings.Contains(statusString, "sshd") && userMatcher.MatchString(statusString) {
					count++
					logger.GetLogger().Debugln("owner of PID ", pid, "matches to the user")
				}
			}
		}
	}
	logger.GetLogger().Debugf("Uid: %s[%s] has %d sshd process\n", userName, userIdString, count)
	return count, nil
}

func GetUserID(username string) (int, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return -1, err
	}
	uid, _ := strconv.Atoi(u.Uid)
	return uid, nil
}

func ReturnExitCode(config configurations.ConfigurationMap, code int) {
	doLog, okLog := config["debug"].(bool)
	if okLog && doLog {
		logger.GetLogger().Debugln("Exit return code: ", code)
	}
	pseudo, okPseudo := config["pseudo"].(bool)
	if okPseudo && pseudo {
		os.Exit(0)
	} else {
		os.Exit(code)
	}
}

func getUsersIdToUserNameMap(force bool) UserIDToUserNameMap {
	// Open the passwd file.
	if len(userIdMapToUserName) > 0 && !force {
		return userIdMapToUserName
	}

	f, err := os.Open("/etc/passwd")
	if err != nil {
		logger.GetLogger().Errorln("can not open passwd file", err)
		return UserIDToUserNameMap{
			0:    "root",
			1000: "sadeq",
		}
	}
	// Close the passwd file before return.
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.GetLogger().Errorln("can not close passwd file")
		}
	}(f)

	// Create a map of user IDs to usernames.
	userMap := make(UserIDToUserNameMap)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Split the line into the user ID and username.
		line := scanner.Text()
		fields := strings.Split(line, ":")
		userID, err := strconv.Atoi(fields[2])
		if err != nil {
			logger.GetLogger().Warnln(err)
			continue
		}
		userName := fields[0]

		// Add the user to the map.
		userMap[userID] = userName
	}
	userIdMapToUserName = userMap
	return userMap

}

func GetUserNameFromUserId(userId int, force bool) (userName string, found bool) {
	m := getUsersIdToUserNameMap(force)
	userName, found = m[userId]
	return
}
