#!/bin/sh
#
# Tests in this file are separated by two blank lines. Each test is
# self-sufficient and should cleanup after itself (use the cleanup function).
# No state should be maintained between each.
#

. ./_functions.sh

cleanup

announce "set an incremental alias"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias cat# play stuff1.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias cat# play stuff2.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias cat# play stuff3.ogg"
cat > test.expected <<EOF
PRIVMSG #test :ok (created as "cat3")
EOF
assert_output && pass
cleanup


announce "set an incremental alias (error)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias cat## play stuff.ogg"
cat > test.expected <<EOF
PRIVMSG #test :error: too many '#'
EOF
assert_output && pass
cleanup


announce "set an incremental alias (already exist)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias dog# play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias dog# play stuff.ogg"
cat > test.expected <<EOF
PRIVMSG #test :error: already exists as 'dog1'
EOF
assert_output && pass
cleanup


announce "use a recursive alias in a multi-command"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla babble;blabla"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias babble blabla;babble"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: babble"
cat > test.expected <<EOF
PRIVMSG #test :lexer/expand error: max recursion reached
EOF
assert_output && pass
cleanup


announce "alias with percent sign"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias 60% play stuff"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias 60%"
cat > test.expected <<EOF
PRIVMSG #test :60%="play stuff" (created by jimmy on 2000-01-01T00:00:00Z)
EOF
assert_output && pass
cleanup


announce "unalias usage"
rm -f test.aliases
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: unalias"
cat > test.expected <<EOF
PRIVMSG #test :usage: unalias name
EOF
assert_output && pass


announce "try to delete a non-existing alias"
rm -f test.aliases
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias notblabla play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: unalias blabla"
cat > test.expected <<EOF
PRIVMSG #test :error: unknown alias
EOF
assert_output && pass


announce "delete an existing alias"
rm -f test.aliases
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: unalias blabla"
# make sure it has really gone
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla"
cat > test.expected <<EOF
PRIVMSG #test :error: unknown alias
EOF
assert_output && pass
cleanup


announce "say stuff (unknown minions)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: say stuff"
cat > test.expected <<EOF
PRIVMSG #ygor :error: unable to load queue URLs, minion not found: pi1
EOF
assert_output && pass
cleanup


announce "bad roster file (wrong param count)"
echo "123	123	123" > test.roster
test_line_error "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
cat > test.expected <<EOF
EOF
assert_output || fail
cat > test.expected <<EOF
minions file error: minion line is missing parameters
EOF
assert_stderr && pass
cleanup


announce "bad roster file (bad timestamp)"
echo "123	123	123	qwe" > test.roster
test_line_error "minion user_id_234234234 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2"
cat > test.expected <<EOF
EOF
assert_output || fail
cat > test.expected <<EOF
minions file error: minion line has an invalid timestamp
EOF
assert_stderr && pass
cleanup


announce "say stuff"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi1"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: say stuff"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi1 say -v bruce stuff
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2 say -v bruce stuff
EOF
assert_output && pass
cleanup


announce "say stuff (wrong channel)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #stuff :whygore: say stuff"
cat > test.expected <<EOF
PRIVMSG #ygor :error: #stuff has no queue(s) configured
EOF
assert_output && pass
cleanup


# test:
#  - toggle function
#  - unknown chatter
#  - unknown command
#  - ignore private content
#  - multiple commands (;)
