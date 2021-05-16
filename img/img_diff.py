from PIL import Image
from imagehash import dhash
from binascii import a2b_hex
from io import BytesIO
from base14 import get_base14, from_base14
from numba import jit

last_char = '㴁'

#base14.init_dll('./build/libbase14.so')

def get_dhash_b14(datas: bytes) -> str:
    b14_dhash = "0" * 16
    with BytesIO(datas) as dataio:
        with Image.open(dataio) as img:
            b14_dhash = get_base14(a2b_hex(str(dhash(img))))[:-1]
    return b14_dhash

def get_dhash_b14_io(dataio: BytesIO) -> str:
    b14_dhash = "0" * 16
    with Image.open(dataio) as img:
        b14_dhash = get_base14(a2b_hex(str(dhash(img))))[:-1]
    return b14_dhash

def decode_dhash(img_name_no_ext_b14: str) -> bytes:
    return from_base14((img_name_no_ext_b14 + last_char).encode("utf-16-be"))

@jit
def hamm_img(byte_res1: bytes, byte_res2: bytes, length: int) -> int:
    """
    汉明距离，汉明距离越小说明越相似，等 0 说明是同一张图片，大于40说明完全不相似
    """
    num = 0  # 用来计算汉明距离
    for i in range(length):
        a, b = byte_res1[i], byte_res2[i]
        n = 8
        while n:
            if a % 2 != b % 2: num += 1
            a //= 2
            b //= 2
            n -= 1
    return num
