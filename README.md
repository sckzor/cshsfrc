# CSHS FRC Robot Demo

This is an online IDE toolkit designed to allow multiple different students to
work on programming an FRC robot in Python using a simplified interface

### It works by:

```
Golang webserver: This server hosts web IDE for the students
 |
 v

Python (jail) interpreter: This takes the code that the student writes and
converts it to XML
 |
 v

Java robot code: At the adminstrator's will the XML containing movement
instructions is sent over the network to the robot that acts on them 
```


