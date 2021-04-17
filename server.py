#!/usr/bin/env python3
from http.server import HTTPServer, BaseHTTPRequestHandler
import time, urllib.request, base14
from hashlib import md5

host = ('localhost', 8847)

def get_uuid():
    return base14.get_base14(md5(str(time.time()).encode()).digest())[:2]

class Resquest(BaseHTTPRequestHandler):
    def send_200(self, data, content_type):
        self.send_response(200)
        self.send_header('Content-type', content_type)
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        get_path = self.path[1:]
        get_path_len = len(get_path)
        if get_path_len == 6 and get_path == "signup":
            self.send_200(get_uuid().encode("utf-8"), "application/octet-stream")
        if get_path_len >= 47:
            if get_path_len > 4 and get_path[:4] == "vote":
                cli_req = get_path[5:]
                cli_uuid = urllib.request.unquote(cli_req[5:23])
                cli_img = cli_req[28:34]
                cli_cls = cli_req[41:]
                print("req:", cli_req, "uuid:", cli_uuid, "img:", cli_img, "class:", cli_cls)

if __name__ == '__main__':
    server = HTTPServer(host, Resquest)
    print("Starting ICQS, listen at: %s:%s" % host)
    server.serve_forever()