# My Guitar Pro Journey

I have been playing guitar for many years. Over time I have accumulated a lot of Guitar Pro files. From Guitar Pro 4 to 8. There are good tools for Guitar Pro, each tool has its strengths and weaknesses. In order to be able to design my own software according to my ideas, I need a library to manage these files.

Here are some projects that can be used with Guitar Pro. Some of these also contain good tips on how to process the files.

- [https://github.com/Perlence/PyGuitarPro]
PyGuitarPro (Python)

- [https://github.com/juliangruber/parse-gp5]
Javascript Parser for GP5

- [https://github.com/slundi/guitarpro/blob/master/lib/README.md] Slundi, Describes how Guitar Pro files are structured and contain valuable information.

- [https://github.com/PhilPotter/gp_parser] C++ replication of Julian Gruber's Javascript code.

- [https://github.com/ImmaculatePine/guitar_pro_parser]
 Ruby version - Read GP-Files. I don't know how good this source is, I've only just found it. Since I started with Ruby and understand this language more quickly - I've documented it immediately so that the URL doesn't get lost :-)

[https://github.com/Antti/rust-gpx-reader] Important information about how a GPX file is structured.

## Status

This library can currently read information from Gutiar Pro 4 - 6 (no GPX) such as
Song title, author, subtitle, creator, copyright etc.
Tracks and music information is not yet implemented.

I have not yet been able to read information from a GPX file. However, since I need this information, I am actively working on it.

## about me

I'm not a Go pro, but I chose this language because I simply love Go (not a love-hate relationship like with ruby ​​:-) ).
With this project I want to grow and delve deeper into Go. I would like to encourage everyone to join me, improve this library and learn together.

I would be happy if someone would join me in building this library.

## TODO List

- Simple Songinformation from
  - [X] GP3
  - [X] GP4
  - [X] GP5
  - [ ] GPX

- Read music data
  - [ ] GP3
  - [ ] GP4
  - [ ] GP5
  - [ ] GPX
