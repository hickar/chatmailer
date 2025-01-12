## Chatmailer

Chatmailer is a email forwarder daemon tool. It provides features to specify any number of
source email inboxes and choose specific email letters by filters to be watched and notified upon
it's receipt by various backends.

Currently, only Telegram forwarder backend is supported.

## Setup local environment

Requirements:
1. Go 1.23 installed
2. Docker engine installed

Setup steps:
1. `cp config.example.yaml config.yaml`
2. Change Telegram Bot API Token and chat room in configuration file to your own ones.
3. Change IMAP credentials specified in configuration file to your own ones.
4. `make up`
