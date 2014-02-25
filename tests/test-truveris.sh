#!/bin/sh
#
# Tests in this file are separated by two blank lines. Each test is
# self-sufficient and should cleanup after itself (use the cleanup function).
# No state should be maintained between each.
#

. ./_functions.sh

# $1 command
cmd() {
	sleep 0.1
	echo "$@"
	sleep 0.1
}

# $1 command
test_line() {
	cmd "$@" \
		| ../ygor-truveris/ygor-truveris --nickname=whygore \
		2> test.stderr \
		> test.output
	if [ "$?" != 0 ]; then
		fail "wrong return code (check test.stderr)"
	fi
}


cleanup


announce "auto-joins"
test_line ""
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
EOF
assert_output && pass
cleanup


announce "unknown chatter"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :blabla"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
EOF
assert_output && pass
cleanup


announce "set a new alias"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #test :ok (created)
EOF
assert_output && pass
cleanup


announce "set a new alias (permission error)"
touch aliases.cfg
chmod 000 aliases.cfg
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #test :failed: open aliases.cfg: permission denied
EOF
assert_output && pass
cleanup


announce "get this new alias"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #ygor :loaded 1 aliases
PRIVMSG #test :'blabla' is an alias for 'play stuff.ogg'
EOF
assert_output && pass
cleanup


announce "change this alias"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play things.ogg"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #ygor :loaded 1 aliases
PRIVMSG #test :ok (replaced)
EOF
assert_output && pass
cleanup


announce "get this updated alias"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play things.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #ygor :loaded 1 aliases
PRIVMSG #test :'blabla' is an alias for 'play things.ogg'
EOF
assert_output && pass
cleanup


announce "get unknown alias (empty registry)"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias whatevs"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #test :error: unknown alias
EOF
assert_output && pass
cleanup


announce "get unknown alias (non-empty registry)"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias things play things.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias whatevs"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #ygor :loaded 2 aliases
PRIVMSG #test :error: unknown alias
EOF
assert_output && pass
cleanup


announce "list all known aliases alphabetically"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias zelda play zelda.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias beer play beer.ogg"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: aliases"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #ygor :loaded 3 aliases
PRIVMSG #test :known aliases: beer, blabla, zelda
EOF
assert_output && pass
cleanup


announce "alias with percent sign"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias 60% play stuff"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias 60%"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #ygor :loaded 1 aliases
PRIVMSG #test :'60%' is an alias for 'play stuff'
EOF
assert_output && pass
cleanup


announce "say stuff"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: say stuff"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
[SQS-SendToMinion] say stuff
EOF
assert_output && pass
cleanup


announce "use alias"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: alias 60% play stuff"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: 60%"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
PRIVMSG #ygor :loaded 1 aliases
[SQS-SendToMinion] play stuff
EOF
assert_output && pass
cleanup


announce "sshhhh"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: sshhhh"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
[SQS-SendToMinion] shutup
PRIVMSG #test :ok...
EOF
assert_output && pass
cleanup


announce "sshhhh by ignored nick"
test_line ":douchebot!dev@truveris.com PRIVMSG #test :whygore: sshhhh"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
EOF
assert_output && pass
cleanup


announce "sshhhh privately (not owner)"
test_line ":jimmy!dev@truveris.com PRIVMSG whygore :whygore: sshhhh"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
EOF
assert_output && pass
cleanup


announce "sshhhh privately (owner)"
test_line ":hippalectryon!hippalectryon@truveris.com PRIVMSG whygore :whygore: sshhhh"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
[SQS-SendToMinion] shutup
PRIVMSG hippalectryon :ok...
EOF
assert_output && pass
cleanup


announce "xombrero"
test_line ":jimmy!dev@truveris.com PRIVMSG #test :whygore: xombrero open http://www.truveris.com/"
cat > test.expected <<EOF
JOIN #test
JOIN #ygor
[SQS-SendToMinion] xombrero open http://www.truveris.com/
PRIVMSG #test :sure
EOF
assert_output && pass
cleanup
