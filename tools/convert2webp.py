#!/usr/bin/env python3
import glob, os, threading, sys, time
from PIL import Image
'''
多线程批量转换某文件夹下的图片到webp
第一个参数是待转换图片的文件夹，第二个参数是输出文件夹
待转换图片的文件夹内可以有其他文件
'''
def create_image(infile, outfolder):
    split_fn = os.path.splitext(os.path.basename(infile))
    with Image.open(infile) as im:
        print("save to", outfolder + split_fn[0] + '_' + split_fn[1][1:] + ".webp")
        im.save(outfolder + split_fn[0] + '_' + split_fn[1][1:] + ".webp", "WEBP")

def from_folder(folder, outfolder):
    for infile in glob.glob(folder + "*.[j p J P][p n P N][e g E G]*"):
        t = threading.Thread(target=create_image, args=[infile, outfolder])
        t.start()
        time.sleep(0.023)

if __name__ == "__main__":
    if len(sys.argv) == 3:
        folder = sys.argv[1]
        if folder[:-1] != '/': folder += '/'
        outfolder = sys.argv[2]
        if outfolder[:-1] != '/': outfolder += '/'
        from_folder(folder, outfolder)
