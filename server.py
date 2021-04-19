#!/usr/bin/env python3
from http.server import HTTPServer, BaseHTTPRequestHandler
import time, urllib.request, base14, sys, os, cgi, random, img_diff, threading, socket, io, form_fsm, shutil
from hashlib import md5
from signal import signal, SIGPIPE, SIG_DFL
from PIL import Image

host = ('0.0.0.0', 80)
byte_succ = "succ".encode()
byte_erro = "erro".encode()
byte_null = "null".encode()

base14.init_dll('./build/libbase14.so')

def get_uuid():
	return base14.get_base14(md5(str(time.time()).encode()).digest())[:2]

class Resquest(BaseHTTPRequestHandler):
	def send_200(self, data, content_type):
		self.send_response(200)
		self.send_header('Content-type', content_type)
		self.end_headers()
		self.wfile.write(data)

	def do_GET(self):
		get_path = self.path[1:]
		get_path_len = len(get_path)
		#print("get_path_len:", get_path_len)
		if get_path_len == 6 and get_path == "signup":	# 注册
			new_uuid = get_uuid()
			os.makedirs(user_dir + new_uuid)
			self.send_200(new_uuid.encode("utf-8"), "application/octet-stream")
		elif get_path_len == 0 or (get_path_len == 10 and get_path == "index.html"):
			with open("./index.html", "rb") as f:
				self.send_200(f.read(), "text/html")
		elif get_path_len == 25 and get_path[:6] == "pickdl":
			user_uuid = urllib.request.unquote(get_path[7:])
			if len(user_uuid) == 2:		#base14检测
				user_path = user_dir + user_uuid +'/'
				#print("User dir:", user_path)
				if os.path.exists(user_path):
					voted_imgs_list = os.listdir(user_path)
					all_imgs_list = [name[:-5] for name in os.listdir(image_dir)]
					all_imgs_len = len(all_imgs_list)
					if len(voted_imgs_list) < all_imgs_len:
						pick_img_name = all_imgs_list[random.randint(0, all_imgs_len-1)]
						while pick_img_name in voted_imgs_list:
							pick_img_name = all_imgs_list[random.randint(0, all_imgs_len-1)]
						img_path = image_dir + pick_img_name + ".webp"
						try:
							with open(img_path, "rb") as f:
								self.send_200(f.read(), "image/webp")
						except: self.send_200(byte_erro, "text/plain")
					else: self.send_200(byte_null, "text/plain")
				else: self.send_200(byte_erro, "text/plain")
			else: self.send_200(byte_erro, "text/plain")
		elif get_path_len == 23 and get_path[:4] == "pick":
			user_uuid = urllib.request.unquote(get_path[5:])
			if len(user_uuid) == 2:		#base14检测
				user_path = user_dir + user_uuid +'/'
				#print("User dir:", user_path)
				if os.path.exists(user_path):
					voted_imgs_list = os.listdir(user_path)
					all_imgs_list = [name[:-5] for name in os.listdir(image_dir)]
					all_imgs_len = len(all_imgs_list)
					if len(voted_imgs_list) < all_imgs_len:
						pick_img_name = all_imgs_list[random.randint(0, all_imgs_len-1)]
						while pick_img_name in voted_imgs_list:
							pick_img_name = all_imgs_list[random.randint(0, all_imgs_len-1)]
						self.send_200(urllib.request.quote(pick_img_name).encode(), "text/plain")
					else: self.send_200(byte_null, "text/plain")
				else: self.send_200(byte_erro, "text/plain")
			else: self.send_200(byte_erro, "text/plain")
		elif get_path_len >= 72:		# 投票
			if get_path_len > 4 and get_path[:4] == "vote":
				try:
					cli_req = get_path[5:]
					cli_uuid = urllib.request.unquote(cli_req[5:23])
					if len(cli_uuid) == 2:			#base14检测
						cli_img = urllib.request.unquote(cli_req[28:73])
						if len(cli_img) == 5:		#base14检测
							cli_cls = cli_req[80:]
							print("uuid:", cli_uuid, "img:", cli_img, "class:", cli_cls)
							cli_dir = user_dir + cli_uuid + '/'
							#os.makedirs(cli_dir, exist_ok=True)
							with open(cli_dir + cli_img, "w") as f: f.write(cli_cls)
							self.send_200(byte_succ, "text/plain")
						else: self.send_200(byte_erro, "text/plain")
					else: self.send_200(byte_erro, "text/plain")
				except: self.send_200(byte_erro, "text/plain")
		elif get_path_len == 45:
			target_img_name = urllib.request.unquote(get_path)
			if len(target_img_name) == 5:		#base14检测
				img_path = image_dir + target_img_name + ".webp"
				#print("Get img:", img_path)
				if os.path.exists(img_path):
					try:
						with open(img_path, "rb") as f:
							self.send_200(f.read(), "image/webp")
					except: self.send_200(byte_erro, "text/plain")
				else: self.send_200(byte_null, "text/plain")
			else: self.send_200(byte_erro, "text/plain")
		else: self.send_200(byte_null, "text/plain")
	
	def do_POST(self):
		if self.path == "/upload":			#上传图片
			self.save_img(self.rfile.read(int(self.headers.get('content-length'))))
		elif self.path == "/upform":		#表单上传图片
			size = int(self.headers.get('content-length'))
			skip = 0
			if size > 1024:
				state = 0
				while skip < 1024:
					skip += 1
					state = form_fsm.scan(state, self.rfile.read(1))
					if state == 11:
						skip += 3
						self.rfile.read(3)
						self.do_form_post(size, skip)
						break
		else: self.send_200(byte_null, "text/plain")

	def do_form_post(self, size, skip):
		skip += 9
		file_type = self.rfile.read(9).decode()
		print("post form type:", file_type)
		if file_type == "image/web" or file_type == "image/png" or file_type == "image/jpe":
			if file_type == "image/png":
				skip += 4
				self.rfile.read(4)
			else:
				skip += 5
				self.rfile.read(5)
			datas = self.rfile.read(size - skip - 46)		#掐头去尾
			self.save_img(datas)
		else: self.send_200(byte_erro, "text/plain")
	
	def save_img(self, datas):
		is_converted = False
		with Image.open(io.BytesIO(datas)) as img2save:
			if img2save.format != "WEBP":		#转换webp
				converted = io.BytesIO()
				img2save.save(converted, "WEBP")
				converted.seek(0)
				is_converted = True
		fname = img_diff.get_dhash_b14_io(converted) if is_converted else img_diff.get_dhash_b14(datas) 
		no_similar = True
		all_imgs_list = os.listdir(image_dir)
		this_hash = img_diff.decode_dhash(fname)
		hash_len = len(this_hash)
		for img_name in all_imgs_list:
			if img_diff.hamm_img(this_hash, img_diff.decode_dhash(img_name), hash_len) <= 12:
				no_similar = False
				break
		if no_similar:
			print("Recv file:", fname)
			fn = os.path.join(image_dir, fname + ".webp")	#生成文件存储路径
			if not os.path.exists(fn):
				if is_converted: converted.seek(0)
				with open(fn, 'wb') as f: shutil.copyfileobj(converted, f) if is_converted else f.write(datas)
				if is_converted: converted.close()
				self.send_200(byte_succ, "text/plain")
			else: self.send_200(byte_erro, "text/plain")
		else:  self.send_200(byte_null, "text/plain")

