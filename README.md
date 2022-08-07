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
/track {am,dh,zw,db}
```

Where:
- am: IND Amsterdam
- dh: IND Den Haag
- zw: IND Zwolle
- db: IND Den Bosch

After that you will receive notifications about open windows with mention of the first available window and number of
other possible options. Notifications do not start after the first one and might repeat the same information.

To stop tracking execute command:
```
/stoptrack
```
