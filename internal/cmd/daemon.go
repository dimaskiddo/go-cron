package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	cron "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"

	"github.com/dimaskiddo/go-cron/pkg/env"
	"github.com/dimaskiddo/go-cron/pkg/log"
	"github.com/dimaskiddo/go-cron/pkg/str"
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

		log.Println(log.LogLevelInfo, "initialize go-cron")
		var strSchedule, strCommand, strShowResultIDs string
		var arrSchedule, arrCommand, arrShowResultIDs []string

		cronTabFile, err := env.GetEnvString("CRON_CRONTAB_FILE")
		if err != nil {
			cronTabFile, err = cmd.Flags().GetString("file")
			if err != nil {
				log.Println(log.LogLevelFatal, err.Error())
			}
		}

		log.Println(log.LogLevelInfo, "parse go-cron configuration")
		_, err = os.Stat(cronTabFile)
		if err != nil {
			log.Println(log.LogLevelWarn, "crontab file not found, load configuration from parameters")

			strSchedule, err = env.GetEnvString("CRON_SCHEDULE_LIST")
			if err != nil {
				strSchedule, err = cmd.Flags().GetString("cron-schedule")
				if err != nil {
					log.Println(log.LogLevelFatal, err.Error())
				}
			}

			if len(strSchedule) <= 0 {
				log.Println(log.LogLevelFatal, "cron-schedule parameter is empty")
			}

			strCommand, err = env.GetEnvString("CRON_COMMAND_LIST")
			if err != nil {
				strCommand, err = cmd.Flags().GetString("cron-command")
				if err != nil {
					log.Println(log.LogLevelFatal, err.Error())
				}
			}

			if len(strCommand) <= 0 {
				log.Println(log.LogLevelFatal, "cron-command parameter is empty")
			}

			arrSchedule = strings.Split(strSchedule, ";")
			arrCommand = strings.Split(strCommand, ";")
		} else {
			cronTabContents, err := ioutil.ReadFile(cronTabFile)
			if err != nil {
				log.Println(log.LogLevelFatal, err.Error())
			}

			if len(string(cronTabContents)) <= 0 {
				log.Println(log.LogLevelFatal, "crontab file has empty contents")
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
			log.Println(log.LogLevelFatal, fmt.Sprintf("cron-schedule and cron-command has mismatch range (%v:%v)", len(arrSchedule), len(arrCommand)))
		}

		strShowResultIDs, err = env.GetEnvString("CRON_SHOW_RESULT_IDS_LIST")
		if err != nil {
			strShowResultIDs, err = cmd.Flags().GetString("cron-show-ids")
			if err != nil {
				log.Println(log.LogLevelFatal, err.Error())
			}
		}

		arrShowResultIDs = strings.Split(strShowResultIDs, ";")

		cronRoutines := cron.New(cron.WithChain(
			cron.Recover(cron.DefaultLogger),
		))

		for i := 0; i < len(arrSchedule); i++ {
			log.Println(log.LogLevelInfo, fmt.Sprintf("initialize go-cron routine [id: %v, routine: [%v] => [%v]]", i, arrSchedule[i], arrCommand[i]))

			go func(i int) {
				cronRoutines.AddFunc(arrSchedule[i], func() {
					log.Println(log.LogLevelInfo, fmt.Sprintf("id: %v, executing cron routine", i))

					cronCommand := str.SplitWithEscapeN(arrCommand[i], " ", -1, true)
					cronExecute := exec.Command(cronCommand[0], cronCommand[1:]...)

					var cronStdout, cronStderr bytes.Buffer

					cronExecute.Stdout = &cronStdout
					cronExecute.Stderr = &cronStderr

					err := cronExecute.Run()
					if err != nil {
						if str.IsStringsContains(fmt.Sprintf("%v", i), arrShowResultIDs) {
							log.Println(log.LogLevelError, fmt.Sprintf("id: %v, cron routine execution result:", i))
							log.Println(log.LogLevelError, fmt.Sprintf("id: %v, -----------------------------------------", i))
							log.Println(log.LogLevelError, fmt.Sprintf("id: %v, execution result:\n%v", i, string(cronStderr.String())))
							log.Println(log.LogLevelError, fmt.Sprintf("id: %v, -----------------------------------------", i))
						}
						log.Println(log.LogLevelError, fmt.Sprintf("id: %v, cron routine executed with an error", i))
						return
					}

					if str.IsStringsContains(fmt.Sprintf("%v", i), arrShowResultIDs) {
						log.Println(log.LogLevelInfo, fmt.Sprintf("id: %v, cron routine execution result:", i))
						log.Println(log.LogLevelInfo, fmt.Sprintf("id: %v, -----------------------------------------", i))
						log.Println(log.LogLevelInfo, fmt.Sprintf("id: %v, execution result:\n%v", i, string(cronStdout.String())))
						log.Println(log.LogLevelInfo, fmt.Sprintf("id: %v, -----------------------------------------", i))
					}
					log.Println(log.LogLevelInfo, fmt.Sprintf("id: %v, cron routine executed successfully", i))
				})
			}(i)
		}

		log.Println(log.LogLevelInfo, "starting go-cron routine")
		cronRoutines.Start()

		select {
		case <-sig:
			fmt.Println("")
			log.Println(log.LogLevelInfo, "stopping go-cron routine")

			log.Println(log.LogLevelInfo, "stopping go-cron")
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
