# Britto

Britto is a small and configurable reminder tool for events and birthdays. It's designed to be simple, lightweight, and efficient.

## Installation

To use Britto, simply add it to your shell's startup configuration file.

## Configuration

Britto uses a simple configuration file (`britto.toml`) to store event and birthday information. Each reminder is added as an entry in the config file, with details including the name, date, and an optional custom message.

Example `britto.toml`:

```toml
[[Birthday]]
  Name = "John Doe"
  Date = "25/12/1985"

[[Birthday]]
  Name = "Alice Smith"
  Date = "10/05"

[[Reminder]]
  Name = "Project Deadline"
  Date = "15/07"
  Message = "Submit the project report"

[[Reminder]]
  Name = "Meeting"
  Date = "20/07/2024"
  OneTimeEvent = true

[ReminderRange]
  Birthdays = 10
  Events = 15
```

## Usage
Britto automatically reminds you of upcoming events, such as birthdays, every time you open your terminal or launch it. It checks the configuration file for upcoming events and displays reminders if they are within a configurable 10-day window. If there aren't any events nearing the day, Britto will not output anything.

## License

This project is licensed under the RABRMS License - see the [LICENSE](LICENSE) file for details. A flexible and non-copylefted license.
