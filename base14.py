#!/usr/bin/env python3
from ctypes import CDLL, c_void_p, c_char_p, c_uint64, POINTER, Structure, string_at

dllpath = './build/libbase14.so'
dll = CDLL(dllpath)

class LENDAT(Structure):
    _fields_=[('data',c_void_p),
             ('len',c_uint64)]

dll.encode.restype = POINTER(LENDAT)#确定test这个函数的返回值的类型

def get_base14(byte_str):
    byte_len = len(byte_str)
    #print("data length:", byte_len)
    t = dll.encode(byte_str, byte_len)
    encl = t.contents.len
    encd = string_at(t.contents.data, encl)
    #print("encode length:", encl, len(encd))
    #print(encd.decode("utf-16-be"))
    return encd.decode("utf-16-be")
