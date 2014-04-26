#!/bin/sh
#
# Tests in this file are separated by two blank lines. Each test is
# self-sufficient and should cleanup after itself (use the cleanup function).
# No state should be maintained between each.
#

. ./_functions.sh


cleanup


announce "unknown chatter"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :blabla"
cat > test.expected <<EOF
EOF
assert_output && pass
cleanup


announce "unhandled command (channel)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: wtf"
cat > test.expected <<EOF
PRIVMSG #test :command not found: wtf
EOF
assert_output && pass
cleanup


announce "unhandled command (private prefixed)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG whygore :whygore: wtf"
cat > test.expected <<EOF
PRIVMSG jimmy :command not found: wtf
EOF
assert_output && pass
cleanup


announce "unhandled command (private no prefix)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG whygore :wtf"
cat > test.expected <<EOF
PRIVMSG jimmy :command not found: wtf
EOF
assert_output && pass
cleanup


announce "minion registration"
test_line "minion user_id_0123456789 register bobert-von-cheesecake https://nom.nom/super-train/"
cat > test.expected <<EOF
[SQS-SendToMinion] https://nom.nom/super-train/ register success
EOF
assert_output && pass
cleanup


announce "minions list"
test_line "minion user_id_1234567890 register bobert-von-cheesecake https://nom.nom/super-train/bobert"
test_line "minion user_id_0987654321 register jo-mac-whopper https://nom.nom/super-train/jo"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: minions"
cat > test.expected <<EOF
PRIVMSG #test :currently registered: bobert-von-cheesecake, jo-mac-whopper
EOF
assert_output && pass
cleanup


announce "ping minions (outgoing)"
test_line "minion user_id_1234567890 register pi1 https://nom.nom/super-train/bobert"
test_line "minion user_id_1234567891 register pi2 https://nom.nom/super-train/jo"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: ping"
cat > test.expected <<EOF
[SQS-SendToMinion] https://nom.nom/super-train/bobert ping 1136239445000000000
[SQS-SendToMinion] https://nom.nom/super-train/jo ping 1136239445000000000
PRIVMSG #ygor :sent to pi1: ping 1136239445000000000
PRIVMSG #ygor :sent to pi2: ping 1136239445000000000
EOF
assert_output && pass
cleanup


announce "ping minions (late response)"
test_line "minion user_id_1234567890 register pi1 https://nom.nom/super-train/bobert"
test_line "minion user_id_1234567891 register pi2 https://nom.nom/super-train/jo"
{
	sleep 0.1
	echo "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: ping"
	sleep 0.1
	echo "minion user_id_1234567890 pong 1136239945000000000"
	sleep 0.2
} | test_input
cat > test.expected <<EOF
[SQS-SendToMinion] https://nom.nom/super-train/bobert ping 1136239445000000000
[SQS-SendToMinion] https://nom.nom/super-train/jo ping 1136239445000000000
PRIVMSG #ygor :sent to pi1: ping 1136239445000000000
PRIVMSG #ygor :sent to pi2: ping 1136239445000000000
PRIVMSG #ygor :pong: got old ping reponse (1136239945000000000)
EOF
assert_output && pass
cleanup


announce "ping minions (good response)"
test_line "minion user_id_1234567890 register pi1 https://nom.nom/super-train/bobert"
test_line "minion user_id_1234567891 register pi2 https://nom.nom/super-train/jo"
{
	echo "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: ping"
	sleep 0.5
	echo "minion user_id_1234567890 pong 1136239445000000000"
} | test_input
tail -n 1 test.output \
	| sed 's/[0-9]*h[0-9]*m[0-9.]*s/stuff/' \
	> test.tmp
mv test.tmp test.output
cat > test.expected <<EOF
PRIVMSG #test :delay with pi1: stuff
EOF
assert_output && pass
cleanup


announce "play"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: play stuff.ogg"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1 play stuff.ogg
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2 play stuff.ogg
EOF
assert_output && pass
cleanup


