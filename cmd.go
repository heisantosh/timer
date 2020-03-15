package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

// Use the same text in README as well.
const (
	_helpText = `timer version 0.0.1

Set a timer. Play a sound when the timer expires. Receive notification when the
timer expires.

List of available options
	-t,time TIME        time value
	-s,sound NAME       play this sound after timer expires
	-l,sounds           show the list of available sounds
	-n,notify           show notification
	-a,addsound FILE    add FILE to the sound library
	-d,deletesound NAME remove the sound named NAME from the sound library
	-v,verbose          if true print more details on error
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

By default the audacious applicatoin will be used to play the sound. The default command is:
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
	$ timer -sound Rooster`
)

// Places of each argument in a bitmap
// Combinations of these arguements will be a key map to a function which
// can process the set of arguments.
const (
	_argTime = iota + 1
	_argSound
	_argSounds
	_argNotify
	_argAddSound
	_argDeleteSound
)

var (
	errSoundNotFound = errors.New("Sound not found in library")
)

const (
	// Use the default sound command if environment variable
	// TIMER_SOUND_CMD is not set

	// default sound command, requires to have audacious installed
	_defaultSoundCommand = "audacious --headless --quit-after-play FILE"
	// name of environment variable storing custom sound command
	_timerSoundCommand = "TIMER_SOUND_CMD"
)

// cmdArgs is the set of arguments for the command
type cmdArgs struct {
	time        string
	sound       string
	sounds      bool
	notify      bool
	addSound    string
	deleteSound string
	verbose     bool
}

// Cmd represents the command
type Cmd struct {
	args cmdArgs
	// map of argument set to function to process the argument set
	funcs map[int]func() error
	// map of name of sound to location of the soudn file on filesystem
	sounds map[string]string
}

// NewCmd creates a new instance of the command
func NewCmd() *Cmd {
	cmd := &Cmd{}

	// Map argument set to corresponding function
	cmd.funcs = make(map[int]func() error)
	cmd.funcs[1<<_argTime] = cmd.timed
	cmd.funcs[1<<_argTime|1<<_argSound] = cmd.timedSound
	cmd.funcs[1<<_argTime|1<<_argNotify] = cmd.timedNotify
	cmd.funcs[1<<_argTime|1<<_argSound|1<<_argNotify] = cmd.timedSoundNotify
	cmd.funcs[1<<_argSounds] = cmd.listSounds
	cmd.funcs[1<<_argSound] = cmd.playSound
	cmd.funcs[1<<_argAddSound] = cmd.addSound
	cmd.funcs[1<<_argDeleteSound] = cmd.deleteSound

	cmd.sounds = make(map[string]string)
	soundsDir := getSoundsDir()

	createConfigIfNotExists(soundsDir)

	fi, err := ioutil.ReadDir(soundsDir)
	if err != nil {
		fmt.Println("Error reading list of sounds available:", err)
		os.Exit(1)
	}

	for _, v := range fi {
		name := strings.Replace(filepath.Base(v.Name()), filepath.Ext(v.Name()), "", 1)
		cmd.sounds[name] = filepath.Join(soundsDir, v.Name())
	}

	return cmd
}

func createConfigIfNotExists(soundsDir string) {
	_, err := os.Stat(soundsDir)
	if os.IsNotExist(err) {
		if e := os.MkdirAll(soundsDir, 0776); e != nil {
			fmt.Println("Error creating config directory:", err)
			os.Exit(1)
		}
	} else if err != nil {
		fmt.Println("Error checking if sounds config directory exists:", err)
		os.Exit(1)
	}
}

// timed processes the argument set (time).
// Run the timer for the give amount of time.
func (cmd *Cmd) timed() error {
	t, err := time.ParseDuration(cmd.args.time)
	if err != nil {
		fmt.Println("Error parsing time value")
		return err
	}

	unit := t / 100
	ticker := time.NewTicker(t / 100)
	done := make(chan struct{})

	fmt.Printf("\r                                                                                 ")
	fmt.Printf("\r⏲  %3d%% [passed: %v, remaining: %v, total: %v]", 0, 0, t, t)

	go func() {
		pc := 1
		passed := unit
		for {
			select {
			case <-ticker.C:
				fmt.Printf("\r                                                                         ")
				fmt.Printf("\r⏲  %3d%% [passed: %v, remaining: %v, total: %v]", pc, passed, t-passed, t)
				passed += unit
				pc ++
			case <-done:
				return
			}
		}
	}()

	time.Sleep(t)
	done <- struct{}{}

	fmt.Println("\n⏰  Timer expired!")
	return nil
}

// timedSound processes the argument set (time, sound).
// Run the timer for the given amount of time and play the sound.
func (cmd *Cmd) timedSound() error {
	if _, ok := cmd.sounds[cmd.args.sound]; !ok {
		fmt.Printf("Selected sound %s not available\n", cmd.args.sound)
		return errSoundNotFound
	}

	if err := cmd.timed(); err != nil {
		return err
	}
	if err := cmd.playSound(); err != nil {
		return err
	}
	return nil
}

// notify shows a notifcation.
func (cmd *Cmd) notify() error {
	if err := beeep.Notify("Timer", "Time is expired!", ""); err != nil {
		fmt.Println("Error showing notification")
		return err
	}

	return nil
}

