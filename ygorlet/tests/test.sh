#!/bin/sh

. ./_functions.sh

announce "unknown command"
test_command "whatevz anything goes"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
unknown command: &{ whatevz anything goes  fakeUserID}
terminating: EOF
EOF
assert_output && pass


announce "xombrero"
test_command "xombrero open http://imgur.com/fake"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
xombrero: open http://imgur.com/fake
terminating: EOF
EOF
assert_output && pass


announce "playing bad path"
test_command "play not_a_file.ogg"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
sending to soul: play error path should contain a folder
terminating: EOF
EOF
assert_output && pass


announce "playing missing file"
test_command "play tune/not_a_file.ogg"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
sending to soul: play error file not found: tune/not_a_file.ogg
terminating: EOF
EOF
assert_output && pass


announce "playing existing file"
test_command "play tunes/test.mp3"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
play: tunes/test.mp3
play: play full
terminating: EOF
EOF
assert_output && pass


announce "playing existing file with bad duration"
test_command "play tunes/test.mp3 5"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
sending to soul: play error invalid duration: time: missing unit in duration 5
terminating: EOF
EOF
assert_output && pass


announce "playing existing file with duration"
test_command "play tunes/test.mp3 5s"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
play: tunes/test.mp3
play: play with duration (5s)
terminating: EOF
EOF
assert_output && pass


announce "say something"
test_command "say something"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
say: something
terminating: EOF
EOF
assert_output && pass


announce "shutup"
test_command "shutup"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
shutup: deleting 0 items from the noise queue
terminating: EOF
EOF
assert_output && pass


announce "ping"
test_command "ping 1234567890"
cat > test.expected <<EOF
bobert-von-cheesecake ready
sending to soul: register bobert-von-cheesecake fake-queue
sending to soul: pong 1234567890
terminating: EOF
EOF
assert_output && pass


cleanup