announce "play w/ duration"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: play stuff.ogg 5s"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1 play stuff.ogg 5s
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2 play stuff.ogg 5s
EOF
assert_output && pass
cleanup


announce "set a new alias"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
cat > test.expected <<EOF
PRIVMSG #test :ok (created)
EOF
assert_output && pass
cleanup


announce "set a new alias (bad command)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla playz stuff.ogg"
cat > test.expected <<EOF
PRIVMSG #test :error: 'playz' is not a valid command
EOF
assert_output && pass
cleanup


announce "set a new alias (permission error)"
touch aliases.cfg
chmod 000 aliases.cfg
test_line_error "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
cat > test.expected <<EOF
EOF
assert_output
cat > test.expected <<EOF
alias file error: open aliases.cfg: permission denied
EOF
sed 's/^....................//' test.stderr > test.tmp
mv test.tmp test.stderr
assert_stderr && pass
cleanup


announce "get this new alias"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla"
cat > test.expected <<EOF
PRIVMSG #test :'blabla' is an alias for 'play stuff.ogg'
EOF
assert_output && pass
cleanup


announce "change this alias"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play things.ogg"
cat > test.expected <<EOF
PRIVMSG #test :ok (replaced)
EOF
assert_output && pass
cleanup


announce "get this updated alias"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play things.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla"
cat > test.expected <<EOF
PRIVMSG #test :'blabla' is an alias for 'play things.ogg'
EOF
assert_output && pass
cleanup


announce "get unknown alias (empty registry)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias whatevs"
cat > test.expected <<EOF
PRIVMSG #test :error: unknown alias
EOF
assert_output && pass
cleanup


announce "get unknown alias (non-empty registry)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias things play things.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias whatevs"
cat > test.expected <<EOF
PRIVMSG #test :error: unknown alias
EOF
assert_output && pass
cleanup


announce "list all known aliases alphabetically"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias blabla play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias zelda play zelda.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias beer play beer.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: aliases"
cat > test.expected <<EOF
PRIVMSG #test :known aliases: beer, blabla, zelda
EOF
assert_output && pass
cleanup


announce "list aliases by pages of 400 bytes at most"
for each in 0 1 2 3 4 5 6 7 8 9 A B C D E F G H I J K L M N O P Q R S T U V W X Y Z; do
	test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias randomlongaliasfromhell$each play stuff.ogg"
done
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: aliases"
cat > test.expected <<EOF
PRIVMSG #test :known aliases: randomlongaliasfromhell0, randomlongaliasfromhell1, randomlongaliasfromhell2, randomlongaliasfromhell3, randomlongaliasfromhell4, randomlongaliasfromhell5, randomlongaliasfromhell6, randomlongaliasfromhell7, randomlongaliasfromhell8, randomlongaliasfromhell9, randomlongaliasfromhellA, randomlongaliasfromhellB, randomlongaliasfromhellC, randomlongaliasfromhellD, randomlongaliasfromhellE, randomlongaliasfromhellF, randomlongaliasfromhellG
PRIVMSG #test :... randomlongaliasfromhellH, randomlongaliasfromhellI, randomlongaliasfromhellJ, randomlongaliasfromhellK, randomlongaliasfromhellL, randomlongaliasfromhellM, randomlongaliasfromhellN, randomlongaliasfromhellO, randomlongaliasfromhellP, randomlongaliasfromhellQ, randomlongaliasfromhellR, randomlongaliasfromhellS, randomlongaliasfromhellT, randomlongaliasfromhellU, randomlongaliasfromhellV, randomlongaliasfromhellW, randomlongaliasfromhellX
PRIVMSG #test :... randomlongaliasfromhellY, randomlongaliasfromhellZ
EOF
assert_output && pass
cleanup


announce "alias with percent sign"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias 60% play stuff"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias 60%"
cat > test.expected <<EOF
PRIVMSG #test :'60%' is an alias for 'play stuff'
EOF
assert_output && pass
cleanup


