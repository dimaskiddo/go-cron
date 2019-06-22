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
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)

	hlp.LogPrint(hlp.LogLevelInfo, "Initialize Go Cron")
	var listSchedule, listCommand []string

	cronTab := "/etc/gocron/crontab"
	_, err := os.Stat(cronTab)
	if err != nil {
		hlp.LogPrint(hlp.LogLevelWarn, "Cron File Doesn't Exist")
		hlp.LogPrint(hlp.LogLevelInfo, "Loading Cron Environment Variable")

		listSchedule = strings.Split(hlp.GetEnv("CRON_SCHEDULE_LIST", "string", true).(string), ";")
		listCommand = strings.Split(hlp.GetEnv("CRON_COMMAND_LIST", "string", true).(string), ";")
	} else {
		hlp.LogPrint(hlp.LogLevelInfo, "Cron File Exist")
		hlp.LogPrint(hlp.LogLevelInfo, "Loading Cron File Content")

		cronFile, err := ioutil.ReadFile(cronTab)
		if err != nil {
			hlp.LogPrint(hlp.LogLevelFatal, "Error While Loading Cron File")
		}

		cronLine := strings.Split(string(cronFile), "\n")
		for i := 0; i < len(cronLine); i++ {
			cronContent := strings.SplitAfterN(cronLine[i], " ", 5)

			listSchedule = append(listSchedule, cronContent[0])
			listCommand = append(listCommand, cronContent[1])
		}
	}

	if len(listSchedule) != len(listCommand) {
		hlp.LogPrint(hlp.LogLevelFatal, "Cron Schedule List and Command List Has Different Total Length")
	}

	hlp.LogPrint(hlp.LogLevelInfo, "Initialize Cron Routine")
	cronRoutine := cron.New()

	for i := 0; i < len(listSchedule); i++ {
		hlp.LogPrint(hlp.LogLevelInfo, "Adding Cron Routine [Cron ID: "+strconv.Itoa(i)+", Schedule: ["+listSchedule[i]+"], Command: ["+listCommand[i]+"]]")
		go func(i int) {
			cronRoutine.AddFunc("0 "+listSchedule[i], func() {
				cronID := strconv.Itoa(i)
				hlp.LogPrint(hlp.LogLevelInfo, "Cron ID: "+cronID+", Executing Cron Command")

				cmdCron := hlp.SplitWithEscapeN(listCommand[i], " ", -1, true)
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
	case <-osSignal:
		fmt.Println("")
		hlp.LogPrint(hlp.LogLevelInfo, "Stopping Cron Routine")
		cronRoutine.Stop()

		hlp.LogPrint(hlp.LogLevelInfo, "Stopping Go Cron")
		os.Exit(0)
	}
}
