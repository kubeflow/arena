#!/usr/bin/env python
import yaml
import json
import os
import sys
import traceback
import arenasdk.common.log as log
import subprocess, threading

logger = log.Log(__name__).get_logger()

class Command(object):
	def __init__(self, *cmd):
		self.cmd = cmd
		self.process = None

	def run(self,timeout_sec=-1):
		cmdstr = ' '.join(self.cmd)
		logger.debug("execute command: [%s]",cmdstr)
		if timeout_sec == -1 or timeout_sec == 0:
			return self.run_without_timeout(cmdstr)
		return self.run_with_timeout(cmdstr,timeout_sec)

	def run_with_communicate(self,accepter):
		cmdstr = ' '.join(self.cmd)
		logger.debug("execute command: [%s]",cmdstr)
		cmd = subprocess.Popen(
				cmdstr, 
				stdin=subprocess.PIPE, 
				stderr=accepter, 
				close_fds=True,
				stdout=accepter, 
				universal_newlines=True, 
				shell=True, bufsize=1
		)
		cmd.communicate()
		return cmd.returncode,"",""

	def run_without_timeout(self,cmdstr):
		self.process = subprocess.Popen(cmdstr,shell=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
		out = self.process.communicate()
		stdout = out[0].decode('utf-8')
		stderr = out[1].decode('utf-8')
		return_code = self.process.returncode
		return return_code,stdout,stderr

	def run_with_timeout(self,cmdstr,timeout_sec):
		self.process = subprocess.Popen(cmdstr,shell=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
		try:
			out = self.process.communicate(timeout=timeout_sec)
			stdout = out[0].decode('utf-8')
			stderr = out[1].decode('utf-8')
			return_code = self.process.returncode
			return return_code,stdout,stderr

		except subprocess.TimeoutExpired:
			self.process.kill()
			err_msg = "timeout for executing command(timeout="+ str(timeout_sec) + "s): ["	+ cmdstr + "]"	
			return 124,"",err_msg 
		
def read_yaml_or_json_file(target_file,fmt=""):
	try:
		if fmt == "":
			name,ext = os.path.splitext(target_file)
			fmt = ext
		with open(target_file,"r") as f:
			if fmt == ".yaml" or fmt == ".yml":
				return 0,yaml.load(f,Loader=yaml.FullLoader)
			elif fmt == ".json":
				return 0,json.load(f)
			else:
				return 1,"unknown file format,it should be end with  .json or .yaml"
	except Exception as err:
		err_log = traceback.format_exc()
		return 2,err_log
