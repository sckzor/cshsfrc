# Constants
METERS = "meters"
FEET = "feet"
DEGREES = "degrees"
RADIANS = "radians"

FEET_TO_METERS = 0.3048
DEG_TO_RAD = 0.0174533

exported = ""

# Functions to drive and move and other stuff
def drive(distance, units):
    global exported
    if units == FEET:
        distance = distance * FEET_TO_METERS
    exported += f"Drive:{distance}\n"

def turn(amount, units):
    global exported
    if units == DEGREES:
        amount = amount * DEG_TO_RAD
    exported += f"Turn:{amount}\n"

def stop():
    global exported
    exported += f"Stop:0\n"

def run():
    global exported
    print(f"=== Action Output ===\n{exported}")