// timedNotify processes the argument set (time, notify).
// Run the timer for the given amount of time and after that show a notification.
func (cmd *Cmd) timedNotify() error {
	if err := cmd.timed(); err != nil {
		return err
	}
	if err := cmd.notify(); err != nil {
		return err
	}
	return nil
}

// timedSoundNotify processes the argument set (time, sound, notify).
// Run the timer for the given amount of time, show the notification and play the sound.
func (cmd *Cmd) timedSoundNotify() error {
	if err := cmd.timed(); err != nil {
		return err
	}
	if err := cmd.notify(); err != nil {
		return err
	}
	if err := cmd.playSound(); err != nil {
		return err
	}
	return nil
}

// listSounds processes the argument set (sounds).
// List the name of available sounds.
func (cmd *Cmd) listSounds() error {
	for k := range cmd.sounds {
		fmt.Println(k)
	}

	return nil
}

// addSound processes the argument set (addsound).
// Add the given file to the sound library by copying it to the configuration
// sounds directory. $HOME/.config/timer/sounds on Linux and %HOME%\AppData\timer\sounds
// on Windows.
func (cmd *Cmd) addSound() error {
	fileLoc := cmd.args.addSound
	data, err := ioutil.ReadFile(fileLoc)
	if err != nil {
		fmt.Println("Error adding sound file")
		return err
	}

	newFileLoc := filepath.Join(getSoundsDir(), filepath.Base(fileLoc))
	if err = ioutil.WriteFile(newFileLoc, data, 0644); err != nil {
		fmt.Println("Error adding sound file")
		return err
	}

	return nil
}

// deleteSound processes the argument set (deletesound).
// Remove the given sound name from the sound library. Delete the corresponding file
// from the sounds configuration directory.
func (cmd *Cmd) deleteSound() error {
	fileLoc, ok := cmd.sounds[cmd.args.deleteSound]
	if !ok {
		fmt.Println("Sound with the given name not found")
		return errSoundNotFound
	}

	if err := os.Remove(fileLoc); err != nil {
		fmt.Println("Unable to remove the sound with given name")
		return err
	}

	return nil
}

// playSound processes the argument set (sound).
// Play the sound with the given name.
func (cmd *Cmd) playSound() error {
	sound := cmd.args.sound
	if _, ok := cmd.sounds[sound]; !ok {
		fmt.Println("Selected sound not found")
		return errSoundNotFound
	}

	command := os.Getenv(_timerSoundCommand)
	if command == "" {
		command = _defaultSoundCommand
	}

	c := strings.Replace(command, "FILE", cmd.sounds[sound], 1)
	s := strings.Split(c, " ")
	ex := exec.Command(s[0], s[1:]...)

	if _, err := ex.CombinedOutput(); err != nil {
		fmt.Println("Error playing sound")
		return err
	}

	return nil
}

// Run runs the command
func (cmd *Cmd) Run() {
	flag.StringVar(&cmd.args.time, "time", "", "time value")
	flag.StringVar(&cmd.args.time, "t", "", "time value")
	flag.StringVar(&cmd.args.sound, "sound", "", "play this sound after timer expires")
	flag.StringVar(&cmd.args.sound, "s", "", "play this sound for 10 seconds when the timer expires")
	flag.BoolVar(&cmd.args.sounds, "l", false, "show the list of available sounds")
	flag.BoolVar(&cmd.args.sounds, "sounds", false, "show the list of available sounds")
	flag.BoolVar(&cmd.args.notify, "notify", false, "show notification")
	flag.BoolVar(&cmd.args.notify, "n", false, "show notification")
	flag.StringVar(&cmd.args.addSound, "addsound", "", "add this sound to the sound library")
	flag.StringVar(&cmd.args.addSound, "a", "", "add this sound to the sound library")
	flag.StringVar(&cmd.args.deleteSound, "deletesound", "", "delete this sound from the sound library")
	flag.StringVar(&cmd.args.deleteSound, "d", "", "delete this sound from the sound library")
	flag.BoolVar(&cmd.args.verbose, "verbose", false, "if provided will print more details on error")
	flag.BoolVar(&cmd.args.verbose, "v", false, "if provided will print more details on error")

	flag.Usage = func() {
		fmt.Println(_helpText)
	}

	flag.Parse()

	argsSet := 0
	if cmd.args.time != "" {
		argsSet |= 1 << _argTime
	}
	if cmd.args.sound != "" {
		argsSet |= 1 << _argSound
	}
	if cmd.args.sounds != false {
		argsSet |= 1 << _argSounds
	}
	if cmd.args.notify != false {
		argsSet |= 1 << _argNotify
	}
	if cmd.args.addSound != "" {
		argsSet |= 1 << _argAddSound
	}
	if cmd.args.deleteSound != "" {
		argsSet |= 1 << _argDeleteSound
	}

	if f, ok := cmd.funcs[argsSet]; ok {
		if err := f(); err != nil {
			if cmd.args.verbose {
				fmt.Println(err)
			}
			os.Exit(1)
		}
		return
	}

	fmt.Println("Received invalid set of options")
	fmt.Println("Type 'timer -help' to see how to use")
	os.Exit(1)
}
