package bluetasks

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"blue-admin.com/configs"
	"blue-admin.com/database"
	"blue-admin.com/utils"

	"github.com/madflojo/tasks"
)

func ScheduledTasks() *tasks.Scheduler {

	//  initalizing scheduler for regullarly running tasks
	scheduler := tasks.New()

	// JWT signature salt will be updated based on the env variable provided
	// this is what the peice of code below does, using golangs task scheduler
	//  Salt Timer Tasks
	clear_run, _ := strconv.Atoi(configs.AppConfig.Get("JWT_SALT_LIFE_TIME"))
	clear_run = int(clear_run)
	jwt_update_interval := time.Minute * time.Duration(clear_run)
	//  Task 2 for testing Make random heartbeat call
	if _, err := scheduler.Add(&tasks.Task{
		Interval: jwt_update_interval,
		TaskFunc: func() error {
			utils.JWTSaltUpdate()
			return nil
		},
	}); err != nil {
		fmt.Println(err)
	}

	// // Add a task to move to Logs Directory Every Interval, Interval to Be Provided From Configuration File
	gormLoggerfile, _ := database.GormLoggerFile()
	//  App should not start
	log_file, _ := Logfile()
	if _, err := scheduler.Add(&tasks.Task{
		Interval: time.Duration(1 * time.Hour),
		TaskFunc: func() error {
			currentTime := time.Now()
			FileName := fmt.Sprintf("%v-%v-%v-%v-%v", currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), currentTime.Minute())
			Command := fmt.Sprintf("cp goblue.log logs/blue-%v.log", FileName)
			Command2 := fmt.Sprintf("cp blue-admin.log logs/gorm-%v.log", FileName)
			if _, err := exec.Command("bash", "-c", Command).Output(); err != nil {
				fmt.Printf("error: %v\n", err)
			}

			if _, err := exec.Command("bash", "-c", Command2).Output(); err != nil {
				fmt.Printf("error: %v\n", err)
			}
			gormLoggerfile.Truncate(0)
			log_file.Truncate(0)
			return nil
		},
	}); err != nil {
		fmt.Println(err)

	}

	return scheduler
}