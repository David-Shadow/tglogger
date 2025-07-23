[Telegram Logger](https://github.com/David-Shadow/TgLogger)
=====================

A Go library for logging messages to Telegram.

Overview
------------

This library provides a simple way to log messages to Telegram using the Telegram Bot API. It allows you to send logs to a Telegram chat or channel, making it easy to monitor and debug your applications.

Installation
---------------

To install the library, run the following command:
```
go get github.com/David-Shadow/telegram-logger
```

Usage
-----

To use this library, simply import it into your Go project and create a new instance of the TelegramLogger struct. You can then use the Log method to send logs to Telegram.

```
import "github.com/David-Shadow/tglogger"

func main() {
  logger := tglogger.NewTelegramLogger("YOUR_BOT_TOKEN", -1001234567890)
  logger.Log("Hello, world!")
}
```

Config Attributes
--------------------

The library uses a Config struct to store configuration settings. The following attributes are available:

* `ChatID`: The ID of the Telegram chat or channel to send logs to.(must include -100 for supergroups or channels)
* `Token`: The bot token obtained from the Telegram BotFather.
* `UpdateInterval`: The interval at which logs are sent to Telegram (default: 3 seconds).
* `MinimumLines`: The minimum number of log lines required to send a message to Telegram (default: 1).
* `PendingLogsSize`: The maximum number of log lines to store in memory before sending to Telegram (default: 20000).
* `MaxMessageSize`: The maximum size of a message sent to Telegram (default: 4096).
* `Title`: The title of the log message (default: "TGLogger-Go").
* `ForumTopicID`: The ID of the forum topic to send logs to (default: 0).
* `ExcludedLogPatterns`: A list of log patterns to exclude from sending to Telegram (default: empty).

Credits
----------

[eyMarv](https://github.com/eyMarv): For working on his awesome [tglogging-black](https://github.com/eyMarv/TGLogging).

[Me](https://github.com/David-Shadow): For implementing it in Golang

License
-------

This library is released under the MIT License.

Contributing
------------

Contributions are welcome! If you'd like to contribute to this library, please fork the repository and submit a pull request.