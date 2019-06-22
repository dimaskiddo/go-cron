package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/robfig/cron"

	hlp "github.com/dimaskiddo/go-cron/helper"
)

func main() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)

	hlp.LogPrint(hlp.LogLevelInfo, "Initialize Go Cron")
	var cronSchedule, cronCommand []string

	cronFile := hlp.GetEnv("CRON_FILE", "string", false)
	if cronFile == nil {
		cronFile = "/etc/gocron/crontab"
	}

	_, err := os.Stat(cronFile.(string))
	if err != nil {
		hlp.LogPrint(hlp.LogLevelWarn, "Cron File Doesn't Exist")
		hlp.LogPrint(hlp.LogLevelInfo, "Loading Cron Environment Variable")

		cronSchedule = strings.Split(hlp.GetEnv("CRON_SCHEDULE_LIST", "string", true).(string), ";")
		cronCommand = strings.Split(hlp.GetEnv("CRON_COMMAND_LIST", "string", true).(string), ";")
	} else {
		hlp.LogPrint(hlp.LogLevelInfo, "Cron File Exist")
		hlp.LogPrint(hlp.LogLevelInfo, "Loading Cron File Content")

		cronStream, err := ioutil.ReadFile(cronFile.(string))
		if err != nil {
			hlp.LogPrint(hlp.LogLevelFatal, "Error While Loading Cron File")
		}

		cronLine := strings.Split(string(cronStream), "\n")
		for i := 0; i < len(cronLine); i++ {
			cronContent := strings.SplitN(cronLine[i], " ", 6)

			cronSchedule = append(cronSchedule, cronContent[0]+" "+cronContent[1]+" "+cronContent[2]+" "+cronContent[3]+" "+cronContent[4])
			cronCommand = append(cronCommand, cronContent[5])
		}
	}

	hlp.LogPrint(hlp.LogLevelInfo, "Done Parsing Cron Schedule List and Command List")
	if len(cronSchedule) != len(cronCommand) {
		hlp.LogPrint(hlp.LogLevelFatal, "Cron Schedule List and Command List Has Different Total Length")
	}

	hlp.LogPrint(hlp.LogLevelInfo, "Initialize Cron Routine")
	cronRoutine := cron.New()

	for i := 0; i < len(cronSchedule); i++ {
		hlp.LogPrint(hlp.LogLevelInfo, "Adding Cron Routine [Cron ID: "+strconv.Itoa(i)+", Schedule: ["+cronSchedule[i]+"], Command: ["+cronCommand[i]+"]]")
		go func(i int) {
			cronRoutine.AddFunc("0 "+cronSchedule[i], func() {
				cronID := strconv.Itoa(i)
				hlp.LogPrint(hlp.LogLevelInfo, "Cron ID: "+cronID+", Executing Cron Command")

				cmdCron := hlp.SplitWithEscapeN(cronCommand[i], " ", -1, true)
				cmdExec := exec.Command(cmdCron[0], cmdCron[1:]...)

				var cmdStdout bytes.Buffer
				var cmdStderr bytes.Buffer

				cmdExec.Stdout = &cmdStdout
				cmdExec.Stderr = &cmdStderr

				err := cmdExec.Run()
				if err != nil {
					hlp.LogPrint(hlp.LogLevelError, "Cron ID: "+cronID+", Execution Error:")
					hlp.LogPrint(hlp.LogLevelError, "Cron ID: "+cronID+", -----------------------------------------")
					hlp.LogPrint(hlp.LogLevelError, "Cron ID: "+cronID+", "+string(cmdStderr.String()))
					hlp.LogPrint(hlp.LogLevelError, "Cron ID: "+cronID+", -----------------------------------------")
					hlp.LogPrint(hlp.LogLevelInfo, "Cron ID: "+cronID+", Cron Command Executed With Error")
					return
				}

				hlp.LogPrint(hlp.LogLevelInfo, "Cron ID: "+cronID+", Execution Result:")
				hlp.LogPrint(hlp.LogLevelInfo, "Cron ID: "+cronID+", -----------------------------------------")
				hlp.LogPrint(hlp.LogLevelInfo, "Cron ID: "+cronID+", "+string(cmdStdout.String()))
				hlp.LogPrint(hlp.LogLevelInfo, "Cron ID: "+cronID+", -----------------------------------------")
				hlp.LogPrint(hlp.LogLevelInfo, "Cron ID: "+cronID+", Cron Command Executed Successfully")
			})
		}(i)
	}

	hlp.LogPrint(hlp.LogLevelInfo, "Starting Cron Routine")
	cronRoutine.Start()

	select {
	case <-sigchan:
		fmt.Println("")
		hlp.LogPrint(hlp.LogLevelInfo, "Stopping Cron Routine")
		cronRoutine.Stop()

		hlp.LogPrint(hlp.LogLevelInfo, "Stopping Go Cron")
		os.Exit(0)
	}
}
