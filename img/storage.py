from img import get_dhash_b14, get_dhash_b14_io, decode_dhash, hamm_img
from PIL import Image
from io import BytesIO
from glob import glob
from os import path, remove
from json import load, dump
from urllib.request import quote
from shutil import copyfileobj

def save_img(datas: bytes, user_uuid: str, image_dir: str, info_json_path: str) -> dict:
	is_converted = False
	with Image.open(BytesIO(datas)) as img2save:
		if img2save.format != "WEBP":		#转换webp
			converted = BytesIO()
			img2save.save(converted, "WEBP")
			converted.seek(0)
			is_converted = True
	fname = get_dhash_b14_io(converted) if is_converted else get_dhash_b14(datas) 
	no_similar = True
	all_imgs_list = [name[-10:-5] for name in glob(image_dir + "*.webp")]
	this_hash = decode_dhash(fname)
	hash_len = len(this_hash)
	for img_name in all_imgs_list:
		if hamm_img(this_hash, decode_dhash(img_name), hash_len) <= 6:
			no_similar = False
			break
	if no_similar:
		print("Recv file:", fname)
		fn = path.join(image_dir, fname + ".webp")	#生成文件存储路径
		if is_converted: converted.seek(0)
		with open(fn, 'wb') as f: copyfileobj(converted, f) if is_converted else f.write(datas)
		if is_converted: converted.close()
		if path.exists(info_json_path):
			if path.getsize(info_json_path) == 0:
				remove(info_json_path)
			with open(info_json_path, "r") as f:
				info_json = load(f)
				info_json[fname] = user_uuid
		else:
			info_json = {}
			info_json[fname] = user_uuid
		with open(info_json_path, "w") as f:
			dump(info_json, f)
		return {"stat":"success", "img": quote(fname)}
	else: return {"stat":"exist", "img": quote(img_name)}