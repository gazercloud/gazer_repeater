#!/bin/sh
cd node
npm run build
cd ..
cp -r node/build/* ../bin/www/node

