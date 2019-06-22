package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/robfig/cron"

	hlp "github.com/dimaskiddo/go-cron/helper"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	hlp.LogPrint(hlp.LogLevelInfo, "Initialize Go Cron")
	var cronSchedule, cronCommand []string

	cronFile := hlp.GetEnv("CRON_FILE", "string", false)
	if cronFile == nil {
		cronFile = "/etc/gocron/crontab"
	}

	hlp.LogPrint(hlp.LogLevelInfo, "Parse Go Cron Variable")
	_, err := os.Stat(cronFile.(string))
	if err != nil {
		cronSchedule = strings.Split(hlp.GetEnv("CRON_SCHEDULE_LIST", "string", true).(string), ";")
		cronCommand = strings.Split(hlp.GetEnv("CRON_COMMAND_LIST", "string", true).(string), ";")
	} else {
		cronStream, err := ioutil.ReadFile(cronFile.(string))
		if err != nil {
			hlp.LogPrint(hlp.LogLevelFatal, "Error While Streaming Cron File")
		}

		cronLine := strings.Split(string(cronStream), "\n")
		for i := 0; i < len(cronLine); i++ {
			cronContent := strings.SplitN(cronLine[i], " ", 6)

			cronSchedule = append(cronSchedule, strings.TrimSpace(cronContent[0]+" "+cronContent[1]+" "+cronContent[2]+" "+cronContent[3]+" "+cronContent[4]))
			cronCommand = append(cronCommand, strings.TrimSpace(cronContent[5]))
		}
	}

	if len(cronSchedule) != len(cronCommand) {
		hlp.LogPrint(hlp.LogLevelFatal, "Error Go Cron Schedule and Command Variable Has Different Array Length")
	}

	hlp.LogPrint(hlp.LogLevelInfo, "Initialize Go Cron Routine")
	cronRoutine := cron.New()

	for i := 0; i < len(cronSchedule); i++ {
		hlp.LogPrint(hlp.LogLevelInfo, fmt.Sprintf("Add Go Cron Routine [Go Cron ID: %v, Routine: [%v] => [%v]]", i, cronSchedule[i], cronCommand[i]))

		go func(i int) {
			cronRoutine.AddFunc("0 "+cronSchedule[i], func() {
				hlp.LogPrint(hlp.LogLevelInfo, fmt.Sprintf("Go Cron ID: %v, Executing Command", i))

				cmdCron := hlp.SplitWithEscapeN(cronCommand[i], " ", -1, true)
				cmdExec := exec.Command(cmdCron[0], cmdCron[1:]...)

				var cmdStdout bytes.Buffer
				var cmdStderr bytes.Buffer

				cmdExec.Stdout = &cmdStdout
				cmdExec.Stderr = &cmdStderr

				err := cmdExec.Run()
				if err != nil {
					hlp.LogPrint(hlp.LogLevelError, fmt.Sprintf("Go Cron ID: %v, Execution Error:", i))
					hlp.LogPrint(hlp.LogLevelError, fmt.Sprintf("Go Cron ID: %v, -----------------------------------------", i))
					hlp.LogPrint(hlp.LogLevelError, fmt.Sprintf("Go Cron ID: %v, %v", i, string(cmdStderr.String())))
					hlp.LogPrint(hlp.LogLevelError, fmt.Sprintf("Go Cron ID: %v, -----------------------------------------", i))
					hlp.LogPrint(hlp.LogLevelError, fmt.Sprintf("Go Cron ID: %v, Command Executed With Error", i))
					return
				}

				hlp.LogPrint(hlp.LogLevelInfo, fmt.Sprintf("Go Cron ID: %v, Execution Result:", i))
				hlp.LogPrint(hlp.LogLevelInfo, fmt.Sprintf("Go Cron ID: %v, -----------------------------------------", i))
				hlp.LogPrint(hlp.LogLevelInfo, fmt.Sprintf("Go Cron ID: %v, %v", i, string(cmdStdout.String())))
				hlp.LogPrint(hlp.LogLevelInfo, fmt.Sprintf("Go Cron ID: %v, -----------------------------------------", i))
				hlp.LogPrint(hlp.LogLevelInfo, fmt.Sprintf("Go Cron ID: %v, Command Executed Successfully", i))
			})
		}(i)
	}

	hlp.LogPrint(hlp.LogLevelInfo, "Starting Go Cron Routine")
	cronRoutine.Start()

	select {
	case <-sig:
		fmt.Println("")
		hlp.LogPrint(hlp.LogLevelInfo, "Stopping Go Cron Routine")
		cronRoutine.Stop()

		hlp.LogPrint(hlp.LogLevelInfo, "Stopping Go Cron")
		os.Exit(0)
	}
}
