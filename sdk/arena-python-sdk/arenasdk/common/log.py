#!/usr/bin/env python 
import logging
import coloredlogs
import os
import sys

class Log(object):
	def __init__(self,logger=None,level='DEBUG'):
		if os.getenv("LOG_LEVEL"):
			level = os.getenv("LOG_LEVEL")
		level,logging_level = parse(level)
		self.logger = logging.getLogger(logger)
		LOG_FORMAT = "%(asctime)s %(levelname)s %(filename)s[line:%(lineno)d] - %(message)s"
		if os.getenv("LOG_WITH_COLOR") and os.getenv("LOG_WITH_COLOR") == "false":
			logging.basicConfig(level=logging_level, format=LOG_FORMAT)
		else:
			coloredlogs.install(
    			level=level,
				logger=self.logger,
    			datefmt='%Y-%m-%d/%H:%M:%S',
    			fmt=LOG_FORMAT
			)	

	def get_logger(self):
		return self.logger

def parse(level):
	level = level.upper()
	logging_level = None
	if level == "DEBUG":
		logging_level = logging.DEBUG
	elif level == "INFO":
		logging_level = logging.INFO
	elif level == "WARNING":
		logging_level = logging.WARNING
	elif level == "ERROR":
		logging_level = logging.ERROR
	elif level == "CRITICAL":
		logging_level = logging.CRITICAL
	else:
		print("Error: unknown log level,supported: {'INFO','WARNING','ERROR','CRITICAL'}")
		sys.exit(1)
	return level,logging_level
