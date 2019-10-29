package util

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gits/config"
)

var (
	maintenanceOn bool
)

func Maintenance() (bool, string) {
	on, message := maintenance()
	if config.Verbose {
		if maintenanceOn != on {
			onOff := "off"
			if on {
				onOff = "on"
			}
			log.Printf("Maintenance mode %s", onOff)
			maintenanceOn = on
		}
	}
	return on, message
}

func maintenance() (bool, string) {
	file := config.MaintenanceFile
	if file == "" {
		return false, ""
	}
	info, err := os.Stat(file)
	if err != nil || info == nil {
		if !strings.Contains(err.Error(), ": no such file or directory") {
			log.Printf("Unable to stat `%s`: %v", file, err)
		}
		return false, ""
	}
	msg := ""
	if info.Mode().IsRegular() && info.Size() > 0 {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("Unable to read `%s`: %v", file, err)
		} else if len(data) > 0 {
			msg = string(data)
		}
	}
	return true, msg
}

func Errors(sep string, maybeErrors ...error) string {
	errs := make([]string, 0, len(maybeErrors))
	for _, err := range maybeErrors {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return "(no errors)"
	}
	if sep == "" {
		sep = ", "
	}
	return strings.Join(errs, sep)
}
