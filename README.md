# IND Appointment helper

Telegram bot that helps with tracking open slots at IND in the Netherlands.

# Build

```shell
make build
```

# Run

```shell
TELEGRAM_API_KEY=${you_api_key} ./bot
```

or during development:
```shell
TELEGRAM_API_KEY=${you_api_key} make run
```

# Interactions with the bot

Find the bot by name in Telegram and initiate chat with it (command `/start` will be sent).
To start tracking a particular location execute command from chat:
```
/track
```

You will be prompted to select one of the available locations:
- IND Amsterdam
- IND Den Haag
- IND Zwolle
- IND Den Bosch

And number of people - 1 to 6.
Optionally if you are only interested in time windows before particular date you can specify it in YYYY-MM-DD
format, otherwise you can specify tracking all time slots.

After that you will receive notifications about open windows with mention of the first available window and number of
other possible options. Notifications do not start after the first one and might repeat the same information.

To stop tracking execute command:
```
/stoptrack
```
