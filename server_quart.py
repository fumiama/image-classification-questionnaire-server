#!/usr/bin/env python3
from quart import Quart, request, Response
from random import randint, choice
from urllib.request import quote, unquote
from time import time
from hashlib import md5
from glob import glob
import sys, os, json
from base14 import init_dll, get_base14
from img import save_img
from platform import system

host = ('0.0.0.0', 8080)
app = Quart(__name__)

init_dll('/usr/local/lib/libbase14.' + ('dylib' if system() == 'Darwin' else ('so' if system() == 'Linux' else 'dll') ))

def get_uuid() -> str:
	return get_base14(md5(str(time()).encode()).digest())[:2]

def get_arg(key: str) -> str:
	return request.args.get(key)

@app.route("/")
@app.route("/index.html")
def index() -> bytes:
	with open("./index_quart.html") as f:
		r = f.read()
	return r

@app.route("/signup")
def signup() -> dict:
	try:
		diff = int(time()) - (int(get_arg("key")) ^ pwd)
		if diff < 10 and diff >= 0:		#验证通过
			new_uuid = get_uuid()
			os.makedirs(user_dir + new_uuid)
			return {"stat":"success", "id":quote(new_uuid)}
		else: return {"stat":"wrong", "id":"null"}
	except: return {"stat":"error", "id":"null"}

@app.route("/vote")
def vote() -> dict:
	try:
		cli_uuid = unquote(get_arg("uuid"))
		if len(cli_uuid) == 2:			#base14检测
			cli_img = unquote(get_arg("img"))
			if len(cli_img) == 5:		#base14检测
				cli_cls = get_arg("class")
				print("uuid:", cli_uuid, "img:", cli_img, "class:", cli_cls)
				cli_dir = user_dir + cli_uuid + '/'
				#os.makedirs(cli_dir, exist_ok=True)
				with open(cli_dir + cli_img, "w") as f: f.write(cli_cls)
				return {"stat":"success"}
			else: return {"stat":"invimg"}
		else: return {"stat":"invid"}
	except: return {"stat":"noid"}

def do_pick(user_uuid: str, send_name_only: bool):
	if len(user_uuid) == 2:		#base14检测
		user_path = user_dir + user_uuid +'/'
		#print("User dir:", user_path)
		if os.path.exists(user_path):
			voted_imgs_list = os.listdir(user_path)
			all_imgs_list = [name[-10:-5] for name in glob(image_dir + "*.webp")]
			all_imgs_len = len(all_imgs_list)
			if len(voted_imgs_list) < all_imgs_len:
				if len(voted_imgs_list) <= all_imgs_len*8//10:
					pick_img_name = all_imgs_list[randint(0, all_imgs_len-1)]
					while pick_img_name in voted_imgs_list:
						pick_img_name = all_imgs_list[randint(0, all_imgs_len-1)]
				else:
					pick_img_name = choice(list(
						set(all_imgs_list).difference(set(voted_imgs_list))))
				if send_name_only:
					if os.path.exists(info_json_path):
						if os.path.getsize(info_json_path) == 0:
							os.remove(info_json_path)
						try:
							with open(info_json_path, "r") as f:
								info_json = json.load(f)
							if pick_img_name in info_json.keys():
								uploader = info_json[pick_img_name]
							else: uploader = "涩酱"
						except: uploader = "涩酱"
					else: uploader = "涩酱"
					return {"stat":"success", "img":quote(pick_img_name), "uploader":quote(uploader)}
				else:
					img_path = image_dir + pick_img_name + ".webp"
					try:
						with open(img_path, "rb") as f:
							return Response(f.read(), content_type="image/webp")
					except: return {"stat":"readimgerr"}
			else: return {"stat":"nomoreimg"}
		else: return {"stat":"noid"}
	else: return {"stat":"invid"}

@app.route("/pickdl")
def pickdl():
	return do_pick(unquote(get_arg("uuid")), False)

@app.route("/pick")
def pick():
	return do_pick(unquote(get_arg("uuid")), True)

@app.route("/img")
def img():
	target_img_name = unquote(get_arg("path"))
	if len(target_img_name) == 5:		#base14检测
		img_path = image_dir + target_img_name + ".webp"
		#print("Get img:", img_path)
		if os.path.exists(img_path):
			try:
				with open(img_path, "rb") as f:
					return Response(f.read(), content_type="image/webp")
			except: return {"stat":"readimgerr"}
		else: return {"stat":"nosuchimg"}
	else: return {"stat":"invimg"}

@app.route("/upload", methods=['POST'])
async def upload() -> dict:
	cli_uuid = unquote(get_arg("uuid"))
	print("post from:", cli_uuid)
	if len(cli_uuid) == 2:
		if os.path.exists(user_dir + cli_uuid):
			return save_img(await request.get_data(), cli_uuid, image_dir, info_json_path)
		else: return {"stat":"noid"}
	else: return {"stat":"invid"}

@app.route("/upform", methods=['POST'])
async def upform() -> dict:
	cli_uuid = unquote(get_arg("uuid"))
	print("post from:", cli_uuid)
	if len(cli_uuid) == 2:
		if os.path.exists(user_dir + cli_uuid):
			re = []
			for f in (await request.files).getlist("img"):
				re.append({"name":f.filename, **save_img(f.read(), cli_uuid, image_dir, info_json_path)})
			return {"result": re}
		else: return {"stat":"noid"}
	else: return {"stat":"invid"}

@app.before_first_request
async def setuid() -> None:
	if server_uid > 0: os.setuid(server_uid)		#监听后降权

if __name__ == '__main__':
	if len(sys.argv) == 4 or len(sys.argv) == 5:
		user_dir = sys.argv[1]
		image_dir = sys.argv[2]
		pwd_file = sys.argv[3]
		server_uid = int(sys.argv[4]) if len(sys.argv) == 5 else -1
		with open(pwd_file, "rb") as f:
			pwd = int.from_bytes(f.read()[2:], byteorder="big")		#两个汉字，四个字节
		if user_dir[-1] != '/': user_dir += '/'
		if os.path.exists(image_dir):
			if image_dir[-1] != '/': image_dir += '/'
			info_json_path = image_dir + "info.json"
			print("Starting ICQS at: %s:%s" % host, "storage dir:", user_dir, "image dir:", image_dir)
			app.run(*host)
		else: print("Error: image dir", image_dir, "is not exist.")
	else: print("Usage: <user_dir> <image_dir> <pwd_path> (server_uid)")
