#!/usr/bin/env python3
import threading
from time import sleep
import sys, os
from urllib3 import PoolManager
sys.path.append('..')
from img import save_img
from base14 import init_dll
from platform import system

'''
从API自动爬取图片并保存
仅有一个参数，即图片保存位置
会将输出重定向到当前目录的log.txt与log_err.txt
'''

#RANDOM_IMG_API = "https://api.pixivweb.com/anime18r.php?return=img"
RANDOM_IMG_API = "https://api.pingping6.com/tools/acg2/index.php"
#RANDOM_IMG_API = "http://www.dmoe.cc/random.php"
THREAD_NUM = 8
DELAY = 8

init_dll('/usr/local/lib/libbase14.' + ('dylib' if system() == 'Darwin' else ('so' if system() == 'Linux' else 'dll') ))

def flush_io() -> None:
	sys.stdout.flush()
	sys.stderr.flush()

# Launch listener threads.
class Thread(threading.Thread):
	def __init__(self, i: int) -> None:
		self.i = i
		self.p = PoolManager()
		#signal(SIGPIPE, SIG_IGN)		# 忽略管道错误
		threading.Thread.__init__(self)
		sleep(DELAY*i)
		print("Thread", i, "start.")
		self.start()
	def run(self) -> None:
		while True:
			r = self.p.request('GET', RANDOM_IMG_API, preload_content=False)
			data = r.read()
			print(save_img(data, "涩酱", image_dir, json_dir))
			r.release_conn()
			sleep(DELAY*THREAD_NUM)

def handle_client() -> None:
	thread_pool = [Thread(i) for i in range(THREAD_NUM)]
	i = THREAD_NUM
	while True:		#监控线程退出情况
		for i in range(len(thread_pool)):
			t = thread_pool[i]
			if not t.is_alive():
				thread_pool[i] = Thread(i)
				print("Thread", i, "dead. Create a new one called", i+1)
				i += 1
		sleep(DELAY*THREAD_NUM)

if __name__ == '__main__':
	if len(sys.argv) == 2:
		image_dir = sys.argv[1]
		json_dir = image_dir + "info.json"
		if os.fork() == 0:		#创建daemon
			os.setsid()
			#创建孙子进程，而后子进程退出
			if os.fork() > 0: sys.exit(0)
			#重定向标准输入流、标准输出流、标准错误
			flush_io()
			si = open("/dev/null", 'r')
			so = open("./log.txt", 'w')
			se = open("./log_err.txt", 'w')
			os.dup2(si.fileno(), sys.stdin.fileno())
			os.dup2(so.fileno(), sys.stdout.fileno())
			os.dup2(se.fileno(), sys.stderr.fileno())
			pid = os.fork()
			while pid > 0:			#监控服务是否退出
				#signal(SIGCHLD, SIG_IGN)
				#signal(SIGPIPE, SIG_IGN)		# 忽略管道错误
				os.wait()
				print("Subprocess exited, restarting...")
				pid = os.fork()
			if pid < 0: print("Fork error!")
			else: handle_client()
		else: print("Creating daemon...")
	else: print("Usage: <image_dir>")
