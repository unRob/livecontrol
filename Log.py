import logging
logging.basicConfig(filename='/Users/rob/ableton.log',level=logging.DEBUG)

class Log():

	@staticmethod
	def info(str):
		logging.info(str)