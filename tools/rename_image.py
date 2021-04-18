#!/usr/bin/env python3
import base14, sys, os, imagehash, binascii
from PIL import Image
from hashlib import md5
'''
以和服务器命名规则相同的格式批量重命名给定文件夹下的所有文件
该文件下只能存在图片文件，否则会报错
'''
base14.init_dll('../build/libbase14.so')

if __name__ == "__main__":
    if len(sys.argv) == 2:
        work_dir = sys.argv[1]
        if os.path.exists(work_dir):
            if work_dir[-1] != '/': work_dir += '/'
            for img_name in os.listdir(work_dir):
                with Image.open(work_dir + img_name) as f:
                    new_img_name = base14.get_base14(binascii.a2b_hex(str(imagehash.dhash(f))))[:-1] + '.' + img_name.split('.')[-1]
                    print("New img name:", new_img_name)
                    os.rename(work_dir + img_name, work_dir + new_img_name)
