#!/bin/sh
cd home
npm run build
cd ..
cp -r home/build/* ../bin/www/home

