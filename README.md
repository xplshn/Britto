# Britto

Britto is a configurable birthday reminder tool that keeps track of birthdays and reminds you of upcoming ones. It's designed to be simple, lightweight, and efficient.

## (non)Features
- Britto is small, it doesn't receive any CLi arguments
- There won't be any new features apart from simply reminding you of dates

## Installation

To use Britto, simply add it to your shell's startup configuration file.

## Configuration

Britto uses a simple configuration file (`britto.ini`) to store birthday information. Each person's birthday is added as a section in the config file, with the name of the person as the section header and the date of birth in the format `DD/MM/YYYY` or `DD/MM`.

Example `britto.ini`:

```ini
[John Doe]
Date=25/12/1985

[Alice Smith]
Date=10/05
```

## Usage
Britto automatically reminds you of upcoming birthdays every time you open your terminal or launch it. It checks the configuration file for upcoming birthdays and displays reminders if they are within a 10-day window. If no one's birthday is near, Britto will not output anything.

## License

This project is licensed under the 3BSD License - see the [LICENSE](LICENSE) file for details.
