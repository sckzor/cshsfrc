import sys
import os
import robot

path = os.path.abspath(os.path.dirname(__file__))

os.chdir(path+"/chrootdir")
os.chroot(path+"/chrootdir")

bad_words = ["sys", "os", "open", "exec", "while", "import", "exit", "close", "print" ]
code = ""

if len(sys.argv) <= 1:
    exit()
    
code = str(sys.argv[1])
for line in code.splitlines():
    for word in bad_words:
        if word in line:
            print("A disallowed word was used in your code!\nDo not use the following words:")
            for bad in bad_words:
                print(" - " + bad)
            exit()
try:            
    exec(code)
except Exception as e:
    err = str(e)
    err.replace("status", "code")
    print(e)
