import sys
import os
import robot

path = os.path.abspath(os.path.dirname(__file__))

os.chdir(path+"/chrootdir")
os.chroot(path+"/chrootdir")

if len(sys.argv) > 1:
    exec(str(sys.argv[1]))    
