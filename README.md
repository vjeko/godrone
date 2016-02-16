# Cross-compile:

Use the deployment script to compile and deploy the firmware.

# Internals

The accelerometers and gyroscopes constitute a low-cost inertial
measurement unit (IMU). The cost of this IMU is less
than 10 USD. A Bosch BMA150 3-axis accelerometer using a
10 bits A/D converter is used. It has a +/- 2g range. The two
axis gyro is an Invensense IDG500. It is an analog sensor. It is
digitalized by the PIC 12 bits A/D converter, and can measure
rotation rates up to 500 degrees/s. On the vertical axis, a more
accurate gyroscope is considered. It is an Epson XV3700. It
has an auto-zero function to minimize heading drift. The IMU
is running at a 200Hz rate.