announce "unalias usage"
rm -f aliases.cfg
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: unalias"
cat > test.expected <<EOF
PRIVMSG #test :usage: unalias name
EOF
assert_output && pass


announce "try to delete a non-existing alias"
rm -f aliases.cfg
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias notblabla play stuff.ogg"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: unalias blabla"
cat > test.expected <<EOF
PRIVMSG #test :error: unknown alias
EOF
assert_output && pass


announce "delete an existing alias"
rm -f aliases.cfg
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
PRIVMSG #ygor :error: unable to load queue URLs, minion not found
EOF
assert_output && pass
cleanup


announce "bad minions file (wrong param count)"
echo "123	123	123" > minions.cfg
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
cat > test.expected <<EOF
PRIVMSG #ygor :register: error: minion line is missing parameters
EOF
assert_output && pass
cleanup


announce "bad minions file (bad timestamp)"
echo "123	123	123	qwe" > minions.cfg
test_line "minion user_id_234234234 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
cat > test.expected <<EOF
PRIVMSG #ygor :register: error: minion line has an invalid timestamp
EOF
assert_output && pass
cleanup


announce "say stuff"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: say stuff"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1 say stuff
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2 say stuff
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


announce "use alias"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: alias 60% play stuff"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: 60%"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1 play stuff
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2 play stuff
EOF
assert_output && pass
cleanup


announce "sshhhh"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: sshhhh"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1 shutup
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2 shutup
PRIVMSG #test :ok...
EOF
assert_output && pass
cleanup


announce "stopwhining"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: stopwhining"
cat > test.expected <<EOF
EOF
assert_output && pass
cleanup


announce "sshhhh by ignored nick"
test_line "irc :douchebot!dev@truveris.com PRIVMSG #test :whygore: sshhhh"
cat > test.expected <<EOF
EOF
assert_output && pass
cleanup


announce "sshhhh privately (not owner)"
test_line "irc :jimmy!dev@truveris.com PRIVMSG whygore :whygore: sshhhh"
cat > test.expected <<EOF
EOF
assert_output && pass
cleanup


announce "image"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: image http://imgur.com/stuff"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1 xombrero open http://truveris.github.io/fullscreen-image/?http://imgur.com/stuff
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2 xombrero open http://truveris.github.io/fullscreen-image/?http://imgur.com/stuff
EOF
assert_output && pass
cleanup


announce "xombrero"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: xombrero open http://www.truveris.com/"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1 xombrero open http://www.truveris.com/
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2 xombrero open http://www.truveris.com/
EOF
assert_output && pass
cleanup


announce "xombrero ack"
test_line "minion user_id_1234567890 register pi1 https://nom.nom/super-train/bobert"
test_line "minion user_id_1234567891 register pi2 https://nom.nom/super-train/jo"
test_line "minion user_id_1234567890 xombrero ok"
cat > test.expected <<EOF
PRIVMSG #ygor :unhandled minion message: xombrero ok
EOF
assert_output && pass
cleanup


announce "xombrero error"
test_line "minion user_id_1234567890 register pi1 https://nom.nom/super-train/bobert"
test_line "minion user_id_1234567891 register pi2 https://nom.nom/super-train/jo"
test_line "minion user_id_1234567890 xombrero error stuff"
cat > test.expected <<EOF
PRIVMSG #ygor :unhandled minion message: xombrero error stuff
EOF
assert_output && pass
cleanup


announce "reboot"
test_line "minion user_id_234234234 register pi1 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1"
test_line "minion user_id_123123123 register pi2 http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2"
test_line "irc :jimmy!dev@truveris.com PRIVMSG #test :whygore: reboot"
cat > test.expected <<EOF
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi1 reboot
[SQS-SendToMinion] http://sqs.us-east-1.amazonaws.com/000000000000/ygor-minion-pi2 reboot
PRIVMSG #test :attempting to reboot #test minions...
EOF
assert_output && pass
cleanup
