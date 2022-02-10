# Duty Bot
**Duty Bot** schedules duties and notifies about a change through the notification channel you have provided. You can have multiple projects with different settings which include what a notification message would look like, how often a person should be changed, should it be changed on day offs.

## Usage

You don't have to set up any dependencies to use the Duty Bot, all you need to do is simpy:
```
make
./bin/duty_bot -config $path_to_config
```

Duty Bot uses yaml for configuration, you can derive your own from the self-documented [example](https://github.com/gibsn/duty_bot/blob/main/duty_bot_example.yaml) in this repository.

## Notification channel
Currently only MyTeam is supported, but you can make a pull request for any other notification channel you need. There are two things you need to do:
1. Implement the notifyChannel interface:
```golang
type notifyChannel interface {
	Send(string) error
	Shutdown() error
}
```

2. Provide a way to configure your notification channel from yaml

## Persistence
Duty Bot is robust to restarts because it persists current state for each project on disk. It does not use any external dependency like MySQL or any other DB but stores states as simple files on FS.

## Determining day offs
Duty Bot can be set up to skip scheduling on day offs. It periodically polls [isDayOff](https://isdayoff.ru) to find info about holidays (currently only Russian) and caches it for some period of time. You can tune poll period ant cache TTL, however defaults should work fine.
