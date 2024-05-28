package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/ini.v1"
)

const configFile = "britto.ini"

func main() {
	xdgConfigDir := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigDir == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Failed to get current user: %v", err)
		}
		xdgConfigDir = filepath.Join(usr.HomeDir, ".config")
	}

	configPath := filepath.Join(xdgConfigDir, configFile)

	cfg, err := ini.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	now := time.Now()
	reminderDays := 10

	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection {
			continue
		}

		birthdayStr := section.Key("Date").String()
		var birthday time.Time

		var birthYear int
		var err error

		if len(birthdayStr) == 5 {
			// Format: DD/MM
			birthday, err = time.Parse("02/01", birthdayStr)
			if err != nil {
				log.Printf("Failed to parse birthday for [%s]: %v", section.Name(), err)
				continue
			}
			// Set the year to the current year
			birthday = birthday.AddDate(now.Year(), 0, 0)
			birthYear = now.Year()
		} else if len(birthdayStr) == 10 {
			// Format: DD/MM/YYYY
			birthday, err = time.Parse("02/01/2006", birthdayStr)
			if err != nil {
				log.Printf("Failed to parse birthday for [%s]: %v", section.Name(), err)
				continue
			}
			birthYear, err = strconv.Atoi(birthdayStr[6:])
			if err != nil {
				log.Printf("Failed to parse birth year for [%s]: %v", section.Name(), err)
				continue
			}
		} else {
			log.Printf("Invalid birthday format for [%s]", section.Name())
			continue
		}

		// If the birthday has already passed this year, set it to next year
		if birthday.Before(now) {
			birthday = birthday.AddDate(1, 0, 0)
			birthYear++
		}

		// Calculate the next birthday
		nextBirthday := time.Date(now.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, time.UTC)

		// If the next birthday is today
		if nextBirthday.Sub(now).Hours() < 24 && nextBirthday.Sub(now).Hours() >= 0 {
			if now.Year()-birthYear == 0 {
				fmt.Printf("[%s]'s birthday is today! %s\n", section.Name(), birthday.Format("02/01"))
			} else {
				fmt.Printf("[%s] is turning %d years old today! %s\n", section.Name(), now.Year()-birthYear, birthday.Format("02/01/2006"))
			}
		}

		// If the next birthday is tomorrow or after, print the number of days until the birthday
		daysUntilBirthday := int(nextBirthday.Sub(now).Hours() / 24)
		if daysUntilBirthday > 0 && daysUntilBirthday <= reminderDays {
			var due string
			if daysUntilBirthday == 0 {
				due = "today"
			} else if daysUntilBirthday == 1 {
				due = "tomorrow"
			} else {
				due = fmt.Sprintf("in %d days", daysUntilBirthday)
			}
			age := now.Year() - birthYear
			if age == 0 || age > 0 { // Print if the birth year is unknown (age == 0) or known (age > 0)
				if age == 0 {
					fmt.Printf("[%s]'s birthday is %s! %s\n", section.Name(), due, birthday.Format("02/01"))
				} else {
					fmt.Printf("[%s]'s birthday is %s! %s\n", section.Name(), due, birthday.Format("02/01/2006"))
				}
			}
		}
	}
}
