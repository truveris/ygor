#!/bin/sh

. ./_functions.sh

# $1 command
cmd() {
	sleep 0.1
	echo "$@"
	sleep 0.1
}

# $1 command
test_command() {
	cmd "$@" \
		| ../ygor-minion/ygor-minion \
		2>&1 \
		| remove_timestamp \
		> test.output
}


announce "playing bad path"
test_command "play not_a_file.ogg"
cat > test.expected <<EOF
ygor-minion ready!
got message: "play not_a_file.ogg"
play: path should contain a folder
terminating: EOF
EOF
assert_output && pass


announce "playing missing file"
test_command "play tune/not_a_file.ogg"
cat > test.expected <<EOF
ygor-minion ready!
got message: "play tune/not_a_file.ogg"
play: file not found (tune/not_a_file.ogg)
terminating: EOF
EOF
assert_output && pass


announce "playing existing file"
test_command "play tunes/test.mp3"
cat > test.expected <<EOF
ygor-minion ready!
got message: "play tunes/test.mp3"
play: path: tunes/test.mp3
terminating: EOF
EOF
assert_output && pass


announce "playing existing file with duration"
test_command "play tunes/test.mp3 5"
cat > test.expected <<EOF
ygor-minion ready!
got message: "play tunes/test.mp3 5"
play: path: tunes/test.mp3
terminating: EOF
EOF
assert_output && pass


announce "say something"
test_command "say something"
cat > test.expected <<EOF
ygor-minion ready!
got message: "say something"
say(something)
terminating: EOF
EOF
assert_output && pass


announce "shutup"
test_command "shutup"
cat > test.expected <<EOF
ygor-minion ready!
got message: "shutup"
deleting 0 items from the noise queue
terminating: EOF
EOF
assert_output && pass


cleanup
