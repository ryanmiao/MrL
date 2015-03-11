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

r = redis.Redis('localhost')

val = r.get('name_list')
if val == None:
	print "failed to access value"
	exit()

cmd = "echo \'%s 1 PRIVMSG %s :%s> name list is %s\' | nc 127.0.0.1 %s" % (server, channel, user, val, port)
os.system(cmd.encode('utf-8'))
