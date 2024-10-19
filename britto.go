package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
)

const defaultConfigFile = "britto.toml"

type Reminder struct {
	Name          string `toml:"Name"`
	Date          string `toml:"Date"`
	Message       string `toml:"Message,omitempty"`
	OneTimeEvent  bool   `toml:"OneTimeEvent,omitempty"`
	ReminderRange *int   `toml:"ReminderRange,omitempty"`
}

type ReminderRange struct {
	Birthdays int `toml:"Birthdays"`
	Events    int `toml:"Events"`
}

type TemplateConfig struct {
	DueToday        string `toml:"due_today"`
	DueTomorrow     string `toml:"due_tomorrow"`
	DueIn           string `toml:"due_in"`
	DateFormatShort string `toml:"date_format_short"`
	DateFormat      string `toml:"date_format"`
	Birthday0       string `toml:"Birthday0"`
	Birthday        string `toml:"Birthday"`
	Reminder        string `toml:"Reminder"`
}

type Config struct {
	Birthdays     []Reminder     `toml:"Birthday"`
	Reminders     []Reminder     `toml:"Reminder"`
	ReminderRange ReminderRange  `toml:"ReminderRange"`
	Template      TemplateConfig `toml:"template"`
}

var defaultTemplate = TemplateConfig{
	DueToday:        "today",
	DueTomorrow:     "tomorrow",
	DueIn:           "in {{.AgeOrDays}} days", // This refers to the number of days for reminders
	DateFormatShort: "01/06",
	DateFormat:      "02/01/2006",
	Birthday0:       "[{{.Name}}]'s birthday is {{.Due}}! {{.Date}}\n", // Use Name and Date
	Birthday:        "[{{.Name}}] is turning {{.AgeOrDays}} years old {{.Due}}! {{.Date}}",
	Reminder:        "[{{.Name}}] is due {{.Due}}! {{.Date}} - {{.Due}}",
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
	Template: defaultTemplate,
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

func saveDefaultConfig(configDir, configPath string) error {
	// Ensure the directory exists
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Create the default config file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(defaultConfig); err != nil {
		return fmt.Errorf("failed to encode default config: %v", err)
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

func processReminders(reminders []Reminder, now time.Time, defaultMsg string, isBirthday bool, defaultRange int, templateCfg TemplateConfig) {
	for _, reminder := range reminders {
		date, year, err := parseDate(reminder.Date, now, reminder.OneTimeEvent)
		if err != nil {
			log.Printf("[%s]: Failed to parse date: %v", reminder.Name, err)
			continue
		}

		// Override the global range if specified
		rangeDays := defaultRange
		if reminder.ReminderRange != nil {
			rangeDays = *reminder.ReminderRange
		}

		printReminder := func(daysUntilDate int, nextDate time.Time, year int) {
			var due string
			if daysUntilDate == 0 {
				due = templateCfg.DueToday
			} else if daysUntilDate == 1 {
				due = templateCfg.DueTomorrow
			} else {
				due = strings.ReplaceAll(templateCfg.DueIn, "{{.AgeOrDays}}", strconv.Itoa(daysUntilDate))
			}

			msg := reminder.Message
			if msg == "" {
				msg = defaultMsg
			}

			if isBirthday {
				age := nextDate.Year() - year
				tmpl := templateCfg.Birthday
				if age == 0 {
					tmpl = templateCfg.Birthday0
				}
				formattedMsg := formatTemplate(tmpl, reminder.Name, strconv.Itoa(age), due, nextDate.Format(templateCfg.DateFormat))
				fmt.Println(formattedMsg)
			} else {
				formattedMsg := formatTemplate(templateCfg.Reminder, reminder.Name, strconv.Itoa(daysUntilDate), due, nextDate.Format(templateCfg.DateFormatShort))
				fmt.Println(formattedMsg)
			}

			// Print the message if it exists
			if msg != "" {
				fmt.Println(msg)
			}
		}

		for _, yearsAhead := range []int{0, 1} {
			nextDate := time.Date(now.Year()+yearsAhead, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
			if nextDate.Before(now) {
				continue
			}

			daysUntilDate := int(nextDate.Sub(now).Hours() / 24)
			if daysUntilDate <= rangeDays && daysUntilDate >= 0 {
				printReminder(daysUntilDate, nextDate, year)
				break
			}
		}
	}
}

func formatTemplate(tmplStr string, name string, ageOrDays string, due string, formattedDate string) string {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := map[string]string{
		"Name":      name,
		"AgeOrDays": ageOrDays, // Depending on whether it's a birthday or event
		"Due":       due,
		"Date":      formattedDate,
	}
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	return buf.String()
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

		configDir := filepath.Join(xdgConfigDir, "britto")
		configPath = filepath.Join(configDir, defaultConfigFile)

		// If the config file doesn't exist, create the default config and directory
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Printf("Config file does not exist. Creating a default config.")
			err := saveDefaultConfig(configDir, configPath)
			if err != nil {
				log.Fatalf("Failed to save default config: %v", err)
			}
			log.Printf("Default config saved to %s. Please edit it with your reminders.", configPath)
		}
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	now := time.Now().Truncate(24 * time.Hour) // Truncate to remove the time component

	// Process birthday reminders
	processReminders(config.Birthdays, now, "Birthday reminder", true, config.ReminderRange.Birthdays, config.Template)
	// Process other reminders
	processReminders(config.Reminders, now, "Reminder", false, config.ReminderRange.Events, config.Template)
}
