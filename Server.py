import subprocess
import microjson
import socket
import select
import threading
import signal
import os
import time
from Log import Log

class Server():

    def __init__(self):
        self.live = False
        try:
            self._start_socket(1111)
        except Exception, err:
            Log.info(err)
            Log.info("Accept called")
        self._callbacks = {}
        port = self._open_port()
        path = os.path.dirname(os.path.realpath(__file__))
        full_path = '/'.join([path, 'livecontrol'])
        self.server = subprocess.Popen([full_path, str(port)])
        Log.info("Server started / pid: "+str(self.server.pid)+" port: "+str(port))
        time.sleep(10)
        self._start_socket(self._open_port())



    def _open_port(self):
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.bind(("",0))
        s.listen(1)
        port = s.getsockname()[1]
        s.close()
        return 1234
        return port

    def on(self, evt, callback):
        self._callbacks[evt] = callback

    def off(self, evt):
        del self._callbacks[evt]

    def send(self, evt, data={}):
        if self.live == False:
            return

        message = microjson.to_json({"evt": evt, "data": data})
        Log.info("Sending message: "+message)
        self._socket.send(message+"\n")


    def read(self):
        if self.live:
            # Log.info("reading")
            try:
                msg = self._socket.recv(4096)
            except socket.timeout, e:
                Log.info(e)
            except socket.error, e:
                return
            else:
                Log.info("Received: "+msg)
                data = microjson.from_json(msg)
                evt = data["evt"]
                if evt in self._callbacks:
                    self._callbacks[evt](data["data"])
                else:
                    Log.info("No supe que hacer con "+evt)

    def _start_socket(self, port):
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.connect(('', port))
        self._socket.setblocking(0)

        self.live = True;
        Log.info("Server started")




    def stop(self):
        Log.info("Close")
        os.kill(self.server.pid, signal.SIGTERM)
        self._socket.close()
