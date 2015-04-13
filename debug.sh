#!/usr/bin/env zsh

DST="/Applications/Ableton Live 9 Suite.app/Contents/App-Resources/MIDI Remote Scripts/AInterfaz"

rm "$DST/*.pyc"
cp *.py "$DST/"
osascript -e 'quit app "Ableton Live 9 Suite"'
sleep 1
open "/Applications/Ableton Live 9 Suite.app"