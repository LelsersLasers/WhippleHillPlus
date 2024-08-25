#!/bin/bash

cd /home/millankumar/code/WhippleHillPlus

cd frontend
npm i
npm run build

cd ../backend
go build
./backend