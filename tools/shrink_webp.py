#!/usr/bin/env python3
import sys, os, io
from PIL import Image
'''
缩小webp大小以节省空间
接收2个参数，第一个为阈值(KB)，第二个为图片所在文件夹
'''
if __name__ == "__main__":
    if len(sys.argv) == 3:
        limit = int(sys.argv[1])
        work_dir = sys.argv[2]
        if os.path.exists(work_dir):
            if work_dir[-1] != '/': work_dir += '/'
            for img_name in os.listdir(work_dir):
                target = work_dir + img_name
                t_sz = os.path.getsize(target) // 1024
                if t_sz > limit:
                    with Image.open(target) as f:
                        with io.BytesIO() as buffer:
                            f.save(buffer, "WEBP")
                            f.close()
                            buffer.seek(0)
                            with open(target, "wb") as out:
                                out.write(buffer.read())
                                print("shrink:", img_name, "from", t_sz, "KB to", os.path.getsize(target) // 1024, "KB")