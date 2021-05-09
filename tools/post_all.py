#!/usr/bin/env python3
import os, sys, stat
'''
生成unix下的脚本文件实现批量post图片
第一个参数为图片所在文件夹，第二个参数为post url
'''
if __name__ == "__main__":
    if len(sys.argv) == 3:
        parent_dir = sys.argv[1]
        if parent_dir[-1] != '/': parent_dir += '/'
        files = os.listdir(parent_dir)
        post_url = sys.argv[2]
        #print(files)
        with open("./post_all.sh", "w") as f:
            f.write("#!/usr/bin/env bash\n")
            for file_name in files:
                if len(file_name) >= 5 and file_name[-5:] == ".webp":
                    f.write("wget -qO- --post-file=" + parent_dir + file_name + ' ' + post_url + '\n')
        os.chmod("./post_all.sh", stat.S_IRWXU)
        os.system("./post_all.sh")
    else: print("usage: <img_floder> <post_url>")
