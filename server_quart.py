#!/usr/bin/env python3
from quart import Quart, request, Response
from random import randint
from io import BytesIO
from shutil import copyfileobj
from urllib.request import quote, unquote
from time import time, sleep
from hashlib import md5
from PIL import Image
import base14, sys, os, img_diff, form_fsm, json

host = ('0.0.0.0', 80)
app = Quart(__name__)

base14.init_dll('./build/libbase14.so')

def get_uuid() -> str:
	return base14.get_base14(md5(str(time()).encode()).digest())[:2]

def flush_io() -> None:
	sys.stdout.flush()
	sys.stderr.flush()

@app.route("/")
@app.route("/index.html")
def index():
	with open("./index_quart.html") as f:
		r = f.read()
	return r

@app.route("/signup")
def signup():
	try:
		diff = int(time()) - (int(request.args.get("key")) ^ pwd)
		if diff < 10 and diff >= 0:		#验证通过
			new_uuid = get_uuid()
			os.makedirs(user_dir + new_uuid)
			return {"stat":"success", "id":quote(new_uuid)}
		else: return {"stat":"wrong", "id":"null"}
	except: return {"stat":"error", "id":"null"}

@app.route("/vote")
def vote():
	try:
		cli_uuid = unquote(request.args.get("uuid"))
		if len(cli_uuid) == 2:			#base14检测
			cli_img = unquote(request.args.get("img"))
			if len(cli_img) == 5:		#base14检测
				cli_cls = request.args.get("class")
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
			all_imgs_list = [name[:-5] for name in os.listdir(image_dir)]
			all_imgs_len = len(all_imgs_list)
			if len(voted_imgs_list) < all_imgs_len:
				pick_img_name = all_imgs_list[randint(0, all_imgs_len-1)]
				while pick_img_name in voted_imgs_list:
					pick_img_name = all_imgs_list[randint(0, all_imgs_len-1)]
				if send_name_only:
					if os.path.exists(info_json_path):
						if os.path.getsize(info_json_path) == 0:
							os.remove(info_json_path)
						with open(info_json_path, "r") as f:
							info_json = json.load(f)
						if pick_img_name in info_json.keys():
							uploader = info_json[pick_img_name]
						else: uploader = "系统"
					else: uploader = "系统"
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
	return do_pick(unquote(request.args.get("uuid")), False)

@app.route("/pick")
def pick():
	return do_pick(unquote(request.args.get("uuid")), True)

@app.route("/img")
def img():
	target_img_name = unquote(request.args.get("path"))
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

def save_img(datas: bytes, user_uuid: str) -> dict:
	is_converted = False
	with Image.open(BytesIO(datas)) as img2save:
		if img2save.format != "WEBP":		#转换webp
			converted = BytesIO()
			img2save.save(converted, "WEBP")
			converted.seek(0)
			is_converted = True
	fname = img_diff.get_dhash_b14_io(converted) if is_converted else img_diff.get_dhash_b14(datas) 
	no_similar = True
	all_imgs_list = os.listdir(image_dir)
	this_hash = img_diff.decode_dhash(fname)
	hash_len = len(this_hash)
	for img_name in all_imgs_list:
		if img_diff.hamm_img(this_hash, img_diff.decode_dhash(img_name), hash_len) <= 6:
			no_similar = False
			break
	if no_similar:
		print("Recv file:", fname)
		fn = os.path.join(image_dir, fname + ".webp")	#生成文件存储路径
		if is_converted: converted.seek(0)
		with open(fn, 'wb') as f: copyfileobj(converted, f) if is_converted else f.write(datas)
		if is_converted: converted.close()
		if os.path.exists(info_json_path):
			if os.path.getsize(info_json_path) == 0:
				os.remove(info_json_path)
			with open(info_json_path, "r") as f:
				info_json = json.load(f)
				info_json[fname] = user_uuid
		else:
			info_json = {}
			info_json[fname] = user_uuid
		with open(info_json_path, "w") as f:
			json.dump(info_json, f)
		return {"stat":"success"}
	else: return {"stat":"exist"}

@app.route("/upload", methods=['POST'])
async def upload():
	cli_uuid = unquote(request.args.get("uuid"))
	print("post from:", cli_uuid)
	if len(cli_uuid) == 2:
		if os.path.exists(user_dir + cli_uuid):
			return save_img(await request.get_data(), cli_uuid)
		else: return {"stat":"noid"}
	else: return {"stat":"invid"}

def do_form_post(f, size: int, skip: int, cli_uuid: str) -> dict:
	skip += 9
	file_type = f.read(9).decode()
	print("post form type:", file_type)
	if file_type == "image/web" or file_type == "image/png" or file_type == "image/jpe":
		if file_type == "image/png":
			skip += 4
			f.read(4)
		else:
			skip += 5
			f.read(5)
		datas = f.read(size - skip - 46)		#掐头去尾
		return save_img(datas, cli_uuid)
	else: return {"stat":"typeerr"}

@app.route("/upform", methods=['POST'])
async def upform():
	cli_uuid = unquote(request.args.get("uuid"))
	print("post from:", cli_uuid)
	datas = await request.get_data()
	if len(cli_uuid) == 2:
		if os.path.exists(user_dir + cli_uuid):
			size = len(datas)
			skip = 0
			if size > 1024:
				with BytesIO(datas) as rfile:
					rfile.seek(0, 0)
					state = 0
					while skip < 1024:
						skip += 1
						state = form_fsm.scan(state, rfile.read(1)[0])
						if state == 11:
							skip += 3
							rfile.read(3)
							return do_form_post(rfile, size, skip, cli_uuid)
							break
		else: return {"stat":"noid"}
	else: return {"stat":"invid"}

@app.before_first_request
async def setuid():
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
			app.run(host[0], host[1])
		else: print("Error: image dir", image_dir, "is not exist.")
	else: print("Usage:", sys.argv[0], "<user_dir> <image_dir> <pwd_path> (server_uid)")