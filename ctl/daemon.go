package ctl

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
	"github.com/spf13/cobra"

	"github.com/dimaskiddo/go-cron/hlp"
)

// Daemon Variable Structure
var Daemon = &cobra.Command{
	Use:   "daemon",
	Short: "Run Go-Cron as daemon",
	Long:  "Run Go-Cron as Deamon Service",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		hlp.LogPrintln(hlp.LogLevelInfo, "initialize go-cron")
		var strSchedule, strCommand, strShowResultIDs string
		var arrSchedule, arrCommand, arrShowResultIDs []string

		cronTabFile, err := cmd.Flags().GetString("file")
		if err != nil {
			hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
		}

		hlp.LogPrintln(hlp.LogLevelInfo, "parse go-cron configuration")
		_, err = os.Stat(cronTabFile)
		if err != nil {
			hlp.LogPrintln(hlp.LogLevelWarn, "crontab file not found, load configuration from parameters")

			strSchedule, err = cmd.Flags().GetString("cron-schedule")
			if err != nil {
				hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
			}

			if len(strSchedule) <= 0 {
				hlp.LogPrintln(hlp.LogLevelFatal, "cron-schedule parameter is empty")
			}

			strCommand, err = cmd.Flags().GetString("cron-command")
			if err != nil {
				hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
			}

			if len(strCommand) <= 0 {
				hlp.LogPrintln(hlp.LogLevelFatal, "cron-command parameter is empty")
			}

			arrSchedule = strings.Split(strSchedule, ";")
			arrCommand = strings.Split(strCommand, ";")
		} else {
			cronTabContents, err := ioutil.ReadFile(cronTabFile)
			if err != nil {
				hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
			}

			if len(string(cronTabContents)) <= 0 {
				hlp.LogPrintln(hlp.LogLevelFatal, "crontab file has empty contents")
			}

			cronTabLines := strings.Split(string(cronTabContents), "\n")
			for i := 0; i < len(cronTabLines); i++ {
				cronTabDatas := strings.Split(cronTabLines[i], " ")

				arrSchedule = append(arrSchedule, cronTabDatas[0], cronTabDatas[1], cronTabDatas[2], cronTabDatas[3], cronTabDatas[4])
				arrCommand = append(arrCommand, cronTabDatas[5])
			}
		}

		if len(arrSchedule) != len(arrCommand) {
			hlp.LogPrintln(hlp.LogLevelFatal, "cron-schedule and cron-command has mismatch range")
		}

		strShowResultIDs, err = cmd.Flags().GetString("cron-show-ids")
		if err != nil {
			hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
		}
		arrShowResultIDs = strings.Split(strShowResultIDs, ";")

		cronRoutines := cron.New()
		for i := 0; i < len(arrSchedule); i++ {
			hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("initialize go-cron routine [id: %v, routine: [%v] => [%v]]", i, arrSchedule[i], arrCommand[i]))

			go func(i int) {
				cronRoutines.AddFunc("0 "+arrSchedule[i], func() {
					hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, executing cron routine", i))

					cronCommand := hlp.SplitWithEscapeN(arrCommand[i], " ", -1, true)
					cronExecute := exec.Command(cronCommand[0], cronCommand[1:]...)

					var cronStdout, cronStderr bytes.Buffer

					cronExecute.Stdout = &cronStdout
					cronExecute.Stderr = &cronStderr

					err := cronExecute.Run()
					if err != nil {
						if hlp.IsStringsContains(fmt.Sprintf("%v", i), arrShowResultIDs) {
							hlp.LogPrintln(hlp.LogLevelError, fmt.Sprintf("id: %v, cron routine execution result:", i))
							hlp.LogPrintln(hlp.LogLevelError, fmt.Sprintf("id: %v, -----------------------------------------", i))
							hlp.LogPrintln(hlp.LogLevelError, fmt.Sprintf("id: %v, %v", i, string(cronStderr.String())))
							hlp.LogPrintln(hlp.LogLevelError, fmt.Sprintf("id: %v, -----------------------------------------", i))
						}
						hlp.LogPrintln(hlp.LogLevelError, fmt.Sprintf("id: %v, cron routine executed with an error", i))
						return
					}

					if hlp.IsStringsContains(fmt.Sprintf("%v", i), arrShowResultIDs) {
						hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, cron routine execution result:", i))
						hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, -----------------------------------------", i))
						hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, %v", i, string(cronStdout.String())))
						hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, -----------------------------------------", i))
					}
					hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, cron routine executed successfully", i))
				})
			}(i)
		}

		hlp.LogPrintln(hlp.LogLevelInfo, "starting go-cron routine")
		cronRoutines.Start()

		select {
		case <-sig:
			fmt.Println("")
			hlp.LogPrintln(hlp.LogLevelInfo, "stopping go-cron routine")

			hlp.LogPrintln(hlp.LogLevelInfo, "stopping go-cron")
			os.Exit(0)
		}
	},
}

func init() {
	Daemon.Flags().String("cron-command", "", "Cron command list to be executed when cron schedule is match with current time (use (;) separator between command)")
	Daemon.Flags().String("cron-schedule", "", "Cron schedule list to run cron command in cron time format (use (;) separator between schedule)")
	Daemon.Flags().String("cron-show-ids", "", "Cron ids list to show cron command execution result (use (;) separator between ids)")
	Daemon.Flags().String("file", "/etc/go-cron/crontab", "Cron crontab file location")
}