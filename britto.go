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
	Birthday0       string `toml:"Birthday0"`
	Birthday        string `toml:"Birthday"`
	Reminder        string `toml:"Reminder"`
	DateFormat      string `toml:"DateFormat"`
	DateFormatShort string `toml:"DateFormatShort"`
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
	DueIn:           "in {{.AgeOrDays}} days",
	Birthday0:       "[{{.Name}}]'s birthday is {{.Due}}! {{.Date}}\n{{.Message}}",
	Birthday:        "[{{.Name}}] is turning {{.AgeOrDays}} years old {{.Due}}! {{.Date}}\n{{.Message}}",
	Reminder:        "[{{.Name}}] is due {{.Due}}! {{.Date}}\n{{.Message}}",
	DateFormat:      "DD/MM/YYYY",
	DateFormatShort: "DD/MM",
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
			Message: "Remember to buy a present",
		},
	},
	Reminders: []Reminder{
		{
			Name:    "Example Event",
			Date:    "12/31",
			Message: "Don't forget about the Example Event!",
		},
		{
			Name:          "Example Event 2",
			Date:          "12/31/2024",
			ReminderRange: intPtr(25),
		},
	},
	ReminderRange: ReminderRange{
		Birthdays: 10,
		Events:    15,
	},
	Template: defaultTemplate,
}

func loadConfig(configDir string) (*Config, error) {
	files, err := filepath.Glob(filepath.Join(configDir, "*.toml"))
	if err != nil {
		return nil, fmt.Errorf("failed to list toml files: %v", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no toml files found in directory: %s", configDir)
	}

	var config Config
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file %s: %v", file, err)
		}
		defer f.Close()

		if _, err := toml.DecodeReader(f, &config); err != nil {
			return nil, fmt.Errorf("failed to decode toml file %s: %v", file, err)
		}
	}

	return &config, nil
}

func saveDefaultConfig(configDir, configPath string) error {
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

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

func convertDateFormat(dateStr string) string {
	replacements := map[string]string{
		"MM":   "01",
		"DD":   "02",
		"YYYY": "2006",
		"YY":   "06",
	}

	for key, value := range replacements {
		dateStr = strings.ReplaceAll(dateStr, key, value)
	}

	return dateStr
}

func parseDate(dateStr string, now time.Time) (time.Time, int, error) {
	if dateStr == "" {
		return time.Time{}, 0, fmt.Errorf("date not provided")
	}

	var date time.Time
	var err error
	var year int

	if len(dateStr) == 5 {
		// Parse date without year (MM/DD)
		date, err = time.Parse("02/01", dateStr)
		if err == nil {
			year = now.Year()
			date = time.Date(year, date.Month(), date.Day(), 0, 0, 0, 0, now.Location())

			// If the parsed date has passed this year, set it to next year
			if date.Before(now) {
				year++
				date = time.Date(year, date.Month(), date.Day(), 0, 0, 0, 0, now.Location())
			}
		}
	} else if len(dateStr) == 10 {
		// Parse date with year (MM/DD/YYYY)
		date, err = time.Parse("02/01/2006", dateStr)
		if err == nil {
			year = date.Year()
		}
	} else {
		return time.Time{}, 0, fmt.Errorf("invalid date format")
	}

	return date, year, nil
}

func processReminders(reminders []Reminder, now time.Time, isBirthday bool, defaultRange int, templateCfg TemplateConfig) {
	for _, reminder := range reminders {
		date, year, err := parseDate(reminder.Date, now)
		if err != nil {
			log.Printf("[%s]: Failed to parse date: %v", reminder.Name, err)
			continue
		}

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

			var formattedMsg string
			if isBirthday {
				age := nextDate.Year() - year
				tmpl := templateCfg.Birthday
				if age == 0 {
					tmpl = templateCfg.Birthday0
				}
				formattedMsg = formatTemplate(tmpl, reminder.Name, strconv.Itoa(age), due, nextDate.Format(templateCfg.DateFormat), reminder.Message)
			} else {
				formattedMsg = formatTemplate(templateCfg.Reminder, reminder.Name, strconv.Itoa(daysUntilDate), due, nextDate.Format(templateCfg.DateFormat), reminder.Message)
			}

			fmt.Println(formattedMsg) // Print the final formatted message
		}

		for _, yearsAhead := range []int{0, 1} {
			nextDate := time.Date(now.Year()+yearsAhead, date.Month(), date.Day(), 0, 0, 0, 0, now.Location())

			daysUntilDate := int(nextDate.Sub(now).Hours() / 24)
			if daysUntilDate <= rangeDays && daysUntilDate >= 0 {
				printReminder(daysUntilDate, nextDate, year)
				break
			}
		}
	}
}

func formatTemplate(tmplStr string, name string, ageOrDays string, due string, formattedDate string, message string) string {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := map[string]string{
		"Name":      name,
		"AgeOrDays": ageOrDays,
		"Due":       due,
		"Date":      formattedDate,
		"Message":   message,
	}
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	return buf.String()
}

func intPtr(i int) *int {
	return &i
}

func main() {
	configDir := flag.String("config", os.Getenv("HOME")+"/.config/britto", "Directory where the config is stored")
	flag.Parse()

	configPath := filepath.Join(*configDir, defaultConfigFile)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := saveDefaultConfig(*configDir, configPath); err != nil {
			log.Fatalf("Failed to create default config file: %v", err)
		}
	}

	config, err := loadConfig(*configDir)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	now := time.Now()

	config.Template.DateFormat = convertDateFormat(config.Template.DateFormat)
	config.Template.DateFormatShort = convertDateFormat(config.Template.DateFormatShort)

	processReminders(config.Birthdays, now, true, config.ReminderRange.Birthdays, config.Template)
	processReminders(config.Reminders, now, false, config.ReminderRange.Events, config.Template)
}
