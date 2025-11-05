# TGLogger

A robust Go library for sending application logs to Telegram channels or chats in real-time.

## Overview

TGLogger is a powerful logging library for Go applications that enables real-time log streaming to Telegram. It provides an efficient way to monitor your application logs through Telegram chats or channels, making remote debugging and monitoring more accessible.

## Features

- Real-time log streaming to Telegram
- Support for both regular chats and forum topics
- Configurable update intervals
- Message size management and buffering
- Pattern-based log exclusion
- Automatic retry mechanism with flood control
- Local log file backup

## Installation

```bash
go get github.com/David-Shadow/tglogger
```

## Configuration

The library can be configured using the `Config` struct:

```go
type Config struct {
    Token               string        // Your Telegram Bot Token
    ChatID              int64         // Target Chat/Channel ID
    ForumTopicID        int          // Optional: Forum Topic ID for forum messages
    Title               string        // Title for log messages (default: "TGLogger-Go")
    ExcludedLogPatterns []string     // Patterns to exclude from logging
    UpdateInterval      time.Duration // Interval between log updates (default: 3s)
    MinimumLines        int          // Minimum lines before sending update (default: 1)
    PendingLogsSize     int          // Buffer size for pending logs (default: 20000)
    MaxMessageSize      int          // Maximum message size (default: 4096)
}
```

## Usage

Basic usage example:

```go
package main

import (
    "github.com/David-Shadow/tglogger"
    "time"
)

func main() {
    config := &tglogger.Config{
        Token:          "YOUR_BOT_TOKEN",
        ChatID:        -100123456789,
        UpdateInterval: 5 * time.Second,
    }

    err := tglogger.InitializeTgLogger(config)
    if err != nil {
        panic(err)
    }

    // Your application code here
    // All logs will be automatically sent to Telegram
}
```

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

This library is released under the Apache License.

Contributing
------------

Contributions are welcome! If you'd like to contribute to this library, please fork the repository and submit a pull request.