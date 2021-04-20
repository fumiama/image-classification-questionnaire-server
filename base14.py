#!/usr/bin/env python3
from ctypes import CDLL, c_void_p, c_char_p, c_uint64, c_uint32, POINTER, Structure, string_at
from platform import machine
from os import name, environ

#dllpath = './build/libbase14.so'
#dll = CDLL(dllpath)

global dll

def init_dll(dll_pth):
    global dll
    dll = CDLL(dll_pth)
    dll.encode.restype = POINTER(LENDAT)
    dll.decode.restype = POINTER(LENDAT)

def this_machine():
    """Return type ofmachine."""
    if name == 'nt' and sys.version_info[:2] < (2,7):
        return environ.get("PROCESSOR_ARCHITEW6432", environ.get('PROCESSOR_ARCHITECTURE',''))
    else: return machine()

def os_bits(machine=this_machine()):
    """Return bitness ofoperating system, or None if unknown."""
    machine2bits = {'AMD64':64, 'x86_64': 64, 'i386': 32, 'x86': 32}
    return machine2bits.get(machine, None)

class LENDAT(Structure):
    _fields_=[('data', c_void_p),
             ('len', c_uint64 if os_bits() == 64 else c_uint32)]

def get_base14(byte_str):
    global dll
    byte_len = len(byte_str)
    #print("data length:", byte_len)
    t = dll.encode(byte_str, byte_len)
    encl = t.contents.len
    encd = string_at(t.contents.data, encl)
    #print("encode length:", encl, len(encd))
    #print(encd.decode("utf-16-be"))
    return encd.decode("utf-16-be")

def from_base14(utf16be_byte_str):
    global dll
    byte_len = len(utf16be_byte_str)
    t = dll.decode(utf16be_byte_str, byte_len)
    decl = t.contents.len
    decd = string_at(t.contents.data, decl)
    return decd
