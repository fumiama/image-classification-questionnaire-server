#!/usr/bin/env python3
from http.server import HTTPServer, BaseHTTPRequestHandler
import time, urllib.request, base14, sys, os, cgi, random
from hashlib import md5

host = ('localhost', 8847)
byte_succ = "succ".encode()
byte_erro = "erro".encode()
byte_null = "null".encode()

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
		elif get_path_len == 10 and get_path == "index.html":
			with open("./index.html", "rb") as f:
				self.send_200(f.read(), "text/html")
		elif get_path_len == 25 and get_path[:6] == "pickdl":
			user_path = user_dir + urllib.request.unquote(get_path[7:]) +'/'
			#print("User dir:", user_path)
			if os.path.exists(user_path):
				voted_imgs_list = os.listdir(user_path)
				all_imgs_list = os.listdir(image_dir)
				all_imgs_len = len(all_imgs_list)
				if len(voted_imgs_list) < all_imgs_len:
					pick_img_name = all_imgs_list[random.randint(0, all_imgs_len)]
					while pick_img_name in voted_imgs_list:
						pick_img_name = all_imgs_list[random.randint(0, all_imgs_len)]
					img_path = image_dir + pick_img_name
					try:
						with open(img_path, "rb") as f:
							self.send_200(f.read(), "image/webp")
					except: self.send_200(byte_erro, "text/plain")
				else: self.send_200(byte_null, "text/plain")
			else: self.send_200(byte_erro, "text/plain")
		elif get_path_len == 23 and get_path[:4] == "pick":
			user_path = user_dir + urllib.request.unquote(get_path[5:]) +'/'
			#print("User dir:", user_path)
			if os.path.exists(user_path):
				voted_imgs_list = os.listdir(user_path)
				all_imgs_list = os.listdir(image_dir)
				all_imgs_len = len(all_imgs_list)
				if len(voted_imgs_list) < all_imgs_len:
					pick_img_name = all_imgs_list[random.randint(0, all_imgs_len)]
					while pick_img_name in voted_imgs_list:
						pick_img_name = all_imgs_list[random.randint(0, all_imgs_len)]
					self.send_200(urllib.request.quote(pick_img_name[:-5]).encode(), "text/plain")
				else: self.send_200(byte_null, "text/plain")
			else: self.send_200(byte_erro, "text/plain")
		elif get_path_len >= 72:		# 投票
			if get_path_len > 4 and get_path[:4] == "vote":
				try:
					cli_req = get_path[5:]
					cli_uuid = urllib.request.unquote(cli_req[5:23])
					cli_img = urllib.request.unquote(cli_req[28:64])
					cli_cls = cli_req[71:]
					print("uuid:", cli_uuid, "img:", cli_img, "class:", cli_cls)
					cli_dir = user_dir + cli_uuid + '/'
					os.makedirs(cli_dir, exist_ok=True)
					with open(cli_dir + cli_img, "w") as f: f.write(cli_cls)
					self.send_200(byte_succ, "text/plain")
				except: self.send_200(byte_erro, "text/plain")
		elif get_path_len == 36:
			img_path = image_dir + urllib.request.unquote(get_path) + ".webp"
			#print("Get img:", img_path)
			if os.path.exists(img_path):
				try:
					with open(img_path, "rb") as f:
						self.send_200(f.read(), "image/webp")
				except: self.send_200(byte_erro, "text/plain")
			else: self.send_200(byte_null, "text/plain")
		else: self.send_200(byte_null, "text/plain")
	
	def do_POST(self):
		if self.path == "/upload":							#上传图片
			datas = self.rfile.read(int(self.headers.get('content-length')))
			fname = base14.get_base14(md5(datas).digest())[:4] + ".webp"
			print("Recv file:", fname)
			fn = os.path.join(image_dir, fname)				#生成文件存储路径
			if not os.path.exists(fn):
				with open(fn, 'wb') as f: f.write(datas)	#将接收到的内容写入文件
				self.send_200(byte_succ, "text/plain")
			else: self.send_200(byte_erro, "text/plain")

if __name__ == '__main__':
	if len(sys.argv) == 3:
		server = HTTPServer(host, Resquest)
		user_dir = sys.argv[1]
		image_dir = sys.argv[2]
		if user_dir[-1] != '/': user_dir += '/'
		if os.path.exists(image_dir):
			if image_dir[-1] != '/': image_dir += '/'
			print("Starting ICQS at: %s:%s" % host, "storage dir:", user_dir, "image dir:", image_dir)
			server.serve_forever()
		else: print("Error: image dir", image_dir, "is not exist.")