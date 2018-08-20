import socket
import sys
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect(('172.16.10.1', 8888))

magicword='000102030405060708092828'.decode('hex')

s.send(magicword)
data = s.recv(106) 
n=0
while n<1000: #write replace by while 1 if you want this to not stop
    data = s.recv(1024)
    sys.stdout.write(data)
    n=n+1
s.close()