# Constants
METERS = "meters"
FEET = "feet"
DEGREES = "degrees"
RADIANS = "radians"

FEET_TO_METERS = 0.3048
DEG_TO_RAD = 0.0174533

exported = "Start:0\n"

# Functions to drive and move and other stuff
def driveX(distance, units):
    global exported
    if units == FEET:
        distance = distance * FEET_TO_METERS
    exported += f"DriveX:{distance}\n"

def driveY(amount, units):
    global exported
    if units == FEET:
        distance = distance * FEET_TO_METERS
    exported += f"DriveY:{amount}\n"

def run():
    global exported
    exported += "End:0\n"
    print(f"=== Action Output ===\n{exported}")
