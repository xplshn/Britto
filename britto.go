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

func parseBirthday(birthdayStr string, now time.Time) (time.Time, int, error) {
	if birthdayStr == "" {
		return time.Time{}, 0, fmt.Errorf("birthday not provided")
	}

	var birthday time.Time
	var birthYear int
	var err error

	if len(birthdayStr) == 5 {
		birthday, err = time.Parse("02/01", birthdayStr)
		if err == nil {
			birthYear = now.Year()
		}
	} else if len(birthdayStr) == 10 {
		birthday, err = time.Parse("02/01/2006", birthdayStr)
		if err == nil {
			birthYear, err = strconv.Atoi(birthdayStr[6:])
			if err != nil {
				return time.Time{}, 0, fmt.Errorf("failed to parse birth year: %v", err)
			}
		}
	} else {
		return time.Time{}, 0, fmt.Errorf("invalid birthday format")
	}

	return birthday, birthYear, nil
}

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
	now := time.Now().Truncate(24 * time.Hour) // Truncate to remove the time component
	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection {
			continue
		}

		birthdayStr := section.Key("Date").String()
		if birthdayStr == "" {
			log.Printf("[%s]: Missing Birth date", section.Name())
			continue
		}

		birthday, birthYear, err := parseBirthday(birthdayStr, now)
		if err != nil {
			log.Printf("[%s]: Failed to parse birth date", section.Name())
			continue
		}

		if birthYear > now.Year() { // Invalid birth year
			log.Printf("[%s]: Invalid birth date, year is in the future", section.Name())
			continue
		}

		// Calculate the next birthday
		nextBirthday := time.Date(now.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, time.UTC)

		// If the next birthday is today, tomorrow or after, print the number of days until the birthday
		daysUntilBirthday := int(nextBirthday.Sub(now).Hours() / 24)
		if daysUntilBirthday <= 10 && daysUntilBirthday >= 0 || nextBirthday.Sub(now).Hours() < 24 && nextBirthday.Sub(now).Hours() >= 0 { // Remind if birthday is within 10 days
			var due string
			if daysUntilBirthday == 0 {
				due = "today"
			} else if daysUntilBirthday == 1 {
				due = "tomorrow"
			} else {
				due = fmt.Sprintf("in %d days", daysUntilBirthday)
			}
			age := now.Year() - birthYear
			if age == 0 || age > 0 {
				if age == 0 {
					fmt.Printf("[%s]'s birthday is %s! %s\n", section.Name(), due, birthday.Format("02/01"))
				} else {
					fmt.Printf("[%s] is turning %d years old %s! %s\n", section.Name(), age, due, birthday.Format("02/01/2006"))
				}
			}
		}
	}
}
