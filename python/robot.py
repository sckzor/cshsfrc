# Constants
METERS = "meters"
FEET = "feet"
DEGREES = "degrees"
RADIANS = "radians"

# Functions to drive and move and other stuff
def drive(distance, units):
    print(f"Driving {distance} {units}.")

def turn(amount, units):
    print(f"Turning {amount} {units}.")

def stop():
    print(f"Stopping!")
