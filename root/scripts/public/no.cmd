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

if args <= 5:
	exit()

r = redis.Redis('localhost')

msg = sys.argv[5]
if not "is" in string.splitfields(msg):
	exit()

pos = string.splitfields(msg).index('is')
key = string.join(string.splitfields(msg)[:pos], '_')
val = string.join(string.splitfields(msg)[pos+1:])

if val is "":
	exit()

if not r.set(key, val):
	exit()

cmd = "echo \'%s 1 PRIVMSG %s :%s> %s\' | nc 127.0.0.1 %s" % (server, channel, user, 'ok', port)
os.system(cmd.encode('utf-8'))
