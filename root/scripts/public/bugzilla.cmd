#!/usr/bin/python

import os
import sys
import string
import urllib2
from bs4 import BeautifulSoup

args = len(sys.argv)
port = sys.argv[1]
server = sys.argv[2]
channel = sys.argv[3]
user = sys.argv[4]

url = "https://bugzilla.redhat.com/show_bug.cgi?id="

def getField(soup, field):
	href = "%s%s" % ("page.cgi?id=fields.html#", field)
	for i in soup.find_all('th', {"class" : "field_label"}):
                f = i.find('a', {"href" : href})
                if f != None:
                        break
        return f

def getReporter(soup):
	m = getField(soup, "reporter")
	if m == None:
		return ""
	return m.parent.findNext('td').find('span').get_text().encode('utf-8').strip()

def getAssignee(soup):
	m = getField(soup, "assigned_to")
        if m == None:
                return ""
	return m.parent.findNext('td').find('span').get_text().encode('utf-8').strip()

def getStatus(soup):
	m = getField(soup, "bug_status")
        if m == None:
                return ""
        l = m.parent.findNext('td').find('span').get_text().encode('utf-8').strip().split()
	return string.join(l)

def getPriSev(soup):
	m = getField(soup, "priority")
        if m == None:
                return ""
	l = m.findNext('td').get_text().encode('utf-8').split()
	if 'Severity' == l[1]:
		return "%s, %s" % (l[0], l[2])
	return l[0]

for bugId in sys.argv[5:]:
	bugUrl = "%s%s" % (url, bugId)
	soup = BeautifulSoup(urllib2.urlopen(bugUrl))
	output = soup.title.string
	if output == "Access Denied":
		cmd = "echo \'%s 1 PRIVMSG %s :Bug %s is unaccessible\' | nc 127.0.0.1 %s" % (server, channel, bugId, port)
	elif output == "Invalid Bug ID":
		cmd = "echo \'%s 1 PRIVMSG %s :Bug %s is invalid\' | nc 127.0.0.1 %s" % (server, channel, bugId, port)
	else:
		status = getStatus(soup)
		pri = getPriSev(soup)
		assignee = getAssignee(soup)
		reporter = getReporter(soup)
		cmd = "echo \'%s 1 PRIVMSG %s :%s %s, %s, %s, %s, %s\' | nc 127.0.0.1 %s" % (server, channel, bugUrl, output, status, pri, reporter, assignee, port)
	os.system(cmd.encode('utf-8'))
