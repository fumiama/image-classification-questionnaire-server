#!/usr/bin/env python3

import glob, os, threading, sys
from PIL import Image

def create_image(infile, outfolder):
    split_fn = os.path.splitext(os.path.basename(infile))
    with Image.open(infile) as im:
        print("save to", outfolder + split_fn[0] + '_' + split_fn[1][1:] + ".webp")
        im.save(outfolder + split_fn[0] + '_' + split_fn[1][1:] + ".webp", "WEBP")

def from_folder(folder, outfolder):
    for infile in glob.glob(folder + "*.[j p J P][p n P N][e g E G]*"):
        t = threading.Thread(target=create_image, args=[infile, outfolder])
        t.start()

if __name__ == "__main__":
    if len(sys.argv) == 3:
        folder = sys.argv[1]
        if folder[:-1] != '/': folder += '/'
        outfolder = sys.argv[2]
        if outfolder[:-1] != '/': outfolder += '/'
        from_folder(folder, outfolder)
