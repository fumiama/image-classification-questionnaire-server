#!/usr/bin/env python3
import base14, sys, os, imagehash, binascii
from PIL import Image
from hashlib import md5

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
