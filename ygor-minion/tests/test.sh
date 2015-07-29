#!/bin/sh

. ./_functions.sh
cleanup


announce "unknown command"
test_command "whatevz anything goes"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
unknown command: &{   whatevz anything goes  fakeUserID %!s(uint64=0) %!s(uint64=0) %!s(uint64=0)}
terminating: EOF
EOF
assert_output && pass


announce "xombrero"
test_command "xombrero open http://imgur.com/fake"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
xombrero: open http://imgur.com/fake
terminating: EOF
EOF
assert_output && pass


announce "playing bad path"
test_command "play not_a_file.ogg"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
send to ygord: play error invalid URL
terminating: EOF
EOF
assert_output && pass


announce "shutup"
test_command "shutup"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
shutup: deleting 0 items from the noise queue
terminating: EOF
EOF
assert_output && pass


announce "ping"
test_command "ping 1234567890"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
send to ygord: pong 1234567890
terminating: EOF
EOF
assert_output && pass


announce "volume bad"
test_command "volume bad"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
send to ygord: volume error invalid input
volume: invalid input 'bad'
terminating: EOF
EOF
assert_output && pass


announce "volume bad command"
test_command "volume 50%"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
send to ygord: volume error starting amixer
volume: error starting amixer: fork/exec ./test.amixer: no such file or directory
terminating: EOF
EOF
assert_output && pass


cp fake-amixer test.amixer
export FAKE_AMIXER_RETURN_CODE=0
announce "volume 50%"
test_command "volume 50%"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
send to ygord: volume ok
terminating: EOF
EOF
assert_output && pass


cp fake-amixer test.amixer
export FAKE_AMIXER_RETURN_CODE=1
announce "volume bad retcode"
test_command "volume 50%"
cat > test.expected <<EOF
bobert-von-cheesecake starting up
send to ygord: register bobert-von-cheesecake fake-queue
send to ygord: volume error during amixer
volume: error during amixer: exit status 1
terminating: EOF
EOF
assert_output && pass


cleanup
