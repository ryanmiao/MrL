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

def getReporter(soup):
	for i in soup.find_all('th', {"class":"field_label"}):
		m = i.find('a', {"href":"page.cgi?id=fields.html#reporter"})
		if m != None:
			break
	if m == None:
		return ""
	return m.parent.findNext('td').find('span').get_text().encode('utf-8').strip()

def getAssignee(soup):
	for i in soup.find_all('th', {"class":"field_label"}):
                m = i.find('a', {"href":"page.cgi?id=fields.html#assigned_to"})
                if m != None:
                        break
        if m == None:
                return ""
	return m.parent.findNext('td').find('span').get_text().encode('utf-8').strip()

def getStatus(soup):
	for i in soup.find_all('th', {"class":"field_label"}):
                m = i.find('a', {"href":"page.cgi?id=fields.html#bug_status"})
                if m != None:
                        break
        if m == None:
                return ""
        l = m.parent.findNext('td').find('span').get_text().encode('utf-8').strip().split()
	return string.join(l)

def getPriSev(soup):
	for i in soup.find_all('th', {"class":"field_label"}):
                m = i.find('a', {"href":"page.cgi?id=fields.html#priority"})
                if m != None:
                        break
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
		cmd = "echo \'%s 1 PRIVMSG %s :Bug %s is unaccessible\' | nc 127.0.0.1 %s" % (server, user, bugId, port)
	elif output == "Invalid Bug ID":
		cmd = "echo \'%s 1 PRIVMSG %s :Bug %s is invalid\' | nc 127.0.0.1 %s" % (server, user, bugId, port)
	else:
		status = getStatus(soup)
		pri = getPriSev(soup)
		assignee = getAssignee(soup)
		reporter = getReporter(soup)
		cmd = "echo \'%s 1 PRIVMSG %s :%s %s, %s, %s, %s, %s\' | nc 127.0.0.1 %s" % (server, user, bugUrl, output, status, pri, reporter, assignee, port)
	os.system(cmd.encode('utf-8'))
