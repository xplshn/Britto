package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
)

const defaultConfigFile = "britto.toml"

type Reminder struct {
	Name         string `toml:"Name"`
	Date         string `toml:"Date"`
	Message      string `toml:"Message,omitempty"`
	OneTimeEvent bool   `toml:"OneTimeEvent,omitempty"`
}

type ReminderRange struct {
	Birthdays int `toml:"Birthdays"`
	Events    int `toml:"Events"`
}

type Config struct {
	Birthdays     []Reminder    `toml:"Birthdays"`
	Reminders     []Reminder    `toml:"Reminders"`
	ReminderRange ReminderRange `toml:"ReminderRange"`
}

var defaultConfig = Config{
	Birthdays: []Reminder{
		{
			Name: "Example Person",
			Date: "01/01/2000",
		},
		{
			Name:    "Example Person 2",
			Date:    "07/01/2000",
			Message: "Example Person's birthday is on 07/01/2000. Remember to buy a present",
		},
	},
	Reminders: []Reminder{
		{
			Name:    "Example Event",
			Date:    "12/31",
			Message: "Don't forget about the Example Event!",
		},
		{
			Name:         "Example Event 2",
			Date:         "12/31/2024",
			OneTimeEvent: true,
		},
	},
	ReminderRange: ReminderRange{
		Birthdays: 10,
		Events:    15,
	},
}

func loadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if _, err := toml.DecodeReader(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveDefaultConfig(configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(defaultConfig); err != nil {
		return err
	}

	return nil
}

func parseDate(dateStr string, now time.Time, oneTimeEvent bool) (time.Time, int, error) {
	if dateStr == "" {
		return time.Time{}, 0, fmt.Errorf("date not provided")
	}

	var date time.Time
	var err error
	var year int

	if len(dateStr) == 5 {
		// Parse date without year
		date, err = time.Parse("02/01", dateStr)
		if err == nil {
			year = now.Year()
			date = time.Date(year, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		}
	} else if len(dateStr) == 10 {
		// Parse date with year
		date, err = time.Parse("02/01/2006", dateStr)
		if err == nil {
			year, err = strconv.Atoi(dateStr[6:])
			if err != nil {
				return time.Time{}, 0, fmt.Errorf("failed to parse year: %v", err)
			}
		}
	} else {
		return time.Time{}, 0, fmt.Errorf("invalid date format")
	}

	if oneTimeEvent && len(dateStr) != 10 {
		return time.Time{}, 0, fmt.Errorf("one-time event requires year specification")
	}

	return date, year, nil
}

func processReminders(reminders []Reminder, now time.Time, defaultMsg string, isBirthday bool, rangeDays int) {
	for _, reminder := range reminders {
		date, year, err := parseDate(reminder.Date, now, reminder.OneTimeEvent)
		if err != nil {
			log.Printf("[%s]: Failed to parse date: %v", reminder.Name, err)
			continue
		}

		nextDate := time.Date(now.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		if nextDate.Before(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		daysUntilDate := int(nextDate.Sub(now).Hours() / 24)

		// Check if the event falls within the reminder range, including crossing into the next year
		if daysUntilDate <= rangeDays && daysUntilDate >= 0 {
			var due string
			if daysUntilDate == 0 {
				due = "today"
			} else if daysUntilDate == 1 {
				due = "tomorrow"
			} else {
				due = fmt.Sprintf("in %d days", daysUntilDate)
			}

			msg := reminder.Message
			if msg == "" {
				msg = defaultMsg
			}
			if isBirthday {
				age := now.Year() - year
				if age == 0 {
					fmt.Printf("[%s]'s birthday is %s! %s\n", reminder.Name, due, date.Format("02/01"))
				} else {
					fmt.Printf("[%s] is turning %d years old %s! %s\n", reminder.Name, age, due, date.Format("02/01/2006"))
				}
			} else {
				fmt.Printf("[%s] is due %s! %s - %s\n", reminder.Name, due, date.Format("02/01"), msg)
			}
		}

		// Handle the case where the event falls into the next year within the range
		if daysUntilDate < 0 {
			nextDate = nextDate.AddDate(1, 0, 0)
			daysUntilDate = int(nextDate.Sub(now).Hours() / 24)

			if daysUntilDate <= rangeDays && daysUntilDate >= 0 {
				var due string
				if daysUntilDate == 0 {
					due = "today"
				} else if daysUntilDate == 1 {
					due = "tomorrow"
				} else {
					due = fmt.Sprintf("in %d days", daysUntilDate)
				}

				msg := reminder.Message
				if msg == "" {
					msg = defaultMsg
				}
				if isBirthday {
					age := now.Year() + 1 - year
					fmt.Printf("[%s] is turning %d years old %s! %s\n", reminder.Name, age, due, date.Format("02/01/2006"))
				} else {
					fmt.Printf("[%s] is due %s! %s - %s\n", reminder.Name, due, date.Format("02/01"), msg)
				}
			}
		}
	}
}

func main() {
	configPathFlag := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	configPath := *configPathFlag
	if configPath == "" {
		xdgConfigDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatalf("Failed to get user config directory: %v", err)
		}
		configPath = filepath.Join(xdgConfigDir, defaultConfigFile)
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Printf("Failed to load config file: %v", err)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Printf("Config file does not exist. Creating a default config.")
			err = saveDefaultConfig(configPath)
			if err != nil {
				log.Fatalf("Failed to save default config: %v", err)
			}
			log.Printf("Default config saved to %s. Please edit it with your reminders.", configPath)
		} else {
			log.Fatalf("Config file exists but cannot be loaded. Please check the file format and contents.")
		}
		return
	}

	now := time.Now().Truncate(24 * time.Hour) // Truncate to remove the time component

	// Process birthday reminders
	processReminders(config.Birthdays, now, "Birthday reminder", true, config.ReminderRange.Birthdays)
	// Process other reminders
	processReminders(config.Reminders, now, "Reminder", false, config.ReminderRange.Events)
}
