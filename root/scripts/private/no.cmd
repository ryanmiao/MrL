#!/usr/bin/python

import redis
import os
import sys
import string

args = len(sys.argv)
port = sys.argv[1]
server = sys.argv[2]
channel = sys.argv[3]
user = sys.argv[4]

"""
cmd = "echo \'%s 1 PRIVMSG %s :%s\' | nc 127.0.0.1 %s" % (server, user, 'no cmd', port)
os.system(cmd.encode('utf-8'))

"""
if args <= 5:
	print "arguments wrong"
	exit()

r = redis.Redis('localhost')

msg = sys.argv[5]
if not "is" in string.splitfields(msg):
	print "cmd format is wrong"
	exit()

pos = string.splitfields(msg).index('is')
key = string.join(string.splitfields(msg)[:pos], '_')
val = string.join(string.splitfields(msg)[pos+1:])
print key
print val

if val is "":
	print "value is blank"
	exit()

if not r.set(key, val):
	print "failed to store key-value"
	exit()

cmd = "echo \'%s 1 PRIVMSG %s :%s> %s\' | nc 127.0.0.1 %s" % (server, user, user, 'ok', port)
os.system(cmd.encode('utf-8'))
