import sys
import os
import robot

os.chdir("/home/sckzor/Code/cshsfrc/python/chrootdir")
os.chroot("/home/sckzor/Code/cshsfrc/python/chrootdir")

if len(sys.argv) > 1:
    exec(str(sys.argv[1]))    
