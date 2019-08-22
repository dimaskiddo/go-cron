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

		cronTabFile, err := hlp.GetEnvString("CRON_CRONTAB_FILE")
		if err != nil {
			cronTabFile, err = cmd.Flags().GetString("file")
			if err != nil {
				hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
			}
		}

		hlp.LogPrintln(hlp.LogLevelInfo, "parse go-cron configuration")
		_, err = os.Stat(cronTabFile)
		if err != nil {
			hlp.LogPrintln(hlp.LogLevelWarn, "crontab file not found, load configuration from parameters")

			strSchedule, err = hlp.GetEnvString("CRON_SCHEDULE_LIST")
			if err != nil {
				strSchedule, err = cmd.Flags().GetString("cron-schedule")
				if err != nil {
					hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
				}
			}

			if len(strSchedule) <= 0 {
				hlp.LogPrintln(hlp.LogLevelFatal, "cron-schedule parameter is empty")
			}

			strCommand, err = hlp.GetEnvString("CRON_COMMAND_LIST")
			if err != nil {
				strCommand, err = cmd.Flags().GetString("cron-command")
				if err != nil {
					hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
				}
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

			cronTabLines := strings.Split(string(cronTabContents), fmt.Sprintf("\n"))
			for i := 0; i < len(cronTabLines); i++ {
				if len(cronTabLines[i]) != 0 {
					cronTabDatas := strings.Split(cronTabLines[i], " ")

					arrSchedule = append(arrSchedule, strings.TrimSpace(strings.Join(cronTabDatas[:5], " ")))
					arrCommand = append(arrCommand, strings.TrimSpace(strings.Join(cronTabDatas[5:], " ")))
				}
			}
		}

		if len(arrSchedule) != len(arrCommand) {
			hlp.LogPrintln(hlp.LogLevelFatal, fmt.Sprintf("cron-schedule and cron-command has mismatch range (%v:%v)", len(arrSchedule), len(arrCommand)))
		}

		strShowResultIDs, err = hlp.GetEnvString("CRON_SHOW_RESULT_IDS_LIST")
		if err != nil {
			strShowResultIDs, err = cmd.Flags().GetString("cron-show-ids")
			if err != nil {
				hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
			}
		}

		arrShowResultIDs = strings.Split(strShowResultIDs, ";")

		cronRoutines := cron.New(cron.WithChain(
			cron.Recover(cron.DefaultLogger),
		))

		for i := 0; i < len(arrSchedule); i++ {
			hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("initialize go-cron routine [id: %v, routine: [%v] => [%v]]", i, arrSchedule[i], arrCommand[i]))

			go func(i int) {
				cronRoutines.AddFunc(arrSchedule[i], func() {
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
							hlp.LogPrintln(hlp.LogLevelError, fmt.Sprintf("id: %v, execution result:\n%v", i, string(cronStderr.String())))
							hlp.LogPrintln(hlp.LogLevelError, fmt.Sprintf("id: %v, -----------------------------------------", i))
						}
						hlp.LogPrintln(hlp.LogLevelError, fmt.Sprintf("id: %v, cron routine executed with an error", i))
						return
					}

					if hlp.IsStringsContains(fmt.Sprintf("%v", i), arrShowResultIDs) {
						hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, cron routine execution result:", i))
						hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, -----------------------------------------", i))
						hlp.LogPrintln(hlp.LogLevelInfo, fmt.Sprintf("id: %v, execution result:\n%v", i, string(cronStdout.String())))
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
	Daemon.Flags().String("cron-command", "", "Cron command list to be executed when cron schedule is match with current time (use (;) separator between command). Can be override using CRON_COMMAND_LIST environment variable")
	Daemon.Flags().String("cron-schedule", "", "Cron schedule list to run cron command in cron time format (use (;) separator between schedule). Can be override using CRON_SCHEDULE_LIST environment variable")
	Daemon.Flags().String("cron-show-ids", "", "Cron IDs list to show cron command execution result (use (;) separator between ids). Can be override using CRON_SHOW_RESULT_IDS_LIST environment variable")
	Daemon.Flags().String("file", "/etc/go-cron/crontab", "Cron crontab file location. Can be override using CRON_CRONTAB_FILE environment variable")
}