# Launch 100 listener threads.
class Thread(threading.Thread):
	def __init__(self, i):
		threading.Thread.__init__(self)
		self.i = i
		signal(SIGPIPE, SIG_DFL)		# 忽略管道错误
		self.daemon = True
		self.start()
	def run(self):
		self.httpd = HTTPServer(host, Resquest, False)
		# Prevent the HTTP server from re-binding every handler.
		# https://stackoverflow.com/questions/46210672/
		self.httpd.socket = sock
		self.httpd.server_bind = self.server_close = lambda self: None
		self.httpd.serve_forever()

if __name__ == '__main__':
	if len(sys.argv) == 3:
		user_dir = sys.argv[1]
		image_dir = sys.argv[2]
		if user_dir[-1] != '/': user_dir += '/'
		if os.path.exists(image_dir):
			if image_dir[-1] != '/': image_dir += '/'
			print("Starting ICQS at: %s:%s" % host, "storage dir:", user_dir, "image dir:", image_dir)
			# Create ONE socket.
			sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
			sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
			sock.bind(host)
			sock.listen(5)
			[Thread(i) for i in range(100)]
			#主进程也开启一个服务
			signal(SIGPIPE, SIG_DFL)		# 忽略管道错误
			httpd = HTTPServer(host, Resquest, False)
			httpd.socket = sock
			httpd.server_bind = lambda self: None
			httpd.serve_forever()
		else: print("Error: image dir", image_dir, "is not exist.")
	else: print("Usage:", sys.argv[0], "<user_dir> <image_dir>")