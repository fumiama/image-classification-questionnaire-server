from PIL import Image
import imagehash, base14, binascii, io

last_char = '㴁'

#base14.init_dll('./build/libbase14.so')

def get_dhash_b14(data):
    return base14.get_base14(binascii.a2b_hex(str(imagehash.dhash(Image.open(io.BytesIO(data))))))[:-1]

def decode_dhash(img_name_no_ext_b14):
    return base14.from_base14((img_name_no_ext_b14 + last_char).encode("utf-16-be"))

def hamm_img(byte_res1, byte_res2):
    """
    汉明距离，汉明距离越小说明越相似，等 0 说明是同一张图片，大于10越上，说明完全不相似
    :param res1:
    :param res2:
    :return:
    """
    str1 = binascii.b2a_hex(byte_res1)
    str2 = binascii.b2a_hex(byte_res2)
    num = 0  # 用来计算汉明距离
    for i in range(len(str1)):
        if str1[i] != str2[i]:
            num += 1
    return num
