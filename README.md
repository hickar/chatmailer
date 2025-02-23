## Chatmailer

Chatmailer is an email forwarder daemon tool. It provides features to specify any number of
source email inboxes and choose specific email letters by filters to be watched and notified upon
it's receipt by various backends.

Currently, only Telegram forwarder backend is supported.


## Setup local environment

> Requirements:
> [Go 1.23](https://go.dev/doc/devel/release) installed & [Docker engine](https://docs.docker.com/engine/) installed


```bash
cp config.example.yaml config.yaml
```

### Telegram

To enable Telegram notifications, you'll need to create a Telegram bot and obtain the chat ID where messages will be sent.

**Create a Telegram Bot**:
- Open the Telegram app and search for @BotFather.
- Start a chat and send the command /newbot.
- Follow the prompts to set up the bot's name and username. After creation, BotFather will provide a bot token. Keep this token secure.

**Obtain Your Chat ID**:
- Add the newly created bot to the desired chat (either a group or personal chat).
- Send a message to the bot in that chat.

To retrieve the chat ID, open a web browser and navigate to:

```bash
https://api.telegram.org/bot<token>/getUpdates
```

Replace <token> with the token you received from BotFather.

In the JSON response, locate the "chat" object. The "id" field within this object is your chat ID.


### IMAP/POP

#### Gmail

> Note: Starting January 2025, IMAP access is always turned on in Gmail, and the option to enable or disable it is no longer available. 
> Your current connections to other email clients aren’t affected, and no action is needed if you're accessing Gmail through IMAP.
> - imap for gmail: `imap.gmail.com:993`


**If you want to use POP**:
- Log in to your Gmail account.
- Click on the gear icon in the top-right corner and select "See all settings."
- Navigate to the "Forwarding and POP/IMAP" tab.
- In the "POP access" section, select "Enable POP."
- Click "Save Changes" at the bottom.


**Generate an App Password**:
> If you have 2-Step Verification enabled on your Google account, you'll need to generate an App Password

- Go to your [Google App Passwords page](https://myaccount.google.com/apppasswords).
- You may be prompted to enter your Google account password.
- Write application name "chatmailer" for example
- Copy generated password
- Approve generated password on Google Account Security page


```bash
make up
```

Compose and sent email to yourself address.
