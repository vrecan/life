# Life [![Build Status](https://travis-ci.org/vrecan/death.svg?branch=master)](https://travis-ci.org/vrecan/death)

Simple wrapper for handling creation and management of a single background goroutine.

# Why?
In most of our go programs we had a lot of boilerplate code that if done wrong would generate bugs.

common mistakes this helps us avoid:
* sync.Waitgroup without a pointer
* done channel with a size above 0 //if 0 and close is called but start isn't you block forever
* start is always in a Once.Do so repeated calls to start don't spin up multiple goroutines
