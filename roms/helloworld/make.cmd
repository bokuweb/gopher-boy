@echo off
cls
rgbasm -ohello.o hello.s
rgblink -p00 -ohello.gb hello.o
rgbfix -v -m00 -p00 -tHello hello.gb 
