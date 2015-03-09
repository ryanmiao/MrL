#!/usr/bin/python

import os
import sys
import urllib2
from bs4 import BeautifulSoup

args = len(sys.argv)
port = sys.argv[1]
server = sys.argv[2]
channel = sys.argv[3]
user = sys.argv[4]

url = "https://bugzilla.redhat.com/show_bug.cgi?id="

for bugId in sys.argv[5:]:
	bugUrl = "%s%s" % (url, bugId)
	soup = BeautifulSoup(urllib2.urlopen(bugUrl))
	output = soup.title.string
	if output == "Access Denied":
		cmd = "echo \'%s 1 PRIVMSG %s :Bug %s is unaccessible\' | nc 127.0.0.1 %s" % (server, user, bugId, port)
	elif output == "Invalid Bug ID":
		cmd = "echo \'%s 1 PRIVMSG %s :%s> Bug %s is invalid\' | nc 127.0.0.1 %s" % (server, user, user, bugId, port)
	else:
		cmd = "echo \'%s 1 PRIVMSG %s :%s> %s\' | nc 127.0.0.1 %s" % (server, user, user, output, port)
	os.system(cmd.encode('utf-8'))
