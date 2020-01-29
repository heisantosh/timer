# timer

A CLI application to run timers.

⏲

## Installation

Using go get:

`go get -u github.com/heisantosh/timer`

## Usage

```bash
$ timer -help
timer version 0.0.1

Set a timer. Play a sound when the timer expires. Receive notification when the
timer expires.

List of available options
	-t,time TIME        time value
	-s,sound NAME       play this sound after timer expires
	-l,sounds             show the list of available sounds
	-n,notify           show notification
	-a,addsound FILE    add FILE to the sound library
	-d,deletesound NAME	remove the sound named NAME from the sound library
	-v,verbose			if true print more details on error
	-h,help             show this help information

Command to play the sound is read from the environment variable SOUND_CMD.
It should contain the placehoder text FILE where the filename should
appear in the command.

Added sounds are stored in $HOME/.config/timer/sounds directory on Linux 
and %HOME%\AppData\timer\sounds on Windows. Name of the file is the name of the sound.

Time value is of the format 1h20m30s. Some valid examples are:
	2h          time of 2 hours
	1h5m        time of 1 hour 5 minutes
	5h10m10s    time of 5 hours 10 minutes 10seconds
	70m         time of 70 minutes
	100s        time of 100 seconds
	2m200s      time of 2 minutes 200 seconds

By default the audacious will be used to pla the sound. The default command is:
	audacious -H -q FILE

where FILE is the location of the audio file.

A custom command can be set via the environment variable TIMER_SOUND_CMD.

Examples:
	$ # set custom sound command to ffplay
	$ export TIMER_SOUND_CMD="ffplay -nodisp -autoexit -i FILE -hide_banner -loglevel panic"
	$ # start a timer of 30 minutes
	$ timer -t 30m
	$ # start a timer of 3 minutes 101 seconds
	$ timer -t 3m101s
	$ # start a timer and play sound when expired
	$ timer -t 30m -s Alien
	$ # start a timer and play sound when expired and show notification
	$ timer -t 30m -s Alien -notify
	$ # listen to a sound
	$ timer -sound Rooster
```